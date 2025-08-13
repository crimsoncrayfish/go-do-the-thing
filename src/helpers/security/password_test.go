package security

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSetPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "validPassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password within limit",
			password: strings.Repeat("a", 70),
			wantErr:  false,
		},
		{
			name:     "very long password - should fail",
			password: strings.Repeat("a", 73),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := SetPassword(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SetPassword() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("SetPassword() unexpected error: %v", err)
				return
			}

			if hash == "" {
				t.Errorf("SetPassword() returned empty hash")
			}

			if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password)); err != nil {
				t.Errorf("SetPassword() generated invalid hash: %v", err)
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	validPassword := "testPassword123"
	validHash, err := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate test hash: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: validPassword,
			hash:     string(validHash),
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongPassword",
			hash:     string(validHash),
			want:     false,
		},
		{
			name:     "empty password with valid hash",
			password: "",
			hash:     string(validHash),
			want:     false,
		},
		{
			name:     "valid password with empty hash",
			password: validPassword,
			hash:     "",
			want:     false,
		},
		{
			name:     "invalid hash format",
			password: validPassword,
			hash:     "invalid-hash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPassword(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordRoundTrip(t *testing.T) {
	passwords := []string{
		"simple",
		"Complex123!@#",
		"unicode-测试-🔒",
		"spaces in password",
	}

	for _, password := range passwords {
		t.Run("roundtrip_"+password, func(t *testing.T) {
			hash, err := SetPassword(password)
			if err != nil {
				t.Fatalf("SetPassword() failed: %v", err)
			}

			if !CheckPassword(password, hash) {
				t.Errorf("CheckPassword() failed for password that was just hashed")
			}

			if CheckPassword(password+"wrong", hash) {
				t.Errorf("CheckPassword() should fail for incorrect password")
			}
		})
	}
}
