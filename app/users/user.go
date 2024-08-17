package users

import (
	"fmt"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"golang.org/x/crypto/bcrypt"
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

func (u User) setPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(bytes)
	return nil
}

func (u User) checkPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return false
	}
	return true
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

	router.HandleFunc("/login", handler.LoginUI)

	return nil
}
