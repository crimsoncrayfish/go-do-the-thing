package users

import (
	"fmt"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/security"
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
	security security.JwtHandler,
) error {
	fmt.Println("Setting up users")
	usersRepo, err := InitRepo(dbConnection)
	if err != nil {
		return err
	}
	handler := New(templates, usersRepo, security)

	router.HandleFunc("GET /login", handler.GetLoginUI)
	router.HandleFunc("POST /login", handler.LoginUI)
	router.HandleFunc("POST /signup", handler.Signup)
	router.Handle("POST /logout", mw(http.HandlerFunc(handler.LogOut)))

	return nil
}
