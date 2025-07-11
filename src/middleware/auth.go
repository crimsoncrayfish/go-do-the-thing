package middleware

import (
	"context"
	"errors"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/constants"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/helpers/slog"
	"net/http"
	"strconv"
	"time"
)

type AuthenticationMiddleware struct {
	JwtHandler security.JwtHandler
	UsersRepo  users_repo.UsersRepo
	Logger     slog.Logger
}

func NewAuthenticationMiddleware(jwt security.JwtHandler, usersRepo users_repo.UsersRepo) AuthenticationMiddleware {
	return AuthenticationMiddleware{
		Logger:     slog.NewLogger("Auth"),
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
			clearTokenCookie(w)
			redirectOnErr(w, r, a.Logger, err, "/login", "Invalid token")
			return
		}

		// Validate token expiry time
		exp := claims["expiry"]
		if exp == "" {
			clearTokenCookie(w)
			redirectOnErr(w, r, a.Logger, errors.New("no expiry on token"), "/login", "Invalid token, expiry time missing")
			return
		}
		userId := claims["user_id"]
		if userId == "" {
			clearTokenCookie(w)
			redirectOnErr(w, r, a.Logger, errors.New("no user in token"), "/login", "Invalid token, user_id missing")
			return
		}
		session := claims["session_id"]
		if session == "" {
			clearTokenCookie(w)
			redirectOnErr(w, r, a.Logger, errors.New("no session id in token"), "/login", "Invalid token, session missing")
			return
		}

		expDate, err := time.Parse(constants.DateTimeFormat, exp.(string))
		if err != nil {
			clearTokenCookie(w)
			redirectOnErr(w, r, a.Logger, err, "/login", "Token expiry malformed %s", exp)
			// TODO : redo this to pass in params for message
			return
		}
		if expDate.Before(time.Now()) {
			clearTokenCookie(w)

			// TODO: reset session for user
			redirectOnErr(w, r, a.Logger, errors.New("token expired"), "/login", "Token expired for user_id %d. Exp date: %s", userId, exp)
			return
		}

		user, err := a.UsersRepo.GetUserByEmail(userId.(string))
		if err != nil {
			clearTokenCookie(w)
			redirectOnErr(w, r, a.Logger, err, "/login", "failed to get user from db")
			return
		}

		if user.SessionId != session.(string) {
			clearTokenCookie(w)
			redirectOnErr(w, r, a.Logger, errors.New("session id does not match user session id"), "/login", "session id mismatch")
			return
		}

		shouldRefresh := expDate.After(time.Now().Add(time.Duration(time.Hour)))
		if shouldRefresh {
			// TODO: refresh token
			a.Logger.Info("Token for user %s is close to expiring", userId)
		}

		values := helpers.HttpContext{Values: map[constants.ContextKey]string{
			constants.AuthUserId:    strconv.FormatInt(user.Id, 10),
			constants.AuthUserEmail: user.Email,
			constants.AuthUserName:  user.FullName,
			constants.AuthIsAdmin:   strconv.FormatBool(user.IsAdmin),
		}}
		ctx := context.WithValue(r.Context(), constants.AuthContext, values)
		request := r.WithContext(ctx)

		next.ServeHTTP(w, request)
	})
}

func clearTokenCookie(w http.ResponseWriter) {
	clearCookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}
	http.SetCookie(w, &clearCookie)
}

func redirectOnErr(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, location, message string, params ...any) {
	// TODO: Add ability to let user know why redirect happened (message on screen?)
	logger.Error(err, message, params...)
	w.Header().Set("HX-Location", location)
	http.Redirect(w, r, location, http.StatusSeeOther)
}
