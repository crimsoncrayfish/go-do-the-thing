package users

import (
	"errors"
	"fmt"
	"go-do-the-thing/app/home"
	"go-do-the-thing/app/shared/models"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/security"
	"net/http"
)

type Handler struct {
	templates helpers.Templates
	model     home.Screens
	repo      Repo
}

func New(templates helpers.Templates, repo Repo) *Handler {
	return &Handler{
		model: home.Screens{
			models.NavBarObject{
				IsHome: true,
			},
		},
		templates: templates,
		repo:      repo,
	}
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	helpers.AcceptHeaderSwitch(w, r, h.loginUI, h.loginApi)
}

func (h Handler) loginApi(w http.ResponseWriter, _ *http.Request) {
	helpers.HttpError("Handler not implemented", errors.New("handler not implemented"), w)
}

func (h Handler) loginUI(w http.ResponseWriter, _ *http.Request) {
	if err := h.templates.RenderOk(w, "index", nil); err != nil {
		fmt.Println("Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) LogOut(w http.ResponseWriter, r *http.Request) {
	helpers.AcceptHeaderSwitch(w, r, h.logoutUI, h.logoutAPI)
}

func (h Handler) logoutUI(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(security.AuthUserId).(string)
	if !ok {
		helpers.HttpErrorUI(h.templates, "Failed to get a userId from context", errors.New("cannot get userid from context"), w)
		return
	}
	err := h.repo.Logout(userId)
	if err != nil {
		helpers.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}

func (h Handler) logoutAPI(w http.ResponseWriter, _ *http.Request) {
	helpers.HttpError("Handler not implemented", errors.New("handler not implemented"), w)
}

func (h Handler) Signup(w http.ResponseWriter, r *http.Request) {
	helpers.AcceptHeaderSwitch(w, r, h.registerUI, h.registerApi)
}

func (h Handler) registerApi(w http.ResponseWriter, _ *http.Request) {
	helpers.HttpErrorUI(
		h.templates,
		"Handler not implemented",
		errors.New("handler not implemented"),
		w,
	)
}

func (h Handler) registerUI(w http.ResponseWriter, _ *http.Request) {
	err := h.templates.RenderOk(w, "register", nil)
	if err != nil {
		helpers.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}
