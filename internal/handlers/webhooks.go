package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/agomesd/chirpy/internal/auth"
	"github.com/agomesd/chirpy/internal/database"
	"github.com/google/uuid"
)

type WebhookService struct {
	DB     *database.Queries
	APIKey string
}

func (s *WebhookService) HandlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if apiKey != s.APIKey {
		respondWithError(w, http.StatusUnauthorized, "invalid api key")
		return
	}
	type requestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	reqBody := requestBody{}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	decoder.Decode(&reqBody)

	if reqBody.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	_, err = s.DB.UpgradeUserChirpyRed(r.Context(), uuid.MustParse(reqBody.Data.UserID))

	if err != nil {
		respondWithError(w, http.StatusNotFound, "user not found")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}
