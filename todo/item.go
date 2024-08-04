package todo

import (
	"fmt"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"net/http"
)

type Item struct {
	Id          int64                `json:"id,omitempty"`
	Description string               `json:"description" json:"description"`
	Status      ItemStatus           `json:"status" json:"status"`
	AssignedTo  string               `json:"assigned_to" json:"assigned_to"`
	DueDate     *database.SqLiteTime `json:"due_date" json:"due_date"`
	CreatedBy   string               `json:"created_by" json:"created_by"`
	CreateDate  *database.SqLiteTime `json:"create_date" json:"create_date"`
	IsDeleted   bool                 `json:"is_deleted" json:"is_deleted"`
}

type ItemStatus int

const (
	Scheduled ItemStatus = iota
	Completed
)

func InitTodo(router *http.ServeMux, templates helpers.Templates) error {
	fmt.Println("Setting up repo")
	todoRepo, err := Init()
	if err != nil {
		fmt.Println("failed to initialize todo repo")
		return err
	}

	todoHandler := New(todoRepo, templates)
	fmt.Println("Setting up routes")
	//control apis
	router.HandleFunc("GET /api/todo/item/{id}", todoHandler.GetItemAPI)
	router.HandleFunc("GET /api/todo/items", todoHandler.ListItemsAPI)
	router.HandleFunc("PUT /api/todo/item", todoHandler.CreateItemAPI)
	router.HandleFunc("POST /api/todo/item/{id}", todoHandler.UpdateItemStatusAPI)
	router.HandleFunc("DELETE /api/todo/item/{id}", todoHandler.DeleteItemAPI)
	router.HandleFunc("POST /api/todo/item/restore/{id}", todoHandler.DeleteItemAPI)

	//UI apis
	router.HandleFunc("/todo/list", todoHandler.ListItemsUI)
	router.HandleFunc("POST /todo/toggle/{id}", todoHandler.ToggleItemUI)
	router.HandleFunc("POST /todo/item", todoHandler.CreateItemUI)
	router.HandleFunc("DELETE /todo/item/{id}", todoHandler.DeleteItemUI)
	//	router.HandleFunc("POST /todo/restore/{id}", todoHandler.RestoreItemUI)
	return nil
}
