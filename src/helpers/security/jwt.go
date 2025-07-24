package security

import (
	"errors"
	"fmt"
	"go-do-the-thing/src/helpers/constants"
	"go-do-the-thing/src/helpers/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtHandler struct {
	*SecretKeyProvider
	logger slog.Logger
}

func NewJwtHandler(env, working_dir string) JwtHandler {
	logger := slog.NewLogger("JWT Handler")

	// 4. Inject the loaded keys into the provider
	provider := newKeyProvider(env, working_dir)
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
			s.logger.Error(errors.New("unexpected signing method"), "Unexpected signing method: %v", t.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
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
