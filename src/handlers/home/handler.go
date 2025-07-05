package home

import (
	"errors"
	home_templ "go-do-the-thing/src/handlers/home/templ"
	"go-do-the-thing/src/helpers"
	app_errors "go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"net/http"
)

type HomeHandler struct {
	logger slog.Logger
}

var source = "HomeHandler"

func SetupHomeHandler(router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(source)
	logger.Info("Setting up the Home Handler")
	handler := &HomeHandler{
		logger: logger,
	}
	router.Handle("/", mw_stack(http.HandlerFunc(handler.Index)))
	router.Handle("GET /error", mw_stack(http.HandlerFunc(handler.error)))
	router.Handle("GET /home", mw_stack(http.HandlerFunc(handler.Home)))
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		app_errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	if err := home_templ.Index(models.ScreenHome).Render(r.Context(), w); err != nil {
		app_errors.InternalServerError(w, r, h.logger, err, "Failed to display home page")
		return
	}
}

func (h *HomeHandler) error(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		app_errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	app_errors.InternalServerError(w, r, h.logger, errors.New("testing"), "Testing Error Toasts")
}

func (h *HomeHandler) Home(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		app_errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	if err := home_templ.Index(models.ScreenHome).Render(r.Context(), w); err != nil {
		app_errors.InternalServerError(w, r, h.logger, err, "Failed to display home page")
		return
	}
}
