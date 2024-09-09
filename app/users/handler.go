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
	security  security.JwtHandler
	model     home.Screens
	repo      Repo
}

func New(templates helpers.Templates, repo Repo, security security.JwtHandler) *Handler {
	return &Handler{
		model: home.Screens{
			models.NavBarObject{
				IsHome: true,
			},
		},
		templates: templates,
		repo:      repo,
		security:  security,
	}
}

func (h Handler) LoginUI(w http.ResponseWriter, r *http.Request) {
	errorForm := models.NewFormData()
	name, errorForm := helpers.GetRequiredPropertyFromRequest(r, "name", errorForm)
	password, errorForm := helpers.GetRequiredPropertyFromRequest(r, "password", errorForm)

	user, err := h.repo.GetUserByName(name)

	if err != nil {
		// NOTE: Not a valid user but Shhhh! dont tell them
		// TODO: Keep track of accounts that have invalid logins and lock them after a set amount of login attempts
		// TODO: keep track of IPs that have invalid logins and ban them after a set count
		fmt.Println("Not a valid user")
		helpers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		return
	}
	if !security.CheckPassword(password, user.PasswordHash) {
		// NOTE: Not a valid password but Shhhh! dont tell them
		// TODO: Keep track of accounts that have invalid logins and lock them after a set amount of login attempts
		// TODO: keep track of IPs that have invalid logins and ban them after a set count

		fmt.Println("Invalid password")
		helpers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		return
	}
	tokenString, err := h.security.NewToken(user.Name)
	if err != nil {
		// NOTE: Failed to create a token. Hmmm. Should probably throw internalServerErr
		fmt.Println("failed to generate token")
		helpers.HttpErrorUI(h.templates, "Failed to generate new token", errors.New("failed to generate token"), w)
		return
	}
	fmt.Fprint(w, tokenString)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h Handler) GetLoginUI(w http.ResponseWriter, _ *http.Request) {
	if err := h.templates.RenderOk(w, "login", nil); err != nil {
		fmt.Println("Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) LogOut(w http.ResponseWriter, r *http.Request) {

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

func (h Handler) RegisterUI(w http.ResponseWriter, _ *http.Request) {
	err := h.templates.RenderOk(w, "register", nil)
	if err != nil {
		helpers.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}
