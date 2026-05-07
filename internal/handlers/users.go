package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/agomesd/chirpy/internal/auth"
	"github.com/agomesd/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type RefreshToken struct {
	Token string `json:"token"`
}

type UserService struct {
	DB        *database.Queries
	JWTSecret string
}

func (s *UserService) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(accessToken, s.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	reqBody := requestBody{}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	decoder.Decode(&reqBody)

	newHashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	params := database.UpdateUserPasswordParams{
		HashedPassword: newHashedPassword,
		ID:             userID,
		Email:          reqBody.Email,
	}

	updatedUserDB, err := s.DB.UpdateUserPassword(r.Context(), params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := User{
		ID:        updatedUserDB.ID,
		CreatedAt: updatedUserDB.CreatedAt.String(),
		UpdatedAt: updatedUserDB.UpdatedAt.String(),
		Email:     updatedUserDB.Email,
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (s *UserService) HandleCreatUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	reqBody := requestBody{}
	err := decoder.Decode(&reqBody)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	hashedPassword, err := auth.HashPassword(reqBody.Password)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid authentication")
	}

	params := database.CreateUserParams{
		Email:          reqBody.Email,
		HashedPassword: hashedPassword,
	}

	dbUser, err := s.DB.CreateUser(r.Context(), params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid authentication")
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt.String(),
		UpdatedAt: dbUser.UpdatedAt.String(),
		Email:     dbUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (s *UserService) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	reqBody := requestBody{}
	decoder.Decode(&reqBody)

	userDB, err := s.DB.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	isValidPassword, err := auth.CheckPasswordHash(reqBody.Password, userDB.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if !isValidPassword {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	expiresIn := time.Hour

	token, err := auth.MakeJWT(userDB.ID, s.JWTSecret, expiresIn)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server failed to generate JWT")
		return
	}

	refreshToken := auth.MakeRefreshToken()
	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    userDB.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	}

	refreshTokenDB, err := s.DB.CreateRefreshToken(r.Context(), refreshTokenParams)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := User{
		ID:           userDB.ID,
		CreatedAt:    userDB.CreatedAt.String(),
		UpdatedAt:    userDB.UpdatedAt.String(),
		Email:        userDB.Email,
		Token:        token,
		RefreshToken: refreshTokenDB.Token,
	}

	respondWithJSON(w, http.StatusOK, user)

}

func (s *UserService) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	tokenDB, err := s.DB.GetRefreshToken(r.Context(), token)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if time.Now().After(tokenDB.ExpiresAt) || tokenDB.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	newToken, err := auth.MakeJWT(tokenDB.UserID, s.JWTSecret, time.Hour)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := RefreshToken{
		Token: newToken,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (s *UserService) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_, err = s.DB.RevokleRefreshToken(r.Context(), token)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}
