package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (config *apiConfig) registerUser(response http.ResponseWriter, request *http.Request) {
	type IncomingJSON struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(request.Body)
	incomingjson := IncomingJSON{}
	err := decoder.Decode(&incomingjson)
	if err != nil {
		fmt.Println(err)
		respondWithError(response, request, "Something went wrong", err)
		return
	}

	sqlUser, err := config.dbQueries.CreateUser(request.Context(), incomingjson.Email)
	if err != nil {
		fmt.Println(err)
		respondWithError(response, request, "An error occured when making the user...", err)
		return
	}

	user := User{sqlUser.ID, sqlUser.CreatedAt, sqlUser.UpdatedAt, sqlUser.Email}
	respondWithJSON(response, request, user, http.StatusCreated)
}
