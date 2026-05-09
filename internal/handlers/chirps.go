package handlers

import (
	"encoding/json"

	"github.com/agomesd/chirpy/internal/auth"
	"github.com/agomesd/chirpy/internal/database"
	"github.com/google/uuid"
	"net/http"
	"sort"
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

func (s *ChirpService) HandleDeleteChirp(w http.ResponseWriter, r *http.Request) {
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
	chirpID := r.PathValue("chirpID")

	chirpDB, err := s.DB.GetChirp(r.Context(), uuid.MustParse(chirpID))

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if chirpDB.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You do not have permission to delete this chirp.")
		return
	}

	if err := s.DB.DeleteChirp(r.Context(), uuid.MustParse(chirpID)); err != nil {
		respondWithError(w, http.StatusNotFound, "Not found.")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

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
	chirps := []database.Chirp{}
	type Chirp struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
	}

	authorID := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort")

	var err error
	if authorID != "" {
		var id uuid.UUID
		id, err = uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid author_id")
			return
		}
		chirps, err = s.DB.GetChirpsByUserId(r.Context(), id)
	} else {
		chirps, err = s.DB.GetChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't fetch chirps")
		return
	}

	if sortParam == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}

	parsedChirps := []Chirp{}
	for _, chirp := range chirps {
		parsedChirps = append(parsedChirps, Chirp{
			ID:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.String(),
			UpdatedAt: chirp.UpdatedAt.String(),
			Body:      chirp.Body,
			UserID:    chirp.UserID.String(),
		})
	}

	respondWithJSON(w, http.StatusOK, parsedChirps)

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
