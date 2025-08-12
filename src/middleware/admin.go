package middleware

import (
	"go-do-the-thing/src/helpers"
	"log/slog"
	"net/http"
)

type AdminMiddleware struct {
	logger *slog.Logger
}

func NewAdminMiddleware() AdminMiddleware {
	logger := slog.With("source", "AdminMiddleware")
	return AdminMiddleware{
		logger: logger,
	}
}

func (a *AdminMiddleware) IsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is_admin := helpers.GetIsAdminFromContext(r.Context())

		if !is_admin {
			current_user_id, _, _, err := helpers.GetUserFromContext(r)
			if err != nil {
				a.logger.Error("failed to determine if user has admin rights", "error", err)
				return
			}
			a.logger.Warn("User attempted to access admin feature w/o admin permissions", "user_id", current_user_id)
			return
		}

		next.ServeHTTP(w, r)
	})
}
