package users

import (
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/security"
	"go-do-the-thing/helpers/slog"
	"go-do-the-thing/middleware"
	"net/http"
)

type User struct {
	Id               int                 `json:"id,omitempty"`
	Email            string              `json:"email,omitempty"`
	FullName         string              `json:"full_name,omitmepty"`
	SessionId        string              `json:"session_id,omitempty"`
	SessionStartTime database.SqLiteTime `json:"session_start_time"`
	SessionValidTill database.SqLiteTime `json:"session_valid_till"`
	LastActiveDate   database.SqLiteTime `json:"last_active_date"`
	PasswordHash     string              `json:"password_hash,omitempty"`
	IsDeleted        bool                `json:"is_deleted,omitempty"`
	IsAdmin          bool                `json:"is_admin,omitempty"`
	CreateDate       database.SqLiteTime `json:"create_date"`
}

func SetupUsers(
	dbConnection database.DatabaseConnection,
	router *http.ServeMux,
	templates helpers.Templates,
	mw middleware.Middleware,
	mw_no_auth middleware.Middleware,
	security security.JwtHandler,
) error {
	logger := slog.NewLogger("Users")
	logger.Info("Setting up users")
	usersRepo, err := InitRepo(dbConnection, logger)
	if err != nil {
		logger.Error(err, "Failed to initialize repo")
		return err
	}
	handler := New(templates, usersRepo, security, logger)

	router.Handle("GET /login", mw_no_auth(http.HandlerFunc(handler.GetLoginUI)))
	router.Handle("GET /register", mw_no_auth(http.HandlerFunc(handler.GetRegisterUI)))
	router.Handle("POST /login", mw_no_auth(http.HandlerFunc(handler.LoginUI)))
	router.Handle("POST /signup", mw_no_auth(http.HandlerFunc(handler.RegisterUI)))
	router.Handle("POST /logout", mw(http.HandlerFunc(handler.LogOut)))

	router.HandleFunc("GET /users", handler.GetAll)
	return nil
}
