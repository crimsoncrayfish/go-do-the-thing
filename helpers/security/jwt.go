package security

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtHandler struct {
	*SecretKeyProvider
}

func NewJwtHandler(keysLocation string) (JwtHandler, error) {
	keyProvider, err := newKeyProvider(keysLocation)
	if err != nil {
		return JwtHandler{}, err
	}
	return JwtHandler{
		SecretKeyProvider: keyProvider,
	}, nil
}

// API code to intercept all requests
func (s *JwtHandler) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")
		if token == "" {
			fmt.Println("No token provided. Should probably redirect to login screen")
		}
		isValid := s.validateToken(token)
		if !isValid {
			fmt.Printf("Failed authentication step with token %s", token)
		} else {
			fmt.Printf("Successfully authenticated with token: %s", token)
		}

		next.ServeHTTP(w, r)
	})
}

func (s *JwtHandler) newToken(claim jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claim)
	return token.SignedString(s.getKey())
}

func (s *JwtHandler) validateToken(signedToken string) bool {
	token, err := jwt.Parse(signedToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		return s.getKey().PublicKey, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			fmt.Println("Not a token")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			fmt.Println("Invalid Signature")
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			fmt.Println("Timing issues")
		case token == nil:
			fmt.Println("what... a nil token")
		default:
			fmt.Println("What even happened")
		}
		return false
	}

	return token.Valid
}

func (s *JwtHandler) validateTokenWithClaims(signedToken string, claims jwt.Claims) bool {
	token, err := jwt.ParseWithClaims(signedToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		return s.SecretKeyProvider.getKey().PublicKey, nil
	})
	if err != nil || token == nil {
		switch {

		case errors.Is(err, jwt.ErrTokenMalformed):
			// todo figure out wtf this even is
			fmt.Println("Not a token")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			// todo figure out wtf this even is
			fmt.Println("Invalid Signature")
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			// todo figure out wtf this even is
			fmt.Println("Timing issues")
		case token == nil:
			fmt.Println("what even happened. Nil token")
		default:
			// todo figure out wtf this even is
			fmt.Println("What even happened")
		}
		return false
	}

	return token.Valid
}

func CreateClaim(
	userId int,
	expiryTime time.Time,
) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiryTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        strconv.Itoa(userId),
		Issuer:    "go-do-the-thing",
		Audience:  jwt.ClaimStrings{"this"},
	}
}
