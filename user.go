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

func (cfg *apiConfig) registerUser(resp http.ResponseWriter, req *http.Request) {
	type incomingJSON struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	incomingjson := incomingJSON{}
	err := decoder.Decode(&incomingjson)
	if err != nil {
		fmt.Println(err)
		respondWithError(resp, req, "Something went wrong")
		return
	}

	u, err := cfg.dbQueries.CreateUser(req.Context(), incomingjson.Email)
	if err != nil {
		fmt.Println(err)
		respondWithError(resp, req, "An error occured when making the user...")
		return
	}

	user := User{u.ID, u.CreatedAt, u.UpdatedAt, u.Email}
	data, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		respondWithError(resp, req, "Error interpreting the user's json ")
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(201)
	resp.Write(data)
}
