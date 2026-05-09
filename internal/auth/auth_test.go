package auth

import (
	"encoding/hex"
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name:      "valid bearer token",
			headers:   http.Header{"Authorization": []string{"Bearer mytoken123"}},
			wantToken: "mytoken123",
			wantErr:   false,
		},
		{
			name:      "valid JWT token",
			headers:   http.Header{"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.abc"}},
			wantToken: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.abc",
			wantErr:   false,
		},
		{
			name:      "case insensitive scheme - lowercase bearer",
			headers:   http.Header{"Authorization": []string{"bearer mytoken123"}},
			wantToken: "mytoken123",
			wantErr:   false,
		},
		{
			name:      "case insensitive scheme - uppercase BEARER",
			headers:   http.Header{"Authorization": []string{"BEARER mytoken123"}},
			wantToken: "mytoken123",
			wantErr:   false,
		},
		{
			name:      "token with trailing whitespace",
			headers:   http.Header{"Authorization": []string{"Bearer mytoken123  "}},
			wantToken: "mytoken123",
			wantErr:   false,
		},
		{
			name:      "token with spaces preserved via SplitN",
			headers:   http.Header{"Authorization": []string{"Bearer token with spaces"}},
			wantToken: "token with spaces",
			wantErr:   false,
		},
		{
			name:    "missing authorization header",
			headers: http.Header{},
			wantErr: true,
		},
		{
			name:    "empty authorization header",
			headers: http.Header{"Authorization": []string{""}},
			wantErr: true,
		},
		{
			name:    "bearer with no token",
			headers: http.Header{"Authorization": []string{"Bearer"}},
			wantErr: true,
		},
		{
			name:    "bearer with only whitespace token",
			headers: http.Header{"Authorization": []string{"Bearer   "}},
			wantErr: true,
		},
		{
			name:    "wrong scheme - Basic auth",
			headers: http.Header{"Authorization": []string{"Basic dXNlcjpwYXNz"}},
			wantErr: true,
		},
		{
			name:    "no scheme - just token",
			headers: http.Header{"Authorization": []string{"mytoken123"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantToken {
				t.Errorf("GetBearerToken() = %q, want %q", got, tt.wantToken)
			}
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		wantKey string
		wantErr bool
	}{
		{
			name:    "valid api key",
			headers: http.Header{"Authorization": []string{"ApiKey myapikey123"}},
			wantKey: "myapikey123",
			wantErr: false,
		},
		{
			name:    "case insensitive scheme - lowercase apikey",
			headers: http.Header{"Authorization": []string{"apikey myapikey123"}},
			wantKey: "myapikey123",
			wantErr: false,
		},
		{
			name:    "case insensitive scheme - uppercase APIKEY",
			headers: http.Header{"Authorization": []string{"APIKEY myapikey123"}},
			wantKey: "myapikey123",
			wantErr: false,
		},
		{
			name:    "key with trailing whitespace",
			headers: http.Header{"Authorization": []string{"ApiKey myapikey123  "}},
			wantKey: "myapikey123",
			wantErr: false,
		},
		{
			name:    "key with spaces preserved via SplitN",
			headers: http.Header{"Authorization": []string{"ApiKey key with spaces"}},
			wantKey: "key with spaces",
			wantErr: false,
		},
		{
			name:    "missing authorization header",
			headers: http.Header{},
			wantErr: true,
		},
		{
			name:    "empty authorization header",
			headers: http.Header{"Authorization": []string{""}},
			wantErr: true,
		},
		{
			name:    "apikey with no key",
			headers: http.Header{"Authorization": []string{"ApiKey"}},
			wantErr: true,
		},
		{
			name:    "apikey with only whitespace",
			headers: http.Header{"Authorization": []string{"ApiKey   "}},
			wantErr: true,
		},
		{
			name:    "wrong scheme - Bearer",
			headers: http.Header{"Authorization": []string{"Bearer mytoken123"}},
			wantErr: true,
		},
		{
			name:    "wrong scheme - Basic auth",
			headers: http.Header{"Authorization": []string{"Basic dXNlcjpwYXNz"}},
			wantErr: true,
		},
		{
			name:    "no scheme - just key",
			headers: http.Header{"Authorization": []string{"myapikey123"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAPIKey(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantKey {
				t.Errorf("GetAPIKey() = %q, want %q", got, tt.wantKey)
			}
		})
	}
}

func TestHashPasswordAndCheckPassword(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" || hash == password {
		t.Fatalf("hash should be non-empty and not equal to plaintext password")
	}

	ok, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash() error = %v", err)
	}
	if !ok {
		t.Fatal("CheckPasswordHash() should succeed for matching password")
	}

	okWrong, err := CheckPasswordHash("wrong-password", hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash wrong pass error = %v", err)
	}
	if okWrong {
		t.Fatal("CheckPasswordHash() should reject wrong password")
	}
}

func TestMakeRefreshToken(t *testing.T) {
	a := MakeRefreshToken()
	b := MakeRefreshToken()

	if len(a) != 64 {
		t.Errorf("MakeRefreshToken() length = %d; want 64 (32-byte hex)", len(a))
	}
	if _, err := hex.DecodeString(a); err != nil {
		t.Errorf("MakeRefreshToken() not valid hex: %v", err)
	}
	if a == b {
		t.Fatal("two refresh tokens collided — extremely unlikely")
	}
}
