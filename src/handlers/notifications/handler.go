package notifications

import (
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	announcement_service "go-do-the-thing/src/services/announcement"
	templ_shared "go-do-the-thing/src/shared/templ"
	"net/http"
	"strconv"
)

type NotificationsHandler struct {
	logger              slog.Logger
	announcementService announcement_service.AnnouncementService
}

var source = "NotificationsHandler"

func SetupNotificationsHandler(announcementService announcement_service.AnnouncementService, router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(source)
	logger.Info("Setting up the Notifications Handler")
	handler := &NotificationsHandler{
		logger:              logger,
		announcementService: announcementService,
	}
	router.Handle("GET /notifications", mw_stack(http.HandlerFunc(handler.GetNotifications)))
	router.Handle("GET /notifications/count", mw_stack(http.HandlerFunc(handler.GetNotificationCount)))
	router.Handle("GET /notifications/banner", mw_stack(http.HandlerFunc(handler.GetAnnouncementBanner)))
	router.Handle("POST /notifications/dismiss/{id}", mw_stack(http.HandlerFunc(handler.DismissNotification)))
}

func (h *NotificationsHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	announcements, err := h.announcementService.GetUserAnnouncements(currentUserId, 5)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve user announcements")
		return
	}

	if err := templ_shared.NotificationsList(announcements).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to render notifications list")
		return
	}
}

func (h *NotificationsHandler) GetNotificationCount(w http.ResponseWriter, r *http.Request) {
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	announcements, err := h.announcementService.GetUserAnnouncements(currentUserId, 5)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve user announcements")
		return
	}

	count := len(announcements)
	if err := templ_shared.NotificationBadge(count).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to render notification badge")
		return
	}
}

func (h *NotificationsHandler) DismissNotification(w http.ResponseWriter, r *http.Request) {
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	announcementId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(w, r, h.logger, err, "Invalid announcement ID provided")
		return
	}

	err = h.announcementService.DismissAnnouncement(currentUserId, announcementId)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to dismiss announcement")
		return
	}

	// Get updated announcements list for the popup
	announcements, err := h.announcementService.GetUserAnnouncements(currentUserId, 10)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve updated announcements")
		return
	}

	// Get updated banner
	latestAnnouncement, err := h.announcementService.GetLatestAnnouncementForUser(currentUserId)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve latest announcement")
		return
	}

	// Calculate count from announcements
	count := len(announcements)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// Determine the primary target based on the request
	targetHeader := r.Header.Get("HX-Target")

	if targetHeader == "#notifications-list" {
		// Dismissing from popup - return notifications list as primary target
		if err := templ_shared.NotificationsList(announcements).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "Failed to render notifications list")
			return
		}
	} else {
		// Dismissing from banner - return empty content for banner container
		// (this will be the primary target)
	}

	// Always send OOB updates for all other components

	// 1. Update banner (OOB)
	w.Write([]byte(`<div id="announcement-banner-container" hx-swap-oob="innerHTML">`))
	if latestAnnouncement != nil {
		if err := templ_shared.AnnouncementBanner(*latestAnnouncement).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "Failed to render announcement banner")
			return
		}
	}
	w.Write([]byte(`</div>`))

	// 2. Update notifications list (OOB) - only if it exists
	w.Write([]byte(`<div id="notifications-list" hx-swap-oob="innerHTML" class="max-h-96 overflow-y-auto">`))
	if err := templ_shared.NotificationsList(announcements).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Failed to render notifications list OOB")
		return
	}
	w.Write([]byte(`</div>`))

	// 3. Update notification badge (OOB)
	if err := templ_shared.NotificationBadge(count).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Failed to render notification badge")
		return
	}
	// Write empty content to ensure HTMX removes the element
	w.Write([]byte(""))
}

func (h *NotificationsHandler) GetAnnouncementBanner(w http.ResponseWriter, r *http.Request) {
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	announcement, err := h.announcementService.GetLatestAnnouncementForUser(currentUserId)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve latest announcement")
		return
	}

	if announcement == nil {
		// No announcement to show
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := templ_shared.AnnouncementBanner(*announcement).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to render announcement banner")
		return
	}
}
