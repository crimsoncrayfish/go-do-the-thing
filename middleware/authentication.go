package middleware

import (
	"log"
	"net/http"
)

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Authentication called")
		next.ServeHTTP(w, r)
	})
}
