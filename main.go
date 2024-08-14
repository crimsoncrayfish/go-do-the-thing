package main

import (
	"embed"
	"errors"
	"fmt"
	"go-do-the-thing/app/home"
	"go-do-the-thing/app/todo"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"go-do-the-thing/middleware"
	"net/http"
	"os"
)

//go:generate npm run build

//go:embed static
var static embed.FS
var faviconLocation string

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, faviconLocation)
}

func main() {
	router := http.NewServeMux()
	workingDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
		panic(err)
	}
	fmt.Printf("Running project in dir %s\n", workingDir)
	faviconLocation = workingDir + "/static/img/todo.ico"
	renderer := helpers.NewRenderer(workingDir)
	fmt.Println("Setting up TODO items")

	dbConnection, err := database.Init("todo")
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	err = todo.SetupTodo(dbConnection, router, *renderer)
	if err != nil {
		println("Failed to initialize todo")
		panic(err)
	}
	home.SetupHome(router, *renderer)
	router.Handle("/static/", http.FileServer(http.FS(static)))
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		_, err := fmt.Fprintf(writer, "Hello World")
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
		}
	})
	stack := middleware.CreateStack(middleware.Logging, middleware.Authentication)
	server := http.Server{
		Addr:    ":8080",
		Handler: stack(router),
	}

	fmt.Println("Start server")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Something went wrong")
		panic(err)
	}
}
