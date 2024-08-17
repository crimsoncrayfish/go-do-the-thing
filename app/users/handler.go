package users

import (
	"go-do-the-thing/app/shared"
	"go-do-the-thing/helpers"
	"net/http"
)

type Handler struct {
	templates helpers.Templates
	repo      Repo
}

func NewHandler(templates helpers.Templates, repo Repo) Handler {
	return Handler{templates, repo}
}

func (h *Handler) LoginUI(w http.ResponseWriter, _ *http.Request) {
	err := h.templates.RenderOk(w, "login", nil)
	if err != nil {
		shared.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}

func (h *Handler) RegisterUI(w http.ResponseWriter, _ *http.Request) {
	err := h.templates.RenderOk(w, "register", nil)
	if err != nil {
		shared.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}
