package middleware

import (
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/slog"
	"net/http"
)

type AdminMiddleware struct {
	Logger slog.Logger
}

func NewAdminMiddleware() AdminMiddleware {
	return AdminMiddleware{
		Logger: slog.NewLogger("IsAdmin"),
	}
}

// API code to intercept all requests
func (a *AdminMiddleware) IsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is_admin := helpers.GetIsAdminFromContext(r.Context())

		if !is_admin {
			current_user_id, _, _, err := helpers.GetUserFromContext(r)
			if err != nil {
				a.Logger.Error(err, "failed to determine if user has admin rights")
				return
			}
			a.Logger.Warn("User attempted to access admin feature w/o admin permissions. User Id: %d", current_user_id)
			return
		}

		next.ServeHTTP(w, r)
	})
}
