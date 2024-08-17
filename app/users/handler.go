package users

import (
	"errors"
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

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	helpers.AcceptHeaderSwitch(w, r, h.loginUI, h.loginApi)
}

func (h *Handler) loginApi(w http.ResponseWriter, _ *http.Request) {
	shared.HttpErrorUI(h.templates, "Handler not implemented", errors.New("handler not implemented"), w)
}

func (h *Handler) loginUI(w http.ResponseWriter, _ *http.Request) {
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
