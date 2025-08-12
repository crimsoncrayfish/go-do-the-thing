package security

import (
	"errors"
	"fmt"
	"go-do-the-thing/src/helpers/constants"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtHandler struct {
	*SecretKeyProvider
	logger *slog.Logger
}

func NewJwtHandler(env string) JwtHandler {
	provider := newKeyProvider(env)
	logger := slog.With("source", "JwtHandler")
	return JwtHandler{
		SecretKeyProvider: provider,
		logger:            logger,
	}
}

func (s *JwtHandler) NewToken(userId, session string, expiry time.Time) (string, error) {
	claims := jwt.MapClaims{}
	claims["user_id"] = userId
	claims["expiry"] = expiry.Format(constants.DateTimeFormat)
	claims["session_id"] = session
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.getKey())
}

func (s *JwtHandler) ValidateToken(signedToken string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(signedToken, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			s.logger.Error("unexpected signing method", "error", errors.New("unexpected signing method"), "method", t.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return s.SecretKeyProvider.getKey().Public(), nil
	})
	if err != nil || token == nil {
		switch {

		case errors.Is(err, jwt.ErrTokenMalformed):
			s.logger.Error("Not a token", "error", err)
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			s.logger.Error("Invalid Signature", "error", err)
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			s.logger.Error("Timing issues", "error", err)
		case token == nil:
			s.logger.Error("what even happened. Nil token", "error", err)
		default:
			s.logger.Error("What even happened", "error", err)
		}
		return nil, err
	}

	return claims, nil
}
