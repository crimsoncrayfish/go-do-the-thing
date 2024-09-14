package home

import (
	"go-do-the-thing/app/shared/models"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/slog"
	"net/http"
)

type Handler struct {
	model     Screens
	templates helpers.Templates
	logger    *slog.Logger
}

func New(templates helpers.Templates, logger *slog.Logger) *Handler {
	return &Handler{
		model: Screens{
			models.NavBarObject{
				ActiveScreens: models.ActiveScreens{IsHome: true},
			},
		},
		templates: templates,
		logger:    logger,
	}
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
