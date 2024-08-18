package users

import (
	"fmt"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"net/http"
)

type User struct {
	Id               int                 `json:"id,omitempty"`
	Name             string              `json:"name,omitempty"`
	Surname          string              `json:"surname,omitempty"`
	Email            string              `json:"email,omitempty"`
	SessionId        string              `json:"session_id,omitempty"`
	SessionStartTime database.SqLiteTime `json:"session_start_time"`
	PasswordHash     string              `json:"password_hash,omitempty"`
	IsDeleted        bool                `json:"is_deleted,omitempty"`
	IsAdmin          bool                `json:"is_admin,omitempty"`
}

func SetupUsers(
	dbConnection database.DatabaseConnection,
	router *http.ServeMux,
	templates helpers.Templates,
) error {
	fmt.Println("Setting up users")
	usersRepo, err := InitRepo(dbConnection)
	if err != nil {
		return err
	}
	handler := NewHandler(templates, usersRepo)

	router.HandleFunc("GET /login", handler.GetLogin)
	router.HandleFunc("POST /login", handler.Login)
	//	router.HandleFunc("POST /logout", handler)
	router.HandleFunc("POST /signup", handler.Signup)

	return nil
}
