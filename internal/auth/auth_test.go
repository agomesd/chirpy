package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	tests := []struct {
		name      string
		userID    uuid.UUID
		secret    string
		expiresIn time.Duration
		wantErr   bool
	}{
		{
			name:      "valid token creation",
			userID:    userID,
			secret:    secret,
			expiresIn: time.Hour,
			wantErr:   false,
		},
		{
			name:      "empty secret",
			userID:    userID,
			secret:    "",
			expiresIn: time.Hour,
			wantErr:   false, // HS256 allows empty secret, token still signs
		},
		{
			name:      "zero expiry",
			userID:    userID,
			secret:    secret,
			expiresIn: 0,
			wantErr:   false,
		},
		{
			name:      "negative expiry (already expired)",
			userID:    userID,
			secret:    secret,
			expiresIn: -time.Hour,
			wantErr:   false, // creation succeeds, validation would fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := MakeJWT(tt.userID, tt.secret, tt.expiresIn)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("MakeJWT() returned empty token")
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	validToken, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("failed to create token for tests: %v", err)
	}

	expiredToken, err := MakeJWT(userID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("failed to create expired token for tests: %v", err)
	}

	tests := []struct {
		name       string
		token      string
		secret     string
		wantUserID uuid.UUID
		wantErr    bool
	}{
		{
			name:       "valid token",
			token:      validToken,
			secret:     secret,
			wantUserID: userID,
			wantErr:    false,
		},
		{
			name:       "expired token",
			token:      expiredToken,
			secret:     secret,
			wantUserID: uuid.UUID{},
			wantErr:    true,
		},
		{
			name:       "wrong secret",
			token:      validToken,
			secret:     "wrong-secret",
			wantUserID: uuid.UUID{},
			wantErr:    true,
		},
		{
			name:       "malformed token",
			token:      "not.a.token",
			secret:     secret,
			wantUserID: uuid.UUID{},
			wantErr:    true,
		},
		{
			name:       "empty token",
			token:      "",
			secret:     secret,
			wantUserID: uuid.UUID{},
			wantErr:    true,
		},
		{
			name:       "empty secret",
			token:      validToken,
			secret:     "",
			wantUserID: uuid.UUID{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := ValidateJWT(tt.token, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantUserID {
				t.Errorf("ValidateJWT() = %v, want %v", gotID, tt.wantUserID)
			}
		})
	}
}

func TestMakeAndValidateJWT_RoundTrip(t *testing.T) {
	userID := uuid.New()
	secret := "round-trip-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	if gotID != userID {
		t.Errorf("round trip userID = %v, want %v", gotID, userID)
	}
}

func TestValidateJWT_DifferentUsers(t *testing.T) {
	secret := "test-secret"
	userID1 := uuid.New()
	userID2 := uuid.New()

	token1, _ := MakeJWT(userID1, secret, time.Hour)
	token2, _ := MakeJWT(userID2, secret, time.Hour)

	gotID1, err := ValidateJWT(token1, secret)
	if err != nil || gotID1 != userID1 {
		t.Errorf("token1 returned wrong user: got %v, want %v", gotID1, userID1)
	}

	gotID2, err := ValidateJWT(token2, secret)
	if err != nil || gotID2 != userID2 {
		t.Errorf("token2 returned wrong user: got %v, want %v", gotID2, userID2)
	}

	if gotID1 == gotID2 {
		t.Error("different users returned the same ID")
	}
}
