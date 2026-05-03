package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/agomesd/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Email     string    `json:"email"`
}

func HandleCreateUser(q *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type requestBody struct {
			Email string `json:"email"`
		}
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		reqBody := requestBody{}
		err := decoder.Decode(&reqBody)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		dbUser, err := q.CreateUser(r.Context(), reqBody.Email)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid email")
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

}
