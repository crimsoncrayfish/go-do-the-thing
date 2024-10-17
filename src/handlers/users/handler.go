package users

import (
	"database/sql"
	"errors"
	"go-do-the-thing/src/database"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/handlers"
	templ_users "go-do-the-thing/src/handlers/users/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	form_models "go-do-the-thing/src/models/forms"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	security security.JwtHandler
	repo     users_repo.UsersRepo
	logger   slog.Logger
}

func SetupUserHandler(
	userRepo users_repo.UsersRepo,
	router *http.ServeMux,
	mw middleware.Middleware,
	mw_no_auth middleware.Middleware,
	security security.JwtHandler,
) {
	logger := slog.NewLogger("UsersHandler")
	logger.Info("Setting up users")

	handler := &Handler{
		repo:     userRepo,
		security: security,
		logger:   logger,
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
	email, err := models.GetPropertyFromRequest(r, "email", "Email", true)
	form.Email = email
	if err != nil {
		form.Errors["email"] = err.Error()
	}

	password, err := models.GetPropertyFromRequest(r, "password", "Password", true)
	// NOTE: Dont add the password back in to the form as i dont want to send it back and forth
	if err != nil {
		form.Errors["password"] = err.Error()
	}
	if len(form.Errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_users.LoginFormContent(form).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, h.logger, "Failed to render template for formData")
		}
		http.SetCookie(w, &emptyAuthCookie)
		return
	}
	user, err := h.repo.GetUserByEmail(email)

	if err != nil {
		// NOTE: Not a valid user but Shhhh! dont tell them
		// TODO: Keep track of accounts that have invalid logins and lock them after a set amount of login attempts
		http.SetCookie(w, &emptyAuthCookie)
		if errors.Is(err, sql.ErrNoRows) {
			h.invalidLogin(w, r, form, "User '%s' not in database", email)
			return
		}
		h.loginError(err, w, r, form, "Failed to read user from DB with email %s", email)
		return
	}

	passwordHash, err := h.repo.GetUserPassword(user.Id)
	if err != nil {
		http.SetCookie(w, &emptyAuthCookie)
		h.loginError(err, w, r, form, "Failed to read password for user %d", user.Id)
		return
	}

	if !security.CheckPassword(password, passwordHash) {
		// NOTE: Not a valid password but Shhhh! dont tell them
		// TODO: Keep track of accounts that have invalid logins and lock them after a set amount of login attempts
		// TODO: keep track of IPs that have invalid logins and ban them after a set count
		h.invalidLogin(w, r, form, "Invalid password")
		http.SetCookie(w, &emptyAuthCookie)
		return
	}

	user.SessionId = uuid.New().String()
	user.SessionStartTime = database.SqLiteNow()

	if err := h.repo.UpdateSession(user.Id, user.SessionId, user.SessionStartTime); err != nil {
		h.loginError(err, w, r, form, "Failed to set session id for user %d", user.Id)
		http.SetCookie(w, &emptyAuthCookie)
		return
	}
	tokenString, err := h.security.NewToken(user.Email, user.SessionId, user.SessionStartTime.Time.Add(time.Duration(time.Hour*4)))
	if err != nil {
		// NOTE: Failed to create a token. Hmmm. Should probably throw internalServerErr
		h.loginError(err, w, r, form, "failed to generate token")
		http.SetCookie(w, &emptyAuthCookie)
		return
	}
	cookie := http.Cookie{Name: "token", Value: tokenString, SameSite: http.SameSiteDefaultMode}
	http.SetCookie(w, &cookie)
	handlers.Redirect("/", w)
}

func (h Handler) loginError(err error, w http.ResponseWriter, r *http.Request, form form_models.LoginForm, message string, params ...any) {
	h.logger.Error(err, message, params...)
	form.SetError("Failed Login", "Something went wrong on the server. Please try again.")
	http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	if err := templ_users.LoginFormContent(form).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render task login form content")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) invalidLogin(w http.ResponseWriter, r *http.Request, form form_models.LoginForm, message string, params ...any) {
	h.logger.Info(message, params...)
	form.SetError("Failed Login", "Invalid login credentials")
	if err := templ_users.LoginFormContent(form).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err, "failed to render task login form content")
	}
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
	// NOTE: confirm logged in
	currentUserId, currentUserEmail, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	if err := h.repo.UpdateSession(currentUserId, "", &database.SqLiteTime{}); err != nil {
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
	name, err := models.GetPropertyFromRequest(r, "name", "Full Name", true)
	form.Name = name
	if err != nil {
		form.SetError("name", err.Error())
	}
	email, err := models.GetPropertyFromRequest(r, "email", "Email", true)
	form.Email = email
	if err != nil {
		form.SetError("email", err.Error())
	}
	password, err := models.GetPropertyFromRequest(r, "password", "Password", true)
	if err != nil {
		form.SetError("password", err.Error())
	}
	password2, err := models.GetPropertyFromRequest(r, "password2", "Confimation Password", true)
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
	if password != password2 {
		h.logger.Info("Failed to register due to passwords not matching")
		form.SetError("Password", "The password fields dont match")
		if err := templ_users.RegistrationFormContent(form).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render signup form")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	user, _ := h.repo.GetUserByEmail(email) // TODO: what to do with this err message
	if (models.User{}) != user {
		h.logger.Info("Registration failure. Email %s already in use", email)
		form.SetError("Email", "Email already in use")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		if err := templ_users.RegistrationFormContent(form).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render signup form")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	passwordHash, err := security.SetPassword(password)
	if err != nil {
		h.logger.Error(err, "Failed to hash password")
		form.SetError("Create", "Failed to create users due to a server error")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		if err := templ_users.RegistrationFormContent(form).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render signup form")
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
	h.logger.Info("Successfully created user %s", user.Email)
	loginForm := form_models.NewLoginForm()
	loginForm.Email = user.Email
	if err := templ_users.LoginFormOOB(loginForm).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
