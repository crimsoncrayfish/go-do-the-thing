package main

import (
	"embed"
	"errors"
	"fmt"
	"go-do-the-thing/helpers"
	"go-do-the-thing/home"
	"go-do-the-thing/todo"
	"io"
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
	faviconLocation = workingDir + "/static/img/todo.ico"
	renderer := helpers.NewRenderer(workingDir)
	fmt.Println("Setting up TODO items")
	err = todo.InitTodo(router, *renderer)
	if err != nil {
		println("Failed to initialize todo")
		panic(err)
	}
	setupHello(router)
	home.SetupHome(router, *renderer)
	router.Handle("/static/", http.FileServer(http.FS(static)))
	router.HandleFunc("/favicon.ico", faviconHandler)
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Start server")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Something went wrong")
		panic(err)
	}
}

func setupHello(router *http.ServeMux) {
	helloWorld := func(w http.ResponseWriter, r *http.Request) {
		param := r.PathValue("name")

		var helloString string
		if len(param) == 0 {
			helloString = "Hello World"
		} else {
			helloString = fmt.Sprintf("Hello, %s!", param)
		}

		_, err := io.WriteString(w, helloString)
		if err != nil {
			panic(err)
		}
		return
	}
	router.HandleFunc("/hello/{name}", helloWorld)
	router.HandleFunc("/hello", helloWorld)
}
