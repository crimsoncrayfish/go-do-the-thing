package users

import (
	"go-do-the-thing/src/handlers"
	templ_users "go-do-the-thing/src/handlers/users/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	fe_errors "go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	form_models "go-do-the-thing/src/models/forms"
	user_service "go-do-the-thing/src/services/user"
	"net/http"
	"time"
)

type Handler struct {
	security    security.JwtHandler
	userService *user_service.UserService
	logger      slog.Logger
}

var source = "UsersHandler"

func SetupUserHandler(
	router *http.ServeMux,
	mw middleware.Middleware,
	mw_no_auth middleware.Middleware,
	security security.JwtHandler,
	userService *user_service.UserService,
) {
	logger := slog.NewLogger(source)
	logger.Info("Setting up users")

	handler := &Handler{
		security:    security,
		userService: userService,
		logger:      logger,
	}

	router.Handle("GET /login", mw_no_auth(http.HandlerFunc(handler.GetLoginUI)))
	router.Handle("GET /register", mw_no_auth(http.HandlerFunc(handler.GetRegisterUI)))
	router.Handle("POST /login", mw_no_auth(http.HandlerFunc(handler.LoginUI)))
	router.Handle("POST /register", mw_no_auth(http.HandlerFunc(handler.RegisterUI)))
	router.Handle("POST /logout", mw(http.HandlerFunc(handler.LogOut)))
}

var emptyAuthCookie = http.Cookie{Name: "token", Value: "", SameSite: http.SameSiteDefaultMode}

func (h Handler) LoginUI(w http.ResponseWriter, r *http.Request) {
	form := form_models.NewLoginForm()
	email, err := models.GetRequiredPropertyFromRequest(r, "email", "Email")
	form.Email = email
	if err != nil {
		form.Errors["email"] = err.Error()
	}
	password, err := models.GetRequiredPropertyFromRequest(r, "password", "Password")
	if err != nil {
		form.Errors["password"] = err.Error()
	}
	if len(form.Errors) > 0 {
		h.logger.Info("LoginUI: invalid form input - email: %s, errors: %v", email, form.Errors)
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_users.LoginFormContent(form).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, source, "Failed to render template for formData")
		}
		http.SetCookie(w, &emptyAuthCookie)
		return
	}
	h.logger.Info("LoginUI: login attempt - email: %s", email)
	user, sessionId, err := h.userService.AuthenticateUser(email, password)
	if err != nil {
		h.logger.Error(err, "LoginUI: login failed - email: %s", email)
		form.Errors["login"] = err.Error()
		w.WriteHeader(http.StatusUnauthorized)
		if err := templ_users.LoginFormContent(form).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		http.SetCookie(w, &emptyAuthCookie)
		return
	}
	h.logger.Info("LoginUI: login succeeded - email: %s, user_id: %d", email, user.Id)
	tokenString, err := h.security.NewToken(user.Email, sessionId, user.SessionStartTime.Add(time.Duration(time.Hour*4)))
	if err != nil {
		http.SetCookie(w, &emptyAuthCookie)
		h.loginError(err, w, r, "failed to generate token")
		return
	}
	cookie := http.Cookie{Name: "token", Value: tokenString, SameSite: http.SameSiteDefaultMode}
	http.SetCookie(w, &cookie)
	handlers.Redirect("/", w)
}

func (h Handler) loginError(err error, w http.ResponseWriter, r *http.Request, message string, params ...any) {
	fe_errors.FrontendError(w, r, h.logger, err, message, params...)
}

func (h Handler) GetLoginUI(w http.ResponseWriter, r *http.Request) {
	form := form_models.NewLoginForm()
	if err := templ_users.Login(form).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) LogOut(w http.ResponseWriter, r *http.Request) {
	currentUserId, currentUserEmail, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")
	if err := h.userService.LogoutUser(currentUserId); err != nil {
		h.logger.Error(err, "failed to logout user %s", currentUserEmail)
	}
	http.SetCookie(w, &emptyAuthCookie)
	handlers.Redirect("/login", w)
}

func (h Handler) GetRegisterUI(w http.ResponseWriter, r *http.Request) {
	form := form_models.NewRegistrationForm()
	if err := templ_users.Register(form).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		h.logger.Error(err, "failed to render form")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) RegisterUI(w http.ResponseWriter, r *http.Request) {
	form := form_models.NewRegistrationForm()
	name, err := models.GetRequiredPropertyFromRequest(r, "name", "Full Name")
	form.Name = name
	if err != nil {
		form.SetError("name", err.Error())
	}
	email, err := models.GetRequiredPropertyFromRequest(r, "email", "Email")
	form.Email = email
	if err != nil {
		form.SetError("email", err.Error())
	}
	password, err := models.GetRequiredPropertyFromRequest(r, "password", "Password")
	if err != nil {
		form.SetError("password", err.Error())
	}
	password2, err := models.GetRequiredPropertyFromRequest(r, "password2", "Confimation Password")
	if err != nil {
		form.SetError("confirmation password", err.Error())
	}
	if len(form.GetErrors()) > 0 {
		h.logger.Info("Failed to register due to invalid form details")
		if err := templ_users.RegistrationFormContent(form).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render signup form")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	user, err := h.userService.RegisterUser(name, email, password, password2)
	if err != nil {
		form.SetError("register", err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_users.RegistrationFormContent(form).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render signup form")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	h.logger.Info("Successfully created user %s", user.Email)
	loginForm := form_models.NewLoginForm()
	loginForm.Email = user.Email
	if err := templ_users.LoginFormOOB(loginForm).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
