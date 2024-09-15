package home

import (
	"go-do-the-thing/app/models"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/slog"
	"go-do-the-thing/middleware"
	"net/http"
)

type Handler struct {
	model     Screens
	templates helpers.Templates
	logger    *slog.Logger
}

func SetupHomeHandler(router *http.ServeMux, templates helpers.Templates, mw_stack middleware.Middleware) {
	logger := slog.NewLogger("Home")
	logger.Info("Setting up the Home screen")
	handler := &Handler{
		model: Screens{
			models.NavBarObject{
				ActiveScreens: models.ActiveScreens{IsHome: true},
			},
		},
		templates: templates,
		logger:    logger,
	}
	router.Handle("/", mw_stack(http.HandlerFunc(handler.Index)))
	router.Handle("GET /home", mw_stack(http.HandlerFunc(handler.Home)))
}

type Screens struct {
	NavBar models.NavBarObject
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	data := h.model
	email, name, err := helpers.GetUserFromContext(r)
	if err != nil {
		h.logger.Error(err, "could not get user details from http context")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data.NavBar = data.NavBar.SetUser(name, email)
	if err := h.templates.RenderOk(w, "index", data); err != nil {
		h.logger.Error(err, "Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	data := h.model
	email, name, err := helpers.GetUserFromContext(r)
	if err != nil {
		h.logger.Error(err, "could not get user details from http context")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data.NavBar = data.NavBar.SetUser(name, email)
	if err := h.templates.RenderOk(w, "home", data); err != nil {
		h.logger.Error(err, "Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
