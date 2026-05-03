package handlers

import (
	"encoding/json"
	"net/http"

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

func HandleGetChirpByID(q *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chirpID := r.PathValue("chirpID")

		chirpDB, err := q.GetChirp(r.Context(), uuid.MustParse(chirpID))
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
}

func HandleGetChirps(q *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		chirpsDB, err := q.GetChirps(r.Context())
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
}

func HandleCreateChirp(q *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type requestBody struct {
			Body   string `json:"body"`
			UserID string `json:"user_id"`
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
			UserID: uuid.MustParse(reqBody.UserID),
		}

		dbChirp, err := q.CreateChirp(r.Context(), params)
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
}
