package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"go-do-the-thing/app/handlers"
	"go-do-the-thing/app/models"
	"go-do-the-thing/app/repos"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/security"
	"go-do-the-thing/helpers/slog"
	"go-do-the-thing/middleware"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	templates helpers.Templates
	security  security.JwtHandler
	model     Screens
	repo      repos.UsersRepo
	logger    *slog.Logger
}

// TODO: What even is this
type Screens struct {
	NavBar models.NavBarObject
}

func SetupUserHandler(
	userRepo repos.UsersRepo,
	router *http.ServeMux,
	templates helpers.Templates,
	mw middleware.Middleware,
	mw_no_auth middleware.Middleware,
	security security.JwtHandler,
) error {
	logger := slog.NewLogger("Users")
	logger.Info("Setting up users")
	handler := &Handler{
		model: Screens{
			NavBar: models.NavBarObject{
				ActiveScreens: models.ActiveScreens{IsHome: true},
			},
		},
		templates: templates,
		repo:      userRepo,
		security:  security,
		logger:    logger,
	}

	router.Handle("GET /login", mw_no_auth(http.HandlerFunc(handler.GetLoginUI)))
	router.Handle("GET /register", mw_no_auth(http.HandlerFunc(handler.GetRegisterUI)))
	router.Handle("POST /login", mw_no_auth(http.HandlerFunc(handler.LoginUI)))
	router.Handle("POST /signup", mw_no_auth(http.HandlerFunc(handler.RegisterUI)))
	router.Handle("POST /logout", mw(http.HandlerFunc(handler.LogOut)))

	router.HandleFunc("GET /users", handler.GetAll)
	return nil
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
	email, errorForm := models.GetRequiredPropertyFromRequest(r, "email", errorForm, true)
	password, errorForm := models.GetRequiredPropertyFromRequest(r, "password", errorForm, false)
	if len(errorForm.Errors) > 0 {
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "login-form-content", errorForm); err != nil {
			handlers.HttpErrorUI(h.templates, "Failed to render template for formData", err, w)
		}
		http.SetCookie(w, nil)
		return
	}

	user, err := h.repo.GetUserByEmail(email)

	if err != nil {
		// NOTE: Not a valid user but Shhhh! dont tell them
		// TODO: Keep track of accounts that have invalid logins and lock them after a set amount of login attempts
		// TODO: keep track of IPs that have invalid logins and ban them after a set count
		// TODO: check err type
		http.SetCookie(w, nil)
		if errors.Is(err, sql.ErrNoRows) {
			h.invalidLogin(w, errorForm, "User not in database")
			return
		}
		h.serverError(err, w, errorForm, "Failed to read user from DB with email %s", email)
		return
	}

	passwordHash, err := h.repo.GetUserPassword(user.Id)
	if err != nil {
		http.SetCookie(w, nil)
		h.serverError(err, w, errorForm, "Failed to read password for user %d", user.Id)
		return
	}

	if !security.CheckPassword(password, passwordHash) {
		// NOTE: Not a valid password but Shhhh! dont tell them
		// TODO: Keep track of accounts that have invalid logins and lock them after a set amount of login attempts
		// TODO: keep track of IPs that have invalid logins and ban them after a set count
		h.invalidLogin(w, errorForm, "Invalid password")
		http.SetCookie(w, nil)
		return
	}
	user.SessionId = uuid.New().String()

	user.SessionStartTime = database.SqLiteNow()
	if err := h.repo.UpdateSession(user); err != nil {
		h.serverError(err, w, errorForm, "Failed to set session id for user %d", user.Id)
		http.SetCookie(w, nil)
		return
	}
	tokenString, err := h.security.NewToken(user.Email, user.SessionId, user.SessionStartTime.Time.Add(time.Duration(time.Hour*4)))
	if err != nil {
		// NOTE: Failed to create a token. Hmmm. Should probably throw internalServerErr
		h.serverError(err, w, errorForm, "failed to generate token")
		http.SetCookie(w, nil)
		return
	}
	cookie := http.Cookie{Name: "token", Value: tokenString, SameSite: http.SameSiteDefaultMode}
	http.SetCookie(w, &cookie)
	// TODO: what to do?
	handlers.Redirect("/", w)
}

func (h Handler) serverError(err error, w http.ResponseWriter, formData models.FormData, message string, params ...any) {
	h.logger.Error(err, message, params...)
	formData.Errors["Failed Login"] = "Something went wrong on the server. Please try again."
	if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "login-form-content", formData); err != nil {
		handlers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
	}
}

func (h Handler) invalidLogin(w http.ResponseWriter, formData models.FormData, message string, params ...any) {
	h.logger.Info(message, params...)
	formData.Errors["Failed Login"] = "Login credentials are invalid."
	if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "login-form-content", formData); err != nil {
		handlers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
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
	userId, ok := r.Context().Value(helpers.AuthUserId).(string)
	if !ok {
		handlers.HttpErrorUI(h.templates, "Failed to get a userId from context", errors.New("cannot get userid from context"), w)
		return
	}
	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to parse int from string", err, w)
		return
	}

	if err = h.repo.Logout(id); err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}

func (h Handler) GetRegisterUI(w http.ResponseWriter, _ *http.Request) {
	err := h.templates.RenderOk(w, "signup", nil)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}

func (h Handler) RegisterUI(w http.ResponseWriter, r *http.Request) {
	errorForm := models.NewFormData()
	name, errorForm := models.GetRequiredPropertyFromRequest(r, "name", errorForm, true)
	email, errorForm := models.GetRequiredPropertyFromRequest(r, "email", errorForm, true)
	password, errorForm := models.GetRequiredPropertyFromRequest(r, "password", errorForm, false)
	password2, errorForm := models.GetRequiredPropertyFromRequest(r, "password2", errorForm, false)

	if len(errorForm.Errors) > 0 {
		h.logger.Info("Failed to register due to invalid form details")
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "signup-form-content", errorForm); err != nil {
			handlers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		}
		return
	}
	if password != password2 {
		h.logger.Info("Failed to register due to passwords not matching")
		errorForm.Errors["Password"] = "The password fields dont match"
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "signup-form-content", errorForm); err != nil {
			handlers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		}
		return
	}
	user, _ := h.repo.GetUserByEmail(email) // TODO: what to do with this err message
	if (models.User{}) != user {
		h.logger.Info("Registration failure. Email %s already in use", email)
		errorForm.Errors["Email"] = "Email already in use"
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "signup-form-content", errorForm); err != nil {
			handlers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		}

		return
	}

	passwordHash, err := security.SetPassword(password)
	if err != nil {
		h.logger.Error(err, "Failed to hash password")
		errorForm.Errors["Create"] = "Failed to create users due to a server error"
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "signup-form-content", errorForm); err != nil {
			handlers.HttpErrorUI(h.templates, "Invalid login details", errors.New("invalid login credentials"), w)
		}

		return

	}

	user = models.User{
		Email:        email,
		FullName:     name,
		PasswordHash: passwordHash,
		IsDeleted:    false,
		IsAdmin:      false,
	}
	_, err = h.repo.Create(user)
	h.logger.Info("Successfully created user")
	handlers.Redirect("/login", w)
}
