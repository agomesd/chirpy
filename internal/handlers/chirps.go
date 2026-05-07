package handlers

import (
	"encoding/json"

	"net/http"

	"github.com/agomesd/chirpy/internal/auth"
	"github.com/agomesd/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserID    string `json:"user_id"`
}

type ChirpService struct {
	DB        *database.Queries
	JWTSecret string
}

func (s *ChirpService) HandleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")

	chirpDB, err := s.DB.GetChirp(r.Context(), uuid.MustParse(chirpID))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	chirp := Chirp{
		ID:        chirpDB.ID.String(),
		CreatedAt: chirpDB.CreatedAt.String(),
		UpdatedAt: chirpDB.UpdatedAt.String(),
		Body:      chirpDB.Body,
		UserID:    chirpDB.UserID.String(),
	}

	respondWithJSON(w, http.StatusOK, chirp)

}

func (s *ChirpService) HandleGetChirps(w http.ResponseWriter, r *http.Request) {

	chirpsDB, err := s.DB.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirps := []Chirp{}

	for _, chirp := range chirpsDB {

		chirps = append(chirps, Chirp{
			ID:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.String(),
			UpdatedAt: chirp.UpdatedAt.String(),
			Body:      chirp.Body,
			UserID:    chirp.UserID.String(),
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)

}

func (s *ChirpService) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	userUUID, err := auth.ValidateJWT(token, s.JWTSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	reqBody := requestBody{}
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp")
		return
	}

	validatedChirp := validateChirp(reqBody.Body)

	if !validatedChirp.isValid {
		respondWithError(w, http.StatusBadRequest, validatedChirp.msg)
		return
	}

	params := database.CreateChirpParams{
		Body:   validatedChirp.chirp,
		UserID: userUUID,
	}

	dbChirp, err := s.DB.CreateChirp(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID.String(),
		CreatedAt: dbChirp.CreatedAt.String(),
		UpdatedAt: dbChirp.UpdatedAt.String(),
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID.String(),
	}

	respondWithJSON(w, http.StatusCreated, chirp)

}
