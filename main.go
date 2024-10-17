package main

import (
	"embed"
	"errors"

	"go-do-the-thing/src/database"
	"go-do-the-thing/src/database/repos"
	"go-do-the-thing/src/handlers/home"
	"go-do-the-thing/src/handlers/project"
	"go-do-the-thing/src/handlers/todo"
	"go-do-the-thing/src/handlers/users"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"net/http"
	"os"
)

//go:generate npm run build

//go:embed static
var static embed.FS
var faviconLocation string

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, faviconLocation)
}

func main() {
	logger := slog.NewLogger("Main")
	router := http.NewServeMux()
	workingDir, err := os.Getwd()
	if err != nil {
		logger.Error(err, "could not get working dir")
		panic(err)
	}
	logger.Info("Running project in dir %s", workingDir)
	faviconLocation = workingDir + "/static/img/todo.ico"
	logger.Info("Setting up TODO items")

	dbConnection, err := database.Init("todo")
	if err != nil {
		logger.Error(err, "could not initialize the db")
		panic(err)
	}

	jwtHandler, err := security.NewJwtHandler(workingDir + "/keys/")
	if err != nil {
		logger.Error(err, "could not setup jwt handler")
		panic(err)
	}

	reposContainer := repos.NewContainer(dbConnection)
	if err != nil {
		logger.Error(err, "could not initialise repositories")
		panic(err)
	}
	authMiddleware := middleware.NewAuthenticationMiddleware(jwtHandler, *reposContainer.GetUsersRepo())
	loggingMW := middleware.NewLoggingMiddleWare()
	rateLimeter := middleware.NewRateLimiter()
	middleware_full := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMiddleware.Authentication)
	middleware_no_auth := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging)

	users.SetupUserHandler(*reposContainer.GetUsersRepo(), router, middleware_full, middleware_no_auth, jwtHandler)
	project.SetupProjectHandler(*reposContainer.GetProjectsRepo(), *reposContainer.GetProjectUsersRepo(), *reposContainer.GetRolesRepo(), router, middleware_full)
	todo.SetupTodoHandler(*reposContainer.GetTasksRepo(), *reposContainer.GetUsersRepo(), router, middleware_full)

	home.SetupHomeHandler(router, middleware_full)
	setupStaticContent(router)

	//This is for https
	server := http.Server{
		Addr:    ":8079",
		Handler: router,
	}

	logger.Info("Start server")

	if err := server.ListenAndServeTLS("public.key", "private.key"); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		logger.Info("Something went wrong")
		panic(err)
	}
}

func setupStaticContent(router *http.ServeMux) {
	router.Handle("/static/", http.FileServer(http.FS(static)))
	router.HandleFunc("/favicon.ico", faviconHandler)
}
