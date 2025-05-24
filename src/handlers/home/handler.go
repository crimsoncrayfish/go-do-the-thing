package home

import (
	home_templ "go-do-the-thing/src/handlers/home/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	templ_shared "go-do-the-thing/src/shared/templ"
	"net/http"
)

type HomeHandler struct {
	logger slog.Logger
}

var activeScreens Screens

var source = "HomeHandler"

func SetupHomeHandler(router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(source)
	logger.Info("Setting up the Home screen")
	activeScreens = Screens{
		models.NavBarObject{
			ActiveScreens: models.ActiveScreens{IsHome: true},
		},
	}
	handler := &HomeHandler{
		logger: logger,
	}
	router.Handle("/", mw_stack(http.HandlerFunc(handler.Index)))
	router.Handle("GET /error", mw_stack(http.HandlerFunc(handler.error)))
	router.Handle("GET /home", mw_stack(http.HandlerFunc(handler.Home)))
}

type Screens struct {
	NavBar models.NavBarObject
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	_, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	if err := home_templ.Index(activeScreens.NavBar).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HomeHandler) error(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusInternalServerError)
	h.logger.Debug("testing error")

	templ_shared.ToastMessage("This is an error", "error").Render(r.Context(), w)
}

func (h *HomeHandler) Home(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	_, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	if err := home_templ.Index(activeScreens.NavBar).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
