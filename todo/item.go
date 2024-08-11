package todo

import (
	"fmt"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"net/http"
)

type Item struct {
	Id          int64                `json:"id,omitempty"`
	Description string               `json:"description"`
	Status      ItemStatus           `json:"status"`
	AssignedTo  string               `json:"assigned_to"`
	DueDate     *database.SqLiteTime `json:"due_date"`
	CreatedBy   string               `json:"created_by"`
	CreateDate  *database.SqLiteTime `json:"create_date"`
	IsDeleted   bool                 `json:"is_deleted"`
	Tag         string               `json:"tag,omitempty"`
}

type ItemStatus int

const (
	Scheduled ItemStatus = iota
	Completed
)

func SetupTodo(
	dbConnection database.DatabaseConnection,
	router *http.ServeMux,
	templates helpers.Templates,
) error {
	fmt.Println("Setting up repo")
	todoRepo, err := Init(dbConnection)
	if err != nil {
		fmt.Println("failed to initialize todo repo")
		return err
	}

	todoHandler := New(todoRepo, templates)
	fmt.Println("Setting up routes")
	router.HandleFunc("GET /todo/item/{id}", todoHandler.GetItem)
	router.HandleFunc("GET /todo/items", todoHandler.ListItems)
	router.HandleFunc("POST /todo/item/status/{id}", todoHandler.UpdateItemStatus)
	router.HandleFunc("POST /todo/item", todoHandler.CreateItem)
	router.HandleFunc("POST /todo/item/{id}", todoHandler.UpdateItem)
	router.HandleFunc("DELETE /todo/item/{id}", todoHandler.DeleteItem)
	router.HandleFunc("GET /error", todoHandler.TestError)
	//	router.HandleFunc("POST /todo/restore/{id}", todoHandler.RestoreItemUI)
	return nil
}
