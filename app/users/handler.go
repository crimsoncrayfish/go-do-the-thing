package users

import (
	"encoding/json"
	"errors"
	"go-do-the-thing/app/home"
	"go-do-the-thing/app/shared/models"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/security"
	"go-do-the-thing/helpers/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	templates helpers.Templates
	security  security.JwtHandler
	model     home.Screens
	repo      Repo
	logger    *slog.Logger
}

func New(templates helpers.Templates, repo Repo, security security.JwtHandler, logger *slog.Logger) *Handler {
	return &Handler{
		model: home.Screens{
			ActiveScreens: models.NavBarObject{
				IsHome: true,
			},
		},
		templates: templates,
		repo:      repo,
		security:  security,
		logger:    logger,
	}
}

func (h Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.GetUsers()
	if err != nil {
		h.logger.Error(err, "failed to get users")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonBytes, err := json.Marshal(users)
	if err != nil {
		h.logger.Error(err, "failed to marshal users")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		h.logger.Error(err, "failed to write response")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h Handler) LoginUI(w http.ResponseWriter, r *http.Request) {
	errorForm := models.NewFormData()
	email, errorForm := helpers.GetRequiredPropertyFromRequest(r, "email", errorForm, true)
	password, errorForm := helpers.GetRequiredPropertyFromRequest(r, "password", errorForm, false)
	if len(errorForm.Errors) > 0 {
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "login-form-content", errorForm); err != nil {
			helpers.HttpErrorUI(h.templates, "Failed to render template for formData", err, w)
		}
		return
	}

	user, err := h.repo.GetUserByEmail(email)

	if err != nil {
		// NOTE: Not a valid user but Shhhh! dont tell them
		// TODO: Keep track of accounts that have invalid logins and lock them after a set amount of login attempts
		// TODO: keep track of IPs that have invalid logins and ban them after a set count
		// TODO: check err type
		h.serverError(err, w, errorForm, "Failed to read user from DB with email %s", email)
		return
	}

	passwordHash, err := h.repo.GetUserPassword(user.Id)
	if err != nil {
		h.serverError(err, w, errorForm, "Failed to read password for user %d", user.Id)
		return
	}

	if !security.CheckPassword(password, passwordHash) {
		// NOTE: Not a valid password but Shhhh! dont tell them
		// TODO: Keep track of accounts that have invalid logins and lock them after a set amount of login attempts
		// TODO: keep track of IPs that have invalid logins and ban them after a set count
		h.invalidLogin(w, errorForm, "Invalid password")
		return
	}
	user.SessionId = uuid.New().String()

	user.SessionStartTime = database.SqLiteNow()
	if err := h.repo.UpdateSession(user); err != nil {
		h.serverError(err, w, errorForm, "Failed to set session id for user %d", user.Id)
		return
	}
	tokenString, err := h.security.NewToken(user.Email, user.SessionId, user.SessionStartTime.Time.Add(time.Duration(time.Hour*4)))
	if err != nil {
		// NOTE: Failed to create a token. Hmmm. Should probably throw internalServerErr
		h.serverError(err, w, errorForm, "failed to generate token")
		return
	}
	// fmt.Fprint(w, tokenString)
	// TODO: add session id to jwt
	cookie := http.Cookie{Name: "token", Value: tokenString}
	http.SetCookie(w, &cookie)
	// TODO: what to do?
	helpers.Redirect("/", w)
}

func (h Handler) serverError(err error, w http.ResponseWriter, formData models.FormData, message string, params ...any) {
	h.logger.Error(err, message, params...)
	formData.Errors["Failed Login"] = "Something went wrong on the server. Please try again."
	if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "login-form-content", formData); err != nil {
		helpers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
	}
}

func (h Handler) invalidLogin(w http.ResponseWriter, formData models.FormData, message string, params ...any) {
	h.logger.Info(message, params...)
	formData.Errors["Failed Login"] = "Login credentials are invalid."
	if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "login-form-content", formData); err != nil {
		helpers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
	}
}

func (h Handler) GetLoginUI(w http.ResponseWriter, _ *http.Request) {
	if err := h.templates.RenderOk(w, "login", nil); err != nil {
		h.logger.Info("Failed to execute template for the home page")
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
	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		helpers.HttpErrorUI(h.templates, "Failed to parse int from string", err, w)
		return
	}

	if err = h.repo.Logout(id); err != nil {
		helpers.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}

func (h Handler) GetRegisterUI(w http.ResponseWriter, _ *http.Request) {
	err := h.templates.RenderOk(w, "signup", nil)
	if err != nil {
		helpers.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}

func (h Handler) RegisterUI(w http.ResponseWriter, r *http.Request) {
	errorForm := models.NewFormData()
	name, errorForm := helpers.GetRequiredPropertyFromRequest(r, "name", errorForm, true)
	email, errorForm := helpers.GetRequiredPropertyFromRequest(r, "email", errorForm, true)
	password, errorForm := helpers.GetRequiredPropertyFromRequest(r, "password", errorForm, false)
	password2, errorForm := helpers.GetRequiredPropertyFromRequest(r, "password2", errorForm, false)

	if len(errorForm.Errors) > 0 {
		h.logger.Info("Failed to register due to invalid form details")
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "signup-form-content", errorForm); err != nil {
			helpers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		}
		return
	}
	if password != password2 {
		h.logger.Info("Failed to register due to passwords not matching")
		errorForm.Errors["Password"] = "The password fields dont match"
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "signup-form-content", errorForm); err != nil {
			helpers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		}
		return
	}
	user, _ := h.repo.GetUserByEmail(email) // TODO: what to do with this err message
	if (User{}) != user {
		h.logger.Info("Registration failure. Email %s already in use", email)
		errorForm.Errors["Email"] = "Email already in use"
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "signup-form-content", errorForm); err != nil {
			helpers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		}

		return
	}

	passwordHash, err := security.SetPassword(password)
	if err != nil {
		h.logger.Error(err, "Failed to hash password")
		errorForm.Errors["Create"] = "Failed to create users due to a server error"
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "signup-form-content", errorForm); err != nil {
			helpers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		}

		return

	}

	user = User{
		Email:        email,
		FullName:     name,
		PasswordHash: passwordHash,
		IsDeleted:    false,
		IsAdmin:      false,
	}
	_, err = h.repo.Create(user)
	h.logger.Info("Successfully created user")
	helpers.Redirect("/login", w)
}
