package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vilebile17/chirpy/internal/auth"
	"github.com/vilebile17/chirpy/internal/database"
)

type IncomingJSON struct {
	Password         string `json:"password"`
	Email            string `json:"email"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

func (config *apiConfig) registerUser(response http.ResponseWriter, request *http.Request) {
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(request.Body)
	incomingjson := IncomingJSON{}
	err := decoder.Decode(&incomingjson)
	if err != nil {
		respondWithError(response, request, "Something went wrong", err, http.StatusBadRequest)
		return
	}

	email := incomingjson.Email
	hashedPassword, err := auth.HashPassword(incomingjson.Password)
	if err != nil {
		respondWithError(response, request, "There was an error hashing the password", err, http.StatusOK)
		return
	}

	sqlUser, err := config.dbQueries.CreateUser(request.Context(), database.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		fmt.Println(err)
		respondWithError(response, request, "An error occured when making the user...", err, http.StatusBadRequest)
		return
	}

	user := User{sqlUser.ID, sqlUser.CreatedAt, sqlUser.UpdatedAt, sqlUser.Email}
	respondWithJSON(response, request, user, http.StatusCreated)
}

func (config *apiConfig) loginHandler(response http.ResponseWriter, request *http.Request) {
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	decoder := json.NewDecoder(request.Body)
	incomingjson := IncomingJSON{}
	err := decoder.Decode(&incomingjson)
	if err != nil {
		respondWithError(response, request, "Something went wrong", err, http.StatusBadRequest)
		return
	}

	if incomingjson.ExpiresInSeconds == 0 || incomingjson.ExpiresInSeconds > 3600 {
		incomingjson.ExpiresInSeconds = 3600
	}

	user, err := config.dbQueries.SearchUsersByEmail(request.Context(), incomingjson.Email)
	if err != nil {
		respondWithError(response, request, "Email address wasn't found", err, http.StatusBadRequest)
		return
	}

	b, err := auth.CheckPasswordHash(incomingjson.Password, user.HashedPassword)
	if err != nil {
		respondWithError(response, request, "An error occured while the password hash was verified", err, http.StatusBadRequest)
		return
	}
	if !b {
		respondWithError(response, request, "Password doesn't match", nil, http.StatusUnauthorized)
		return
	}

	expiresIn := time.Duration(incomingjson.ExpiresInSeconds) * time.Second
	token, err := auth.MakeJWT(user.ID, config.secret, expiresIn)
	if err != nil {
		respondWithError(response, request, "There was an error creating the JWT token", err, 400)
	}

	respondWithJSON(response, request, User{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
		token,
	}, http.StatusOK)
}
