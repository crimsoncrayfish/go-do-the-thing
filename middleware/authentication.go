package middleware

import (
	"log"
	"net/http"
)

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")
		log.Printf("Authentication called with token: %s", token)
		next.ServeHTTP(w, r)
	})
}
