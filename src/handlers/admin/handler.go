package admin

import (
	"fmt"
	templ_admin "go-do-the-thing/src/handlers/admin/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	admin_service "go-do-the-thing/src/services/admin"
	templ_shared "go-do-the-thing/src/shared/templ"
	"net/http"
	"strconv"
)

type AdminHandler struct {
	service admin_service.AdminService
	logger  *logger.Logger
}

var source = "AdminHandler"

func SetupAdminHandler(admin_service admin_service.AdminService, router *http.ServeMux, mw_stack middleware.Middleware) {
	log := logger.NewLogger(source)
	log.Info("Setting up the Admin Handler")
	handler := &AdminHandler{
		service: admin_service,
		logger:  log,
	}
	router.Handle("/admin", mw_stack(http.HandlerFunc(handler.Dashboard)))
	router.Handle("GET /admin/users", mw_stack(http.HandlerFunc(handler.ListUsers)))
	router.Handle("POST /admin/activate-user/{id}", mw_stack(http.HandlerFunc(handler.ActivateUser)))
	router.Handle("POST /admin/deactivate-user/{id}", mw_stack(http.HandlerFunc(handler.DeactivateUser)))
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}
	users, err := h.service.ListUsers(current_user_id)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve inactive users")
		return
	}
	if err := templ_admin.AdminDashboardWithBody(models.ScreenAdmin, users).Render(r.Context(), w); err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display admin dashboard")
		return
	}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}
	users, err := h.service.ListUsers(current_user_id)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve users")
		return
	}
	if err := templ_admin.UserTable(users).Render(r.Context(), w); err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display users")
		return
	}
}

func (h *AdminHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Invalid user ID provided")
		return
	}

	user_email, err := h.service.ActivateUser(current_user_id, id)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to activate user")
		return
	}

	updated_user, err := h.service.GetUserById(current_user_id, id)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to activate user")
		return
	}

	err = templ_shared.RenderTempls(
		templ_admin.UserRow(updated_user),
		templ_shared.ToastMessage(fmt.Sprintf("Activated user %s", user_email), "success"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to update ui")
		return
	}
}

func (h *AdminHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Invalid user ID provided")
		return
	}

	user_email, err := h.service.DeactivateUser(current_user_id, id)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to deactivate user")
		return
	}

	updated_user, err := h.service.GetUserById(current_user_id, id)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to get user")
		return
	}

	err = templ_shared.RenderTempls(
		templ_admin.UserRow(updated_user),
		templ_shared.ToastMessage(fmt.Sprintf("Deactivated user %s", user_email), "success"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to update ui")
		return
	}
}
