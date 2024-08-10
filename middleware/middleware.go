package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func CreateStack(stack ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(stack) - 1; i >= 0; i-- {
			mw := stack[i]
			next = mw(next)
		}
		return next
	}
}
