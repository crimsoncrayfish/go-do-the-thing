package users

import (
	usersRepo "go-do-the-thing/app/users/repo"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/security"
	"go-do-the-thing/helpers/slog"
	"go-do-the-thing/middleware"
	"net/http"
)

func SetupUserHandler(
	userRepo usersRepo.Repo,
	router *http.ServeMux,
	templates helpers.Templates,
	mw middleware.Middleware,
	mw_no_auth middleware.Middleware,
	security security.JwtHandler,
) error {
	logger := slog.NewLogger("Users")
	logger.Info("Setting up users")
	handler := New(templates, userRepo, security, logger)

	router.Handle("GET /login", mw_no_auth(http.HandlerFunc(handler.GetLoginUI)))
	router.Handle("GET /register", mw_no_auth(http.HandlerFunc(handler.GetRegisterUI)))
	router.Handle("POST /login", mw_no_auth(http.HandlerFunc(handler.LoginUI)))
	router.Handle("POST /signup", mw_no_auth(http.HandlerFunc(handler.RegisterUI)))
	router.Handle("POST /logout", mw(http.HandlerFunc(handler.LogOut)))

	router.HandleFunc("GET /users", handler.GetAll)
	return nil
}
