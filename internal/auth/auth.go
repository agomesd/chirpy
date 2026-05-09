package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password string, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})

	signed, err := newToken.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signed, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	newToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	if !newToken.Valid {
		return uuid.UUID{}, fmt.Errorf("Invalid token")
	}

	subjectString := claims.Subject

	parsedUUID, err := uuid.Parse(subjectString)
	if err != nil {
		return uuid.UUID{}, err
	}

	return parsedUUID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header not found")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("authorization header must use Bearer scheme")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("bearer token is empty")
	}

	return token, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)

	return hex.EncodeToString(key)
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header not found")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	if !strings.EqualFold(parts[0], "ApiKey") {
		return "", fmt.Errorf("authorization header must use ApiKey scheme")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("api key is empty")
	}

	return token, nil

}
