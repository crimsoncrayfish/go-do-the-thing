package admin

import (
	templ_admin "go-do-the-thing/src/handlers/admin/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	admin_service "go-do-the-thing/src/services/admin"
	"net/http"
	"strconv"
)

type AdminHandler struct {
	logger  slog.Logger
	service admin_service.AdminService
}

var source = "AdminHandler"

func SetupAdminHandler(admin_service admin_service.AdminService, router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(source)
	logger.Info("Setting up the Admin Handler")
	handler := &AdminHandler{
		logger:  logger,
		service: admin_service,
	}
	router.Handle("/admin", mw_stack(http.HandlerFunc(handler.Dashboard)))
	router.Handle("POST /admin/activate-user/{id}", mw_stack(http.HandlerFunc(handler.ActivateUser)))
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}
	users, err := h.service.ListInactiveUsers(current_user_id)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to retrieve inactive users")
		return
	}
	if err := templ_admin.AdminDashboardWithBody(models.ScreenAdmin, users).Render(r.Context(), w); err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display admin dashboard")
		return
	}
}

func (h *AdminHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.FrontendErrorBadRequest(w, r, h.logger, err, "Invalid user ID provided")
		return
	}

	err = h.service.ActivateUser(current_user_id, id)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to activate user")
		return
	}
}
