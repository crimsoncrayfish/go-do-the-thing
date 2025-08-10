package admin

import (
	"fmt"
	templ_admin "go-do-the-thing/src/handlers/admin/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/constants"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	form_models "go-do-the-thing/src/models/forms"
	admin_service "go-do-the-thing/src/services/admin"
	announcement_service "go-do-the-thing/src/services/announcement"
	templ_shared "go-do-the-thing/src/shared/templ"
	"net/http"
	"strconv"
	"time"
)

type AdminHandler struct {
	logger              slog.Logger
	service             admin_service.AdminService
	announcementService announcement_service.AnnouncementService
}

var source = "AdminHandler"

func SetupAdminHandler(admin_service admin_service.AdminService, announcementService announcement_service.AnnouncementService, router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(source)
	logger.Info("Setting up the Admin Handler")
	handler := &AdminHandler{
		logger:              logger,
		service:             admin_service,
		announcementService: announcementService,
	}
	router.Handle("/admin", mw_stack(http.HandlerFunc(handler.Dashboard)))
	router.Handle("GET /admin/users", mw_stack(http.HandlerFunc(handler.ListUsers)))
	router.Handle("GET /admin/search-users", mw_stack(http.HandlerFunc(handler.SearchUsers)))
	router.Handle("GET /admin/analytics", mw_stack(http.HandlerFunc(handler.Analytics)))
	router.Handle("GET /admin/announcements", mw_stack(http.HandlerFunc(handler.Announcements)))
	router.Handle("POST /admin/announcements", mw_stack(http.HandlerFunc(handler.CreateAnnouncement)))
	router.Handle("DELETE /admin/announcements/{id}", mw_stack(http.HandlerFunc(handler.DeleteAnnouncement)))
	router.Handle("POST /admin/activate-user/{id}", mw_stack(http.HandlerFunc(handler.ActivateUser)))
	router.Handle("POST /admin/deactivate-user/{id}", mw_stack(http.HandlerFunc(handler.DeactivateUser)))
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}
	users, err := h.service.ListUsers(current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve inactive users")
		return
	}
	if err := templ_admin.AdminDashboardWithBody(models.ScreenAdmin, users).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display admin dashboard")
		return
	}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}
	users, err := h.service.ListUsers(current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve users")
		return
	}

	err = templ_shared.RenderTempls(
		templ_admin.UserTable(users),
		templ_admin.AdminNavTabs("users"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display users")
		return
	}
}

func (h *AdminHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		searchTerm = r.URL.Query().Get("search")
	}

	users, err := h.service.SearchUsers(current_user_id, searchTerm)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to search users")
		return
	}

	if err := templ_admin.UserRows(users).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display search results")
		return
	}
}

func (h *AdminHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(w, r, h.logger, err, "Invalid user ID provided")
		return
	}

	user_email, err := h.service.ActivateUser(current_user_id, id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to activate user")
		return
	}

	updated_user, err := h.service.GetUserById(current_user_id, id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to activate user")
		return
	}

	err = templ_shared.RenderTempls(
		templ_admin.UserRow(updated_user),
		templ_shared.ToastMessage(fmt.Sprintf("Activated user %s", user_email), "success"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to update ui")
		return
	}
}

func (h *AdminHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(w, r, h.logger, err, "Invalid user ID provided")
		return
	}

	user_email, err := h.service.DeactivateUser(current_user_id, id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to deactivate user")
		return
	}

	updated_user, err := h.service.GetUserById(current_user_id, id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to get user")
		return
	}

	err = templ_shared.RenderTempls(
		templ_admin.UserRow(updated_user),
		templ_shared.ToastMessage(fmt.Sprintf("Deactivated user %s", user_email), "success"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to update ui")
		return
	}
}

func (h *AdminHandler) Analytics(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	dashboard, err := h.service.GetAnalyticsDashboard(current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve analytics dashboard")
		return
	}

	err = templ_shared.RenderTempls(
		templ_admin.AnalyticsDashboard(*dashboard),
		templ_admin.AdminNavTabs("analytics"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display analytics dashboard")
		return
	}
}

func (h *AdminHandler) Announcements(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	announcements, err := h.announcementService.GetAllAnnouncements(current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve announcements")
		return
	}

	form := form_models.NewAnnouncementForm()

	err = templ_shared.RenderTempls(
		templ_admin.AnnouncementsDashboard(announcements, form),
		templ_admin.AdminNavTabs("announcements"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display announcements dashboard")
		return
	}
}

func (h *AdminHandler) CreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	form := form_models.NewAnnouncementForm()
	form.Title = r.FormValue("title")

	if expiresDateStr := r.FormValue("expires_date"); expiresDateStr != "" {
		if expiresDate, err := time.Parse(constants.DateFormat, expiresDateStr); err == nil {
			form.ExpiresDate = &expiresDate
		}
	}

	if form.Title == "" {
		form.SetError("title", "Announcement text is required")
	}

	if form.HasErrors() {
		announcements, err := h.announcementService.GetAllAnnouncements(current_user_id)
		if err != nil {
			errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve announcements")
			return
		}

		if err := templ_admin.AnnouncementsDashboard(announcements, form).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, h.logger, err, "Failed to display form with errors")
			return
		}
		return
	}

	_, err = h.announcementService.CreateAnnouncement(current_user_id, form.Title, form.ExpiresDate)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to create announcement")
		return
	}

	announcements, err := h.announcementService.GetAllAnnouncements(current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve announcements")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// Return the updated announcements list (primary target)
	if err := templ_admin.AnnouncementsList(announcements).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to render announcements list")
		return
	}

	// Clear the form fields via JavaScript (OOB)
	w.Write([]byte(`<script hx-swap-oob="afterbegin:body">
		if (typeof clearAnnouncementForm === 'function') {
			clearAnnouncementForm();
		}
	</script>`))

	// Return success toast (OOB)
	if err := templ_shared.ToastMessage("Announcement created successfully", "success").Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to render toast")
		return
	}
}

func (h *AdminHandler) DeleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	announcementId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(w, r, h.logger, err, "Invalid announcement ID provided")
		return
	}

	err = h.announcementService.DeleteAnnouncement(current_user_id, announcementId)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to delete announcement")
		return
	}

	// Get updated announcements list to show "no announcements" message if needed
	announcements, err := h.announcementService.GetAllAnnouncements(current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to fetch announcements after deletion")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// Return the updated announcements list
	if err := templ_admin.AnnouncementsList(announcements).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to render announcements list")
		return
	}
}
