package middleware

import (
	"context"
	"errors"
	"go-do-the-thing/app/repos"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/security"
	"go-do-the-thing/helpers/slog"
	"net/http"
	"time"
)

type AuthenticationMiddleware struct {
	JwtHandler security.JwtHandler
	UsersRepo  repos.UsersRepo
	Logger     slog.Logger
}

func NewAuthenticationMiddleware(jwt security.JwtHandler, usersRepo repos.UsersRepo) AuthenticationMiddleware {
	return AuthenticationMiddleware{
		Logger:     *slog.NewLogger("Auth"),
		UsersRepo:  usersRepo,
		JwtHandler: jwt,
	}
}

// API code to intercept all requests
func (a *AuthenticationMiddleware) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		for _, c := range r.Cookies() {
			if c.Name == "token" {
				token = c.Value
			}
		}
		if token == "" {
			redirectOnErr(w, r, a.Logger, errors.New("cookiejar is empty"), "/login", "Missing token")
			return
		}

		claims, err := a.JwtHandler.ValidateToken(token)
		if err != nil {
			redirectOnErr(w, r, a.Logger, err, "/login", "Invalid token")
			return
		}

		//Validate token expiry time
		exp := claims["expiry"]
		if exp == "" {
			redirectOnErr(w, r, a.Logger, errors.New("no expiry on token"), "/login", "Invalid token, expiry time missing")
			return
		}
		userId := claims["user_id"]
		if userId == "" {
			redirectOnErr(w, r, a.Logger, errors.New("no user in token"), "/login", "Invalid token, user_id missing")
			return
		}
		session := claims["session_id"]
		if session == "" {
			redirectOnErr(w, r, a.Logger, errors.New("no session id in token"), "/login", "Invalid token, session missing")
			return
		}

		expDate, err := time.Parse(helpers.DateTimeFormat, exp.(string))
		if err != nil {
			redirectOnErr(w, r, a.Logger, err, "/login", "Token expiry malformed %s", exp)
			//TODO : redo this to pass in params for message
			return
		}
		if expDate.Before(time.Now()) {
			// TODO: reset session for user
			redirectOnErr(w, r, a.Logger, errors.New("token expired"), "/login", "Token expired for user_id %d. Exp date: %s", userId, exp)
			return
		}

		user, err := a.UsersRepo.GetUserByEmail(userId.(string))
		if err != nil {
			redirectOnErr(w, r, a.Logger, err, "/login", "failed to get user from db")
			return
		}

		if user.SessionId != session.(string) {
			redirectOnErr(w, r, a.Logger, errors.New("session id does not match user session id"), "/login", "session id mismatch")
			return
		}

		shouldRefresh := expDate.After(time.Now().Add(time.Duration(time.Hour)))
		if shouldRefresh {
			// TODO: refresh token
			a.Logger.Info("Token for user %s is close to expiring", userId)
		}

		values := helpers.HttpContext{Values: map[string]string{
			helpers.AuthUserId:   user.Email,
			helpers.AuthUserName: user.FullName,
		}}
		ctx := context.WithValue(r.Context(), helpers.AuthContext, values)
		request := r.WithContext(ctx)

		next.ServeHTTP(w, request)
	})
}
func redirectOnErr(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, location, message string, params ...any) {
	// TODO: Add ability to let user know why redirect happened (message on screen?)
	logger.Error(err, message, params...)
	w.Header().Set("HX-Location", location)
	http.Redirect(w, r, location, http.StatusSeeOther)
}
