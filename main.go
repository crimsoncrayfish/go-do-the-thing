package main

import (
	"embed"
	"errors"
	"fmt"
	"go-do-the-thing/app/home"
	"go-do-the-thing/app/todo"
	"go-do-the-thing/app/users"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/security"
	"go-do-the-thing/middleware"
	"log"
	"net"
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

	auth, err := security.NewJwtHandler(workingDir + "/keys/")
	if err != nil {
		panic(err)
	}
	middleware := middleware.CreateStack(middleware.Logging, auth.Authentication)
	err = users.SetupUsers(dbConnection, router, *renderer, middleware)
	if err != nil {
		println("Failed to initialize users")
		panic(err)
	}

	err = todo.SetupTodo(dbConnection, router, *renderer, middleware)
	if err != nil {
		println("Failed to initialize todo")
		panic(err)
	}
	home.SetupHome(router, *renderer, middleware)

	setupRandom(router)

	//This is for https
	server := http.Server{
		Addr:    ":8079",
		Handler: router,
	}

	fmt.Println("Start server")

	if err := server.ListenAndServeTLS("public.key", "private.key"); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Something went wrong")
		panic(err)
	}
}

func setupRandom(router *http.ServeMux) {
	router.Handle("/static/", http.FileServer(http.FS(static)))
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		_, err := fmt.Fprintf(writer, "Hello World")
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func redirectToHTTPS(tlsPort string) {
	httpSrv := http.Server{
		Addr: ":8081",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, _ := net.SplitHostPort(r.Host)
			u := r.URL
			u.Host = net.JoinHostPort(host, tlsPort)
			u.Scheme = "https"
			log.Println(u.String())
			http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		}),
	}
	log.Println(httpSrv.ListenAndServe())
}
