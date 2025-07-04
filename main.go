package main

import (
	"embed"
	"errors"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/database/repos"
	"go-do-the-thing/src/handlers/admin"
	"go-do-the-thing/src/handlers/home"
	"go-do-the-thing/src/handlers/project"
	"go-do-the-thing/src/handlers/task"
	"go-do-the-thing/src/handlers/users"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	adminService "go-do-the-thing/src/services/admin"
	projectService "go-do-the-thing/src/services/project"
	taskService "go-do-the-thing/src/services/task"
	user_service "go-do-the-thing/src/services/user"
	"net/http"
	"os"
)

//go:generate npm run build && templ generate

//go:embed static
var static embed.FS

var (
	faviconLocation  string
	manifestLocation string
	swjsLocation     string
)

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, faviconLocation)
}

func main() {
	logger := slog.NewLogger("Main")
	workingDir, err := os.Getwd()
	assert.NoError(err, "Main", "could not get working dir")
	logger.Info("Running project in dir %s", workingDir)
	jwtHandler := security.NewJwtHandler(workingDir + "/keys/")

	router := http.NewServeMux()
	faviconLocation = workingDir + "/static/img/todo.ico"
	manifestLocation = workingDir + "/static/json/manifest.json"
	swjsLocation = workingDir + "/static/json/manifest.json"

	logger.Info("Setting Up Database")
	connectionString := "postgres://admin:admin@localhost:5432/todo_db?sslmode=disable"
	dbConnection := database.Init(connectionString)
	defer dbConnection.Close()

	reposContainer := repos.NewContainer(dbConnection)

	authMW := middleware.NewAuthenticationMiddleware(jwtHandler, *reposContainer.GetUsersRepo())
	loggingMW := middleware.NewLoggingMiddleWare()
	rateLimeter := middleware.NewRateLimiter()
	adminMW := middleware.NewAdminMiddleware()

	middleware_admin := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMW.Authentication, adminMW.IsAdmin)
	middleware_full := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMW.Authentication)
	middleware_no_auth := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging)

	user_service := user_service.SetupUserService(reposContainer)
	task_service := taskService.SetupTaskService(reposContainer)
	admin_service := adminService.SetupAdminService(reposContainer)
	project_service := projectService.SetupProjectService(reposContainer)

	users.SetupUserHandler(router, middleware_full, middleware_no_auth, jwtHandler, project_service, task_service, &user_service)
	project.SetupProjectHandler(project_service, task_service, router, middleware_full)
	task.SetupTodoHandler(task_service, project_service, router, middleware_full)
	home.SetupHomeHandler(router, middleware_full)
	admin.SetupAdminHandler(admin_service, router, middleware_admin)
	setupStaticContent(router)

	// This is for https
	server := http.Server{
		Addr:    ":8079",
		Handler: router,
	}

	logger.Info("Start server")
	err = server.ListenAndServeTLS("server.crt", "server.key")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Info("Something went wrong")
		assert.NoError(err, "Main", "Failed to start HTTPS server")
	}
}

func setupStaticContent(router *http.ServeMux) {
	router.Handle("/static/", http.FileServer(http.FS(static)))
	router.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, manifestLocation)
	})
	http.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, swjsLocation)
	})
}
