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

type errorJSON struct {
	Error string `json:"error"`
}

func (cfg *apiConfig) chirpHandler(resp http.ResponseWriter, req *http.Request) {
	type incomingJSON struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	incomingjson := incomingJSON{}
	err := decoder.Decode(&incomingjson)
	if err != nil {
		fmt.Println(err)
		respondWithError(resp, req, "Something went wrong")
		return
	}

	if len(incomingjson.Body) > 140 {
		respondWithError(resp, req, "Chirp too long")
		return
	} else if len(incomingjson.Body) == 0 {
		respondWithError(resp, req, "Chirp must be atleast one character long")
		return
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	c, err := cfg.dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{
		cleanProfanity(incomingjson.Body),
		incomingjson.UserID,
	})

	chirp := Chirp{
		c.ID,
		c.CreatedAt,
		c.UpdatedAt,
		c.Body,
		c.UserID,
	}

	if err != nil {
		respondWithError(resp, req, "There was an error creating the chirp")
		return
	}
	respondWithJSON(resp, req, chirp, http.StatusCreated)
}

func respondWithError(resp http.ResponseWriter, _ *http.Request, message string) {
	e := errorJSON{
		message,
	}

	data, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("Error encoding error message into json: %s\n", err)
		resp.WriteHeader(500)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(400)
	resp.Write(data)
}

func respondWithJSON(resp http.ResponseWriter, req *http.Request, payload any, statusCode int) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		respondWithError(resp, req, "Error interpreting the valid chirp json ")
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(statusCode)
	resp.Write(data)
}

func cleanProfanity(s string) string {
	words := strings.Split(s, " ")
	lowerCaseWords := strings.Split(strings.ToLower(s), " ")
	for i, word := range lowerCaseWords {
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
