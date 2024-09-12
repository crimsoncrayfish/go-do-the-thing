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
	Name             string              `json:"name,omitempty"`
	Nicname          string              `json:"nicname,omitmepty"`
	SessionId        string              `json:"session_id,omitempty"`
	SessionStartTime database.SqLiteTime `json:"session_start_time"`
	SessionValidTill database.SqLiteTime `json:"session_valid_till"`
	LastActiveDate   database.SqLiteTime `json:"last_active_date"`
	PasswordHash     string              `json:"password_hash,omitempty"`
	IsDeleted        bool                `json:"is_deleted,omitempty"`
	IsAdmin          bool                `json:"is_admin,omitempty"`
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
	usersRepo, err := InitRepo(dbConnection)
	if err != nil {
		return err
	}
	handler := New(templates, usersRepo, security, logger)

	router.Handle("GET /login", mw_no_auth(http.HandlerFunc(handler.GetLoginUI)))
	router.Handle("POST /login", mw_no_auth(http.HandlerFunc(handler.LoginUI)))
	router.Handle("POST /signup", mw_no_auth(http.HandlerFunc(handler.RegisterUI)))
	router.Handle("POST /logout", mw(http.HandlerFunc(handler.LogOut)))

	return nil
}
