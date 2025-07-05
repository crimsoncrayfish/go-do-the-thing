package users

import (
	"go-do-the-thing/src/handlers"
	templ_users "go-do-the-thing/src/handlers/users/templ"
	"go-do-the-thing/src/helpers"
	fe_errors "go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	form_models "go-do-the-thing/src/models/forms"
	projects_service "go-do-the-thing/src/services/project"
	task_service "go-do-the-thing/src/services/task"
	user_service "go-do-the-thing/src/services/user"
	templ_shared "go-do-the-thing/src/shared/templ"
	"net/http"
	"time"
)

type Handler struct {
	security       security.JwtHandler
	userService    *user_service.UserService
	projectService projects_service.ProjectService
	taskService    task_service.TaskService
	logger         slog.Logger
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
	logger := slog.NewLogger(source)
	logger.Info("Setting up users")

	handler := &Handler{
		security:       security,
		userService:    userService,
		logger:         logger,
		projectService: projectService,
		taskService:    taskService,
	}

	router.Handle("GET /login", mw_no_auth(http.HandlerFunc(handler.GetLoginUI)))
	router.Handle("GET /register", mw_no_auth(http.HandlerFunc(handler.GetRegisterUI)))
	router.Handle("POST /login", mw_no_auth(http.HandlerFunc(handler.Login)))
	router.Handle("POST /register", mw_no_auth(http.HandlerFunc(handler.Register)))
	router.Handle("POST /logout", mw(http.HandlerFunc(handler.LogOut)))
}

var emptyAuthCookie = http.Cookie{Name: "token", Value: "", SameSite: http.SameSiteDefaultMode}

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
	if len(form.Errors) > 0 {
		h.logger.Warn("LoginUI: invalid form input - email: %s, errors: %v", email, form.Errors)
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_users.LoginFormContent(form).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "Failed to render template for formData")
			_ = templ_shared.ToastMessage("An error occurred rendering the form", "error").Render(r.Context(), w)
		}
		http.SetCookie(w, &emptyAuthCookie)
		return
	}
	h.logger.Debug("LoginUI: login attempt - email: %s", email)
	user, sessionId, err := h.userService.AuthenticateUser(email, password)
	if err != nil {
		h.logger.Error(err, "LoginUI: login failed - email: %s", email)
		form.Errors["login"] = err.Error()
		w.WriteHeader(http.StatusUnauthorized)
		if err := templ_users.LoginFormContent(form).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "Failed to render template for formData")
			_ = templ_shared.ToastMessage("An error occurred rendering the form", "error").Render(r.Context(), w)
		}
		http.SetCookie(w, &emptyAuthCookie)
		return
	}
	h.logger.Debug("LoginUI: login succeeded - email: %s, user_id: %d", email, user.Id)
	tokenString, err := h.security.NewToken(user.Email, sessionId, user.SessionStartTime.Add(time.Duration(time.Hour*4)))
	if err != nil {
		http.SetCookie(w, &emptyAuthCookie)
		fe_errors.InternalServerError(w, r, h.logger, err, "Failed to display login page")
		return
	}
	cookie := http.Cookie{Name: "token", Value: tokenString, SameSite: http.SameSiteDefaultMode}
	http.SetCookie(w, &cookie)
	handlers.Redirect("/", w)
}

func (h Handler) GetLoginUI(w http.ResponseWriter, r *http.Request) {
	form := form_models.NewLoginForm()
	if err := templ_users.Login(form).Render(r.Context(), w); err != nil {
		fe_errors.InternalServerError(w, r, h.logger, err, "Failed to display login page")
		return
	}
}

func (h Handler) LogOut(w http.ResponseWriter, r *http.Request) {
	current_user_id, currentUserEmail, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		fe_errors.InternalServerError(w, r, h.logger, err, "User authentication failed")
		return
	}
	if err := h.userService.LogoutUser(current_user_id); err != nil {
		h.logger.Error(err, "failed to logout user %s", currentUserEmail)
	}
	http.SetCookie(w, &emptyAuthCookie)
	handlers.Redirect("/login", w)
}

func (h Handler) GetRegisterUI(w http.ResponseWriter, r *http.Request) {
	form := form_models.NewRegistrationForm()
	if err := templ_users.Register(form).Render(r.Context(), w); err != nil {
		fe_errors.InternalServerError(w, r, h.logger, err, "Failed to display registration page")
		return
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
		h.logger.Debug("Failed to register due to invalid form details")
		if err := templ_users.RegistrationFormContent(form).Render(r.Context(), w); err != nil {
			fe_errors.InternalServerError(w, r, h.logger, err, "Failed to display registration form")
		}
		return
	}
	user, err := h.userService.RegisterUser(name, email, password, password2)
	if err != nil {
		form.SetError("register", "failed to register user")
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_users.RegistrationFormContent(form).Render(r.Context(), w); err != nil {
			fe_errors.InternalServerError(w, r, h.logger, err, "Failed to display registration form")
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
		h.logger.Error(err, "failed to create initial project")
	} else {
		_, err := h.taskService.CreateTask(user.Id, project_id, "My First Task", "Complete my first task", &next_year)
		if err != nil {
			h.logger.Error(err, "failed to create initial project")
		}
	}

	h.logger.Debug("Successfully created user %s", user.Email)
	loginForm := form_models.NewLoginForm()
	loginForm.Email = user.Email
	err = templ_users.LoginFormOOB(loginForm).Render(r.Context(), w)
	if err != nil {
		fe_errors.InternalServerError(w, r, h.logger, err, "Failed to complete registration")
		return
	}
}
