package security

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/slog"
	"net/http"
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
		var token string
		for _, c := range r.Cookies() {
			if c.Name == "token" {
				token = c.Value
			}
		}
		if token == "" {
			helpers.HttpError("Missing token", errors.New("no token in cookiejar"), w)
			return
		}

		claims, err := s.ValidateToken(token)
		if err != nil {
			helpers.HttpError("Invalid token", err, w)
			return
		}
		userId := claims["user_id"]
		if userId == "" {
			helpers.HttpError("Invalid token, user_id missing", err, w)
			return
		}
		exp := claims["expiry"]
		if exp == "" {
			helpers.HttpError("Invalid token, expiry time missing", err, w)
			return
		}
		session := claims["session_id"]
		if session == "" {
			helpers.HttpError("Invalid token, session missing", err, w)
			return
		}
		// TODO: validate all token claims
		// TODO: refresh token???

		ctx := context.WithValue(r.Context(), AuthUserId, userId)
		request := r.WithContext(ctx)

		next.ServeHTTP(w, request)
	})
}

// TODO: Set token as cookie here
func (s *JwtHandler) NewToken(userId, session string, expiry time.Time) (string, error) {
	claims := jwt.MapClaims{}
	claims["user_id"] = userId
	claims["expiry"] = expiry
	claims["session_id"] = session
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.getKey())
}

func (s *JwtHandler) ValidateToken(signedToken string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(signedToken, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			s.logger.Error(errors.New("unexpected signing method"), "Unexpected signing method: %v", t.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		return s.SecretKeyProvider.getKey().Public(), nil
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

	return claims, nil
}
