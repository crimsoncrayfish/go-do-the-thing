package main

import (
	"embed"
	"errors"
	"fmt"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/database/repos"
	"go-do-the-thing/src/handlers/admin"
	"go-do-the-thing/src/handlers/home"
	"go-do-the-thing/src/handlers/project"
	"go-do-the-thing/src/handlers/task"
	"go-do-the-thing/src/handlers/users"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/helpers/security"
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

const (
	DEV  = "development"
	PROD = "production"
)

func main() {
	logger.SetupLogger()
	log := logger.NewLogger("Main")

	env := os.Getenv("ENV")
	if env == "" {
		env = DEV
	}
	log.Debug("This is the env", "env", env)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Debug("This is the port", "port", port)

	var db_connection_string string
	if env == DEV {
		db_connection_string = "postgres://admin:admin@localhost:5432/todo_db?sslmode=disable"
	} else {
		db_connection_string = os.Getenv("DATABASE_URL")
	}

	if db_connection_string == "" {
		log.Fatal("FATAL: DATABASE_URL environment variable not set.")
	}

	working_dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get working directory", "error", err)
	}
	jwtHandler := security.NewJwtHandler(env)
	router := http.NewServeMux()
	faviconLocation = working_dir + "/static/img/todo.ico"
	manifestLocation = working_dir + "/static/json/manifest.json"
	swjsLocation = working_dir + "/static/json/manifest.json"

	log.Info("Setting Up Database")
	clientId := os.Getenv("AZURE_CLIENT_ID")

	log.Debug("connecting to postgres db", "connection_string", db_connection_string, "client_id", clientId)
	dbConnection := database.Init(db_connection_string, clientId)
	defer dbConnection.Close()

	reposContainer := repos.NewContainer(dbConnection)

	authMW := middleware.NewAuthenticationMiddleware(jwtHandler, *reposContainer.GetUsersRepo())
	loggingMW := middleware.NewLoggingMiddleWare()
	rateLimeter := middleware.NewRateLimiter()
	adminMW := middleware.NewAdminMiddleware()

	middleware_admin := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMW.Authentication, adminMW.IsAdmin)
	middleware_full := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMW.Authentication)
	middleware_no_auth := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMW.OptionalAuthentication)

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

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}
	if env == DEV {
		log.Info("Start TLS server")
		err = server.ListenAndServeTLS("server.crt", "server.key")
	} else {
		log.Info("Start server")
		err = server.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("Failed to start server", "error", err)
		panic(fmt.Sprintf("Server startup failed: %v", err))
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
