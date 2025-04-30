package main

import (
	"embed"
	"errors"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/database/repos"
	"go-do-the-thing/src/handlers/home"
	"go-do-the-thing/src/handlers/project"
	"go-do-the-thing/src/handlers/task"
	"go-do-the-thing/src/handlers/users"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	projectService "go-do-the-thing/src/services/project"
	taskService "go-do-the-thing/src/services/task"
	"net/http"
	"os"
)

//go:generate npm run build && templ generate

//go:embed static
var static embed.FS
var faviconLocation string

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, faviconLocation)
}

func main() {
	logger := slog.NewLogger("Main")
	workingDir, err := os.Getwd()
	if err != nil {
		logger.Error(err, "could not get working dir")
		panic(err)
	}
	logger.Info("Running project in dir %s", workingDir)
	jwtHandler := security.NewJwtHandler(workingDir + "/keys/")

	router := http.NewServeMux()
	faviconLocation = workingDir + "/static/img/todo.ico"

	logger.Info("Setting Up Database")
	dbConnection := database.Init("todo")
	defer dbConnection.Connection.Close()

	reposContainer := repos.NewContainer(dbConnection)

	authMiddleware := middleware.NewAuthenticationMiddleware(jwtHandler, *reposContainer.GetUsersRepo())
	loggingMW := middleware.NewLoggingMiddleWare()
	rateLimeter := middleware.NewRateLimiter()
	middleware_full := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMiddleware.Authentication)
	middleware_no_auth := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging)

	users.SetupUserHandler(*reposContainer.GetUsersRepo(), router, middleware_full, middleware_no_auth, jwtHandler)
	project_service := projectService.SetupProjectService(
		*reposContainer.GetProjectsRepo(),
		*reposContainer.GetProjectUsersRepo(),
		*reposContainer.GetRolesRepo(),
		*reposContainer.GetUsersRepo())
	project.SetupProjectHandler(project_service, router, middleware_full)

	task_service := taskService.SetupTaskService(reposContainer)
	task.SetupTodoHandler(task_service, project_service, router, middleware_full)

	home.SetupHomeHandler(router, middleware_full)
	setupStaticContent(router)

	// This is for https
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	logger.Info("Start server")
	if err := server.ListenAndServeTLS("server.crt", "server.key"); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		logger.Info("Something went wrong")
		panic(err)
	}
}

func setupStaticContent(router *http.ServeMux) {
	router.Handle("/static/", http.FileServer(http.FS(static)))
	router.HandleFunc("/favicon.ico", faviconHandler)
}
