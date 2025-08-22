package main

import (
	"context"
	"embed"
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
	"os/signal"
	"syscall"
	"time"
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
	env := os.Getenv("ENV")
	if env == "" {
		env = DEV
	}
	logger.SetupLogger(env == PROD)
	log := logger.NewLogger("Main")
	log.Debug("This is the env", "env", env)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Debug("This is the port", "port", port)

	dbConnString := os.Getenv("DATABASE_URL")
	if dbConnString == "" {
		log.Fatal("FATAL: DATABASE_URL environment variable not set.")
	}

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get working directory", "error", err)
	}
	jwtHandler := security.NewJwtHandler(env)
	router := http.NewServeMux()
	faviconLocation = workDir + "/static/img/todo.ico"
	manifestLocation = workDir + "/static/json/manifest.json"
	swjsLocation = workDir + "/static/json/manifest.json"

	log.Info("Setting Up Database")
	clientID := os.Getenv("AZURE_CLIENT_ID")

	log.Debug("connecting to postgres db", "connection_string", dbConnString, "client_id", clientID)
	dbConnection := database.Init(dbConnString, clientID)
	defer dbConnection.Close()

	reposContainer := repos.NewContainer(dbConnection)

	authMW := middleware.NewAuthenticationMiddleware(jwtHandler, *reposContainer.GetUsersRepo())
	loggingMW := middleware.NewLoggingMiddleWare()
	rateLimeter := middleware.NewRateLimiter()
	adminMW := middleware.NewAdminMiddleware()

	middlewareAdmin := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMW.Authentication, adminMW.IsAdmin)
	middlewareFull := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMW.Authentication)
	middlewareNoAuth := middleware.CreateStack(rateLimeter.RateLimit, loggingMW.Logging, authMW.OptionalAuthentication)

	userService := user_service.SetupUserService(reposContainer)
	taskService := taskService.SetupTaskService(reposContainer)
	adminService := adminService.SetupAdminService(reposContainer)
	projectService := projectService.SetupProjectService(reposContainer)

	users.SetupUserHandler(router, middlewareFull, middlewareNoAuth, jwtHandler, projectService, taskService, &userService)
	project.SetupProjectHandler(projectService, taskService, router, middlewareFull)
	task.SetupTodoHandler(taskService, projectService, router, middlewareFull)
	home.SetupHomeHandler(router, middlewareFull)
	admin.SetupAdminHandler(adminService, router, middlewareAdmin)
	setupStaticContent(router)

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	serverError := make(chan error)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if env == DEV {
			serverError <- server.ListenAndServeTLS("server.crt", "server.key")
		} else {
			serverError <- server.ListenAndServe()
		}
		log.Info("listening at port %s", port)
	}()

	select {
	case err := <-serverError:
		log.Error("Failed to start server", "error", err)
		panic(fmt.Sprintf("Server startup failed: %v", err))
	case sig := <-shutdown:
		log.Warn("Shutdown signal received: %s", sig)
		defer log.Warn("Shutdown complete")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Error("Server shutdown failed: %v", err)
			if err := server.Close(); err != nil {
				log.Error("Server force shutdown failed: %v", err)
				os.Exit(1)
			}
		}
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
