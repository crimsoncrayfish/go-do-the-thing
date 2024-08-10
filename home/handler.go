package home

import (
	"fmt"
	"go-do-the-thing/helpers"
	"go-do-the-thing/navigation"
	"net/http"
)

type Handler struct {
	model     Screens
	templates helpers.Templates
}

func New(templates helpers.Templates) *Handler {
	return &Handler{
		model: Screens{
			navigation.NavBarObject{
				IsHome: true,
			},
		},
		templates: templates,
	}
}

func (h *Handler) Index(w http.ResponseWriter, _ *http.Request) {
	if err := h.templates.RenderOk(w, "index", h.model); err != nil {
		fmt.Println("Failed to execute tmpl for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h *Handler) Home(w http.ResponseWriter, _ *http.Request) {
	if err := h.templates.RenderOk(w, "home", h.model); err != nil {
		fmt.Println("Failed to execute tmpl for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
