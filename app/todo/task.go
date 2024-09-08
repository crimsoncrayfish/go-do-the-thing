package todo

import (
	"fmt"
	"go-do-the-thing/app/users"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"go-do-the-thing/middleware"
	"net/http"
	"time"
)

type Task struct {
	Id             int64                `json:"id,omitempty"`
	Name           string               `json:"name"`
	Description    string               `json:"description,omitempty"`
	Status         ItemStatus           `json:"status"`
	CompleteDate   *database.SqLiteTime `json:"complete_date"`
	AssignedTo     string               `json:"assigned_to"`
	DueDate        *database.SqLiteTime `json:"due_date"`
	CreatedBy      string               `json:"created_by"`
	CreateDate     *database.SqLiteTime `json:"create_date"`
	IsDeleted      bool                 `json:"is_deleted"`
	Tag            string               `json:"tag,omitempty"`
	AssignedToUser users.User           `json:"assigned_to_user,omitempty"`
}

type ItemStatus int

const (
	Scheduled ItemStatus = iota
	Completed
)

func (t *Task) toggleStatus() {
	if t.Status == Scheduled {
		t.Status = Completed
		now := time.Now()
		t.CompleteDate = &database.SqLiteTime{Time: &now}
	} else {
		t.Status = Scheduled
		t.CompleteDate = &database.SqLiteTime{}
	}
}

func SetupTodo(
	dbConnection database.DatabaseConnection,
	router *http.ServeMux,
	templates helpers.Templates,
	mw_stack middleware.Middleware,
) error {
	fmt.Println("Setting up todo repo")
	todoRepo, err := InitRepo(dbConnection)
	if err != nil {
		fmt.Println("failed to initialize todo repo")
		return err
	}

	todoHandler := New(todoRepo, templates)
	fmt.Println("Setting up routes")
	router.Handle("GET /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.GetItem)))
	router.Handle("GET /todo/items", mw_stack(http.HandlerFunc(todoHandler.ListItems)))
	router.Handle("POST /todo/item/status/{id}", mw_stack(http.HandlerFunc(todoHandler.UpdateItemStatus)))
	router.Handle("POST /todo/item", mw_stack(http.HandlerFunc(todoHandler.CreateItem)))
	router.Handle("POST /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.UpdateItem)))
	router.Handle("DELETE /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.DeleteItem)))
	router.Handle("GET /error", mw_stack(http.HandlerFunc(todoHandler.TestError)))
	//	router.HandleFunc("POST /todo/restore/{id}", todoHandler.RestoreItemUI)
	return nil
}
