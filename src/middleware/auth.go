package middleware

import (
	"context"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/constants"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/helpers/security"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type AuthenticationMiddleware struct {
	JwtHandler security.JwtHandler
	UsersRepo  users_repo.UsersRepo
	logger     *slog.Logger
}

func NewAuthenticationMiddleware(jwt security.JwtHandler, usersRepo users_repo.UsersRepo) AuthenticationMiddleware {
	logger := slog.With("source", "AuthMiddleware")
	return AuthenticationMiddleware{
		UsersRepo:  usersRepo,
		JwtHandler: jwt,
		logger:     logger,
	}
}

func (a *AuthenticationMiddleware) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, err := a.tryAuthenticateUser(r)
		if err != nil {
			clearTokenCookie(w)
			a.logger.Warn("Authentication failed", "error", err, "message", "Missing token")
			w.Header().Set("HX-Location", "/login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, request)
	})
}

func (a *AuthenticationMiddleware) OptionalAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, _ := a.tryAuthenticateUser(r)
		next.ServeHTTP(w, request)
	})
}

func (a *AuthenticationMiddleware) tryAuthenticateUser(r *http.Request) (*http.Request, error) {
	var token string
	for _, c := range r.Cookies() {
		if c.Name == "token" {
			token = c.Value
			break
		}
	}
	if token == "" {
		return r, errors.New(errors.AuthFailed, "Missing token cookie")
	}

	claims, err := a.JwtHandler.ValidateToken(token)
	if err != nil {
		return r, errors.New(errors.AuthFailed, "Invalid token: %w", err)
	}

	exp, ok := claims["expiry"].(string)
	if !ok || exp == "" {
		return r, errors.New(errors.AuthFailed, "Invalid token, expiry time missing or malformed")
	}
	userId, ok := claims["user_id"].(string)
	if !ok || userId == "" {
		return r, errors.New(errors.AuthFailed, "Invalid token, user_id missing or malformed")
	}
	session, ok := claims["session_id"].(string)
	if !ok || session == "" {
		return r, errors.New(errors.AuthFailed, "Invalid token, session_id missing or malformed")
	}

	expDate, err := time.Parse(constants.DateTimeFormat, exp)
	if err != nil {
		return r, errors.New(errors.AuthFailed, "Token expiry malformed: %s, error: %w", exp, err)
	}
	if expDate.Before(time.Now()) {
		logger.TODO("Reset session for user when token expires")
		return r, errors.New(errors.AuthFailed, "Token expired for user_id %s. Exp date: %s", userId, exp)
	}

	user, err := a.UsersRepo.GetUserByEmail(userId)
	if err != nil {
		return r, errors.New(errors.AuthFailed, "Failed to get user from DB for user_id %s: %w", userId, err)
	}

	if user.SessionId != session {
		return r, errors.New(errors.AuthFailed, "Session ID mismatch for user %s. Token session: %s, User session: %s", userId, session, user.SessionId)
	}

	shouldRefresh := expDate.After(time.Now().Add(time.Duration(time.Hour)))
	if shouldRefresh {
		logger.TODO("Implement automatic token refresh when close to expiring")
		a.logger.Info("Token for user is close to expiring", "userId", userId)
	}

	values := helpers.HttpContext{Values: map[constants.ContextKey]string{
		constants.AuthIsAuthenticated: strconv.FormatBool(true),
		constants.AuthUserId:          strconv.FormatInt(user.Id, 10),
		constants.AuthUserEmail:       user.Email,
		constants.AuthUserName:        user.FullName,
		constants.AuthIsAdmin:         strconv.FormatBool(user.IsAdmin),
	}}
	ctx := context.WithValue(r.Context(), constants.AuthContext, values)
	request := r.WithContext(ctx)

	return request, nil
}

func clearTokenCookie(w http.ResponseWriter) {
	clearCookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &clearCookie)
}
