package security

import (
	"context"
	"errors"
	"fmt"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtHandler struct {
	*SecretKeyProvider
	logger *slog.Logger
}

func NewJwtHandler(keysLocation string) (JwtHandler, error) {
	keyProvider, err := newKeyProvider(keysLocation)
	if err != nil {
		return JwtHandler{}, err
	}
	return JwtHandler{
		SecretKeyProvider: keyProvider,
		logger:            slog.NewLogger("JWT Handler"),
	}, nil
}

const AuthUserId = "security.middleware.userId"

// API code to intercept all requests
func (s *JwtHandler) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			helpers.HttpError("No token provided. Should probably redirect to login screen", errors.New("no token provided"), w)
			return
		}
		tokenParts := strings.Split(token, " ")
		if tokenParts[0] != "Bearer" {
			helpers.HttpError("Malformed bearer token", errors.New("malformed bearer token"), w)
			return
		}
		claims, err := s.ValidateToken(tokenParts[1])
		if err != nil {
			helpers.HttpError("Invalid token", err, w)
			return
		}
		userId := claims["user_id"]
		if userId == "" {
			helpers.HttpError("Invalid token, user_id malformed", err, w)
			return
		}
		exp := claims["expiry"]
		if exp == "" {
			helpers.HttpError("Invalid token, expiry malformed", err, w)
			return
		}

		// TODO: if the expiry time is passed then redirect to logout?

		ctx := context.WithValue(r.Context(), AuthUserId, userId)
		request := r.WithContext(ctx)

		next.ServeHTTP(w, request)
	})
}

func (s *JwtHandler) NewToken(userId string) (string, error) {
	expiry := time.Now().Add(time.Duration(time.Hour * 24))
	claims := jwt.MapClaims{}
	claims["user_id"] = userId
	claims["expiry"] = expiry
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.getKey())
}

func (s *JwtHandler) ValidateToken(signedToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(signedToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		return s.SecretKeyProvider.getKey().PublicKey, nil
	})
	if err != nil || token == nil {
		switch {

		case errors.Is(err, jwt.ErrTokenMalformed):
			// todo figure out wtf this even is
			s.logger.Error(err, "Not a token")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			// todo figure out wtf this even is
			s.logger.Error(err, "Invalid Signature")
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			// todo figure out wtf this even is
			s.logger.Error(err, "Timing issues")
		case token == nil:
			s.logger.Error(err, "what even happened. Nil token")
		default:
			// todo figure out wtf this even is
			s.logger.Error(err, "What even happened")
		}
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("Token not valid")
}
