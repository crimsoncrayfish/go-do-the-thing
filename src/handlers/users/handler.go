package users

import (
	"go-do-the-thing/src/handlers"
	templ_users "go-do-the-thing/src/handlers/users/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	form_models "go-do-the-thing/src/models/forms"
	projects_service "go-do-the-thing/src/services/project"
	task_service "go-do-the-thing/src/services/task"
	user_service "go-do-the-thing/src/services/user"
	templ_shared "go-do-the-thing/src/shared/templ"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	security       security.JwtHandler
	userService    *user_service.UserService
	projectService projects_service.ProjectService
	taskService    task_service.TaskService
	logger         *slog.Logger
}

var source = "UsersHandler"

func SetupUserHandler(
	router *http.ServeMux,
	mw middleware.Middleware,
	mw_no_auth middleware.Middleware,
	security security.JwtHandler,
	projectService projects_service.ProjectService,
	taskService task_service.TaskService,
	userService *user_service.UserService,
) {
	logger := slog.With("source", source)
	logger.Info("Setting up users")

	handler := &Handler{
		security:       security,
		userService:    userService,
		projectService: projectService,
		taskService:    taskService,
		logger:         logger,
	}

	router.Handle("GET /login", mw_no_auth(http.HandlerFunc(handler.GetLoginUI)))
	router.Handle("GET /register", mw_no_auth(http.HandlerFunc(handler.GetRegisterUI)))
	router.Handle("POST /login", mw_no_auth(http.HandlerFunc(handler.Login)))
	router.Handle("POST /register", mw_no_auth(http.HandlerFunc(handler.Register)))
	router.Handle("POST /logout", mw(http.HandlerFunc(handler.LogOut)))
}

var emptyAuthCookie = http.Cookie{
	Name:     "token",
	Value:    "",
	Path:     "/",
	MaxAge:   -1,
	HttpOnly: true,
	Secure:   true,
	SameSite: http.SameSiteLaxMode,
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
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
	remember_me := models.GetPropertyFromRequest(r, "remember-me") == "on"

	if len(form.Errors) > 0 {
		h.logger.Warn("LoginUI: invalid form input", "email", email, "errors", form.Errors)
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_users.LoginFormContent(form).Render(r.Context(), w); err != nil {
			h.logger.Error("Failed to render template for formData", "error", err)
			_ = templ_shared.ToastMessage("An error occurred rendering the form", "error").Render(r.Context(), w)
		}
		http.SetCookie(w, &emptyAuthCookie)
		return
	}
	h.logger.Debug("LoginUI: login attempt", "email", email)
	user, err := h.userService.AuthenticateUser(email, password)
	if err != nil {
		if err.Error() == "user not enabled" {
			form.Errors["login"] = "user not enabled"
		} else {
			h.logger.Error("LoginUI: login failed", "error", err, "email", email)
			form.Errors["login"] = "failed to log in with an error"
		}
		w.WriteHeader(http.StatusUnauthorized)
		if err := templ_users.LoginFormContent(form).Render(r.Context(), w); err != nil {
			h.logger.Error("Failed to render template for formData", "error", err)
			_ = templ_shared.ToastMessage("An error occurred rendering the form", "error").Render(r.Context(), w)
		}
		http.SetCookie(w, &emptyAuthCookie)
		return
	}
	token_lifetime := time.Duration(time.Hour * 24)
	if remember_me {
		token_lifetime = time.Duration(time.Hour * 24 * 7)
	}
	tokenString, err := h.security.NewToken(user.Email, user.SessionId, user.SessionStartTime.Add(token_lifetime))
	if err != nil {
		http.SetCookie(w, &emptyAuthCookie)
		errors.InternalServerError(w, r, err, "Failed to display login page")
		return
	}
	cookie := http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	handlers.Redirect("/", w)
}

func (h Handler) GetLoginUI(w http.ResponseWriter, r *http.Request) {
	isAuthenticated := helpers.IsAuthenticatedFromContext(r)
	if isAuthenticated {
		h.logger.Debug("User already authenticated - redirecting to home")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	form := form_models.NewLoginForm()
	isHtmxRequest := r.Header.Get("HX-Request") == "true"
	if isHtmxRequest {
		if err := templ_users.LoginFormOOB(form).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, err, "Failed to display login page")
			return
		}
	} else {
		if err := templ_users.Login(form).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, err, "Failed to display login page")
			return
		}
	}
}

func (h Handler) LogOut(w http.ResponseWriter, r *http.Request) {
	current_user_id, currentUserEmail, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		h.logger.Warn("logout failed - user auth failed")
		errors.InternalServerError(w, r, err, "User authentication failed")
		return
	}
	if err := h.userService.LogoutUser(current_user_id); err != nil {
		h.logger.Error("failed to logout user", "error", err, "email", currentUserEmail)
	}
	http.SetCookie(w, &emptyAuthCookie)
	handlers.RedirectFull("/login", w)
}

func (h Handler) GetRegisterUI(w http.ResponseWriter, r *http.Request) {
	isAuthenticated := helpers.IsAuthenticatedFromContext(r)
	if isAuthenticated {
		h.logger.Debug("User already authenticated - redirecting to home")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	form := form_models.NewRegistrationForm()

	isHtmxRequest := r.Header.Get("HX-Request") == "true"
	if isHtmxRequest {
		if err := templ_users.RegistrationFormContianerOOB(form).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, err, "Failed to display registration page")
			return
		}
	} else {
		if err := templ_users.Register(form).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, err, "Failed to display registration page")
			return
		}
	}
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
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
		if err := templ_users.RegistrationFormContent(form).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, err, "Failed to display registration form")
		}
		return
	}
	user, err := h.userService.RegisterUser(name, email, password, password2)
	if err != nil {
		form.SetError("register", "failed to register user")
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_users.RegistrationFormContent(form).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, err, "Failed to display registration form")
		}
		return
	}
	now := time.Now()
	next_year := now.Add(time.Duration(time.Hour * 24 * 365))
	project_id, err := h.projectService.CreateProject(
		user.Id, user.Id,
		"My First Project", "This is my default project",
		&now, &next_year,
	)
	if err != nil {
		h.logger.Error("failed to create initial project", "error", err)
	} else {
		_, err := h.taskService.CreateTask(user.Id, project_id, "My First Task", "Complete my first task", &next_year)
		if err != nil {
			h.logger.Error("failed to create initial task", "error", err)
		}
	}

	loginForm := form_models.NewLoginForm()
	loginForm.Email = user.Email
	err = templ_shared.RenderTempls(
		templ_users.LoginFormOOB(loginForm),
		templ_shared.ToastMessage("Registered successfully!", "success"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, err, "Failed to display new task")
		return
	}
}
