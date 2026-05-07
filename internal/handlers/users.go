package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/agomesd/chirpy/internal/auth"
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

		dbUser, err := q.CreateUser(r.Context(), params)

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

}

func HandleLoginUser(q *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type requestBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		reqBody := requestBody{}
		decoder.Decode(&reqBody)

		userDB, err := q.GetUserByEmail(r.Context(), reqBody.Email)
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

		user := User{
			ID:        userDB.ID,
			CreatedAt: userDB.CreatedAt.String(),
			UpdatedAt: userDB.UpdatedAt.String(),
			Email:     userDB.Email,
		}

		respondWithJSON(w, http.StatusOK, user)

	}
}
