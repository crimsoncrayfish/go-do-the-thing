package security

import (
	"crypto/rand"
	"crypto/rsa"
	"go-do-the-thing/src/helpers/constants"
	"go-do-the-thing/src/helpers/logger"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func createTestJwtHandler() *JwtHandler {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("Failed to generate test RSA key: " + err.Error())
	}

	provider := &SecretKeyProvider{
		keyList: map[string]*rsa.PrivateKey{"test": privateKey},
		kidList: []string{"test"},
	}
	testLogger := logger.NewLogger("TestJwtHandler")

	return &JwtHandler{
		SecretKeyProvider: provider,
		logger:            testLogger,
	}
}

func TestNewToken(t *testing.T) {
	handler := createTestJwtHandler()

	tests := []struct {
		name      string
		userId    string
		sessionId string
		expiry    time.Time
		wantErr   bool
	}{
		{
			name:      "valid token creation",
			userId:    "123",
			sessionId: "session-abc",
			expiry:    time.Now().Add(time.Hour),
			wantErr:   false,
		},
		{
			name:      "empty user id",
			userId:    "",
			sessionId: "session-abc",
			expiry:    time.Now().Add(time.Hour),
			wantErr:   false,
		},
		{
			name:      "past expiry time",
			userId:    "123",
			sessionId: "session-abc",
			expiry:    time.Now().Add(-time.Hour),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := handler.NewToken(tt.userId, tt.sessionId, tt.expiry)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewToken() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewToken() unexpected error: %v", err)
				return
			}

			if token == "" {
				t.Errorf("NewToken() returned empty token")
			}

			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
				return handler.getKey().Public(), nil
			})
			if err != nil {
				t.Errorf("Generated token is not parseable: %v", err)
			}

			if !parsedToken.Valid {
				t.Errorf("Generated token is not valid")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	handler := createTestJwtHandler()

	validExpiry := time.Now().Add(time.Hour)
	validToken, err := handler.NewToken("123", "session-abc", validExpiry)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	expiredExpiry := time.Now().Add(-time.Hour)
	expiredToken, err := handler.NewToken("123", "session-abc", expiredExpiry)
	if err != nil {
		t.Fatalf("Failed to create expired test token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		wantUserId  string
		wantSession string
	}{
		{
			name:        "valid token",
			token:       validToken,
			wantErr:     false,
			wantUserId:  "123",
			wantSession: "session-abc",
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:        "expired token",
			token:       expiredToken,
			wantErr:     false,
			wantUserId:  "123",
			wantSession: "session-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := handler.ValidateToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateToken() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateToken() unexpected error: %v", err)
				return
			}

			if claims == nil {
				t.Errorf("ValidateToken() returned nil claims")
				return
			}

			if userId, ok := claims["user_id"].(string); ok && userId != tt.wantUserId {
				t.Errorf("ValidateToken() user_id = %v, want %v", userId, tt.wantUserId)
			}

			if sessionId, ok := claims["session_id"].(string); ok && sessionId != tt.wantSession {
				t.Errorf("ValidateToken() session_id = %v, want %v", sessionId, tt.wantSession)
			}
		})
	}
}

func TestTokenRoundTrip(t *testing.T) {
	handler := createTestJwtHandler()

	testCases := []struct {
		userId    string
		sessionId string
		expiry    time.Time
	}{
		{"user123", "session456", time.Now().Add(time.Hour)},
		{"", "empty-user", time.Now().Add(time.Minute)},
		{"unicode-用户", "unicode-会话", time.Now().Add(24 * time.Hour)},
	}

	for _, tc := range testCases {
		t.Run("roundtrip", func(t *testing.T) {
			token, err := handler.NewToken(tc.userId, tc.sessionId, tc.expiry)
			if err != nil {
				t.Fatalf("NewToken() failed: %v", err)
			}

			claims, err := handler.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken() failed: %v", err)
			}

			if userId := claims["user_id"].(string); userId != tc.userId {
				t.Errorf("user_id mismatch: got %v, want %v", userId, tc.userId)
			}

			if sessionId := claims["session_id"].(string); sessionId != tc.sessionId {
				t.Errorf("session_id mismatch: got %v, want %v", sessionId, tc.sessionId)
			}

			if expiryStr := claims["expiry"].(string); expiryStr != tc.expiry.Format(constants.DateTimeFormat) {
				t.Errorf("expiry mismatch: got %v, want %v", expiryStr, tc.expiry.Format(constants.DateTimeFormat))
			}
		})
	}
}

func TestValidateTokenWithWrongKey(t *testing.T) {
	handler1 := createTestJwtHandler()
	handler2 := createTestJwtHandler()

	token, err := handler1.NewToken("123", "session", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	_, err = handler2.ValidateToken(token)
	if err == nil {
		t.Errorf("ValidateToken() should fail with wrong key")
	}
}
