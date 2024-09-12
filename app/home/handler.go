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
				IsHome: true,
			},
		},
		templates: templates,
		logger:    logger,
	}
}

func (h *Handler) Index(w http.ResponseWriter, _ *http.Request) {
	if err := h.templates.RenderOk(w, "index", h.model); err != nil {
		h.logger.Error(err, "Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Home(w http.ResponseWriter, _ *http.Request) {
	if err := h.templates.RenderOk(w, "home", h.model); err != nil {
		h.logger.Error(err, "Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
