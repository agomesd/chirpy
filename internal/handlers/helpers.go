package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type errorResponse struct {
	Error string `json:"error"`
}

type validatedChirp struct {
	chirp   string
	isValid bool
	msg     string
}

var invalidWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respBody, err := json.Marshal(errorResponse{
		Error: msg,
	})

	if err != nil {
		log.Printf("Error marshalling JSON: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(respBody)

}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	respBody, err := json.Marshal(payload)

	if err != nil {
		log.Printf("Error marshalling JSON: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(respBody)
}

func validateChirp(chirp string) validatedChirp {
	if len(chirp) > 140 {
		return validatedChirp{
			chirp:   chirp,
			isValid: false,
			msg:     "Chirp is too long",
		}
	}

	sanitizedChirp := sanitizeWords(chirp)

	return validatedChirp{
		chirp:   sanitizedChirp,
		isValid: true,
		msg:     "",
	}

}

func sanitizeWords(s string) string {
	words := strings.Split(s, " ")

	for idx, word := range words {
		if _, ok := invalidWords[strings.ToLower(word)]; ok {
			words[idx] = "****"
		}
	}

	return strings.Join(words, " ")
}
