package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vilebile17/chirpy/internal/auth"
	"github.com/vilebile17/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func respondWithError(response http.ResponseWriter, _ *http.Request, message string, err error, statusCode int) {
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
	response.WriteHeader(statusCode)
	response.Write(data)
}

func respondWithJSON(response http.ResponseWriter, request *http.Request, payload any, statusCode int) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		respondWithError(response, request, "Error interpreting the valid chirp json", err, http.StatusBadRequest)
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

func (config *apiConfig) createChirpHandler(response http.ResponseWriter, request *http.Request) {
	type IncomingJSON struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(request.Body)
	incomingjson := IncomingJSON{}
	err := decoder.Decode(&incomingjson)
	if err != nil {
		fmt.Println(err)
		respondWithError(response, request, "Something went wrong", err, http.StatusBadRequest)
		return
	}

	tokenString, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(response, request, "There was an error retrieving the JWT token", err, http.StatusUnauthorized)
		return
	}
	userID, err := auth.ValidateJWT(tokenString, config.secret)
	if err != nil {
		respondWithError(response, request, "There was an error Validating the JWT token", err, http.StatusUnauthorized)
		return
	}

	maxChirpLength := 140
	if len(incomingjson.Body) > maxChirpLength {
		respondWithError(response, request, "Chirp too long", nil, http.StatusBadRequest)
		return
	} else if len(incomingjson.Body) == 0 {
		respondWithError(response, request, "Chirp must be atleast one character long", nil, http.StatusBadRequest)
		return
	}

	sqlChirp, err := config.dbQueries.CreateChirp(request.Context(), database.CreateChirpParams{
		Body:   cleanProfanity(incomingjson.Body),
		UserID: userID,
	})
	if err != nil {
		respondWithError(response, request, "There was an error creating the chirp", err, http.StatusBadRequest)
		return
	}

	respondWithJSON(response, request, Chirp{
		ID:        sqlChirp.ID,
		CreatedAt: sqlChirp.CreatedAt,
		UpdatedAt: sqlChirp.UpdatedAt,
		Body:      sqlChirp.Body,
		UserID:    sqlChirp.UserID,
	}, http.StatusCreated)
}

func (config *apiConfig) getAllChirpsHandler(response http.ResponseWriter, request *http.Request) {
	sqlChirps, err := config.dbQueries.GetAllChirps(request.Context())
	if err != nil {
		respondWithError(response, request, "There was an error fetching the Chirps", err, http.StatusBadRequest)
		return
	}

	chirps := []Chirp{}
	for _, chirp := range sqlChirps {
		chirps = append(chirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(response, request, chirps, http.StatusOK)
}

func (config *apiConfig) getChirpHandler(response http.ResponseWriter, request *http.Request) {
	chirpID, err := uuid.Parse(request.PathValue("ChirpID"))
	if err != nil {
		respondWithError(response, request, "There was an error parsing that UUID", err, http.StatusBadRequest)
		return
	}

	chirp, err := config.dbQueries.GetOnechirp(request.Context(), chirpID)
	if err != nil {
		respondWithError(response, request, "Chirp not found", err, http.StatusNotFound)
		response.WriteHeader(404)
		return
	}

	respondWithJSON(response, request, Chirp{
		chirp.ID,
		chirp.CreatedAt,
		chirp.UpdatedAt,
		chirp.Body,
		chirp.UserID,
	}, http.StatusOK)
}
