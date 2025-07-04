package admin

import (
	templ_admin "go-do-the-thing/src/handlers/admin/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"net/http"
)

type AdminHandler struct {
	logger slog.Logger
}

var source = "AdminHandler"

func SetupAdminHandler(router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(source)
	logger.Info("Setting up the Admin Handler")
	handler := &AdminHandler{
		logger: logger,
	}
	router.Handle("/admin", mw_stack(http.HandlerFunc(handler.Dashboard)))
	router.Handle("/admin/user-activation", mw_stack(http.HandlerFunc(handler.UserActivation)))
	router.Handle("/admin/activate-user", mw_stack(http.HandlerFunc(handler.ActivateUser)))
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if err := templ_admin.AdminDashboardWithBody(models.ScreenAdmin).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Failed to render admin dashboard")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *AdminHandler) UserRegistrations(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	contentType := r.Header.Get("accept")
	if contentType == "text/html" {
		err = templ_admin.AdminDashboardWithBody(models.ScreenAdmin).Render(r.Context(), w)
		if err != nil {
			h.logger.Error(err, "failed to get projects for user %d", current_user_id)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		err = templ_admin.AdminDashboard(models.ScreenAdmin).Render(r.Context(), w)
		if err != nil {
			h.logger.Error(err, "failed to get projects for user %d", current_user_id)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *AdminHandler) UserActivation(w http.ResponseWriter, r *http.Request) {
	if err := templ_admin.UserActivationTable(nil).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Failed to render user activation table")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AdminHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
}
