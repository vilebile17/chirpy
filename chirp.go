package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vilebile17/chirpy/internal/database"
)

func (config *apiConfig) chirpHandler(response http.ResponseWriter, request *http.Request) {
	type IncomingJSON struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(request.Body)
	incomingjson := IncomingJSON{}
	err := decoder.Decode(&incomingjson)
	if err != nil {
		fmt.Println(err)
		respondWithError(response, request, "Something went wrong", err)
		return
	}

	maxChirpLength := 140
	if len(incomingjson.Body) > maxChirpLength {
		respondWithError(response, request, "Chirp too long", nil)
		return
	} else if len(incomingjson.Body) == 0 {
		respondWithError(response, request, "Chirp must be atleast one character long", nil)
		return
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	sqlChirp, err := config.dbQueries.CreateChirp(request.Context(), database.CreateChirpParams{
		Body:   cleanProfanity(incomingjson.Body),
		UserID: incomingjson.UserID,
	})
	if err != nil {
		respondWithError(response, request, "There was an error creating the chirp", err)
		return
	}

	chirp := Chirp{
		sqlChirp.ID,
		sqlChirp.CreatedAt,
		sqlChirp.UpdatedAt,
		sqlChirp.Body,
		sqlChirp.UserID,
	}
	respondWithJSON(response, request, chirp, http.StatusCreated)
}

func respondWithError(response http.ResponseWriter, _ *http.Request, message string, err error) {
	fmt.Println(err)
	type ErrorJSON struct {
		Error string `json:"error"`
	}

	errorJSON := ErrorJSON{
		message,
	}

	data, err := json.Marshal(errorJSON)
	if err != nil {
		fmt.Printf("Error encoding error message into json: %s\n", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusBadRequest)
	response.Write(data)
}

func respondWithJSON(response http.ResponseWriter, request *http.Request, payload any, statusCode int) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		respondWithError(response, request, "Error interpreting the valid chirp json", err)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)
	response.Write(data)
}

func cleanProfanity(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		if strings.ToLower(word) == "kerfuffle" || word == "sharbert" || word == "fornax" {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
