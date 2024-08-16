package users

import (
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

func (h *Handler) Login(w http.ResponseWriter, _ *http.Request) {
	h.templates.RenderOk(w, "login", nil)
}
