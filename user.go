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
	Password string `json:"password"`
	Email    string `json:"email"`
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
		respondWithError(response, request, "Something went wrong, required format: {'email':'EMAIL', 'password':'PASSWORD'}", err, http.StatusBadRequest)
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
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	decoder := json.NewDecoder(request.Body)
	incomingjson := IncomingJSON{}
	err := decoder.Decode(&incomingjson)
	if err != nil {
		respondWithError(response, request, "Something went wrong, required format: {'email':'EMAIL', 'password':'PASSWORD', 'expires_in_seconds':'EXPIRES_IN_SECONDS(optional)'}", err, http.StatusBadRequest)
		return
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

	jwt, err := auth.MakeJWT(user.ID, config.secret)
	if err != nil {
		respondWithError(response, request, "There was an error creating the JWT token", err, http.StatusBadRequest)
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(response, request, "There was an error creating the refresh token", err, http.StatusBadRequest)
		return
	}

	if _, err = config.dbQueries.CreateRefreshToken(request.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	}); err != nil {
		respondWithError(response, request, "There was an error adding the refresh token to the database", err, http.StatusBadRequest)
		return
	}

	respondWithJSON(response, request, User{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
		jwt,
		refreshToken,
	}, http.StatusOK)
}

func getJWTFromHeader(header http.Header, secret string) (string, uuid.UUID, error) {
	jwtToken, err := auth.GetBearerToken(header)
	if err != nil {
		return "", uuid.Nil, err
	}
	userID, err := auth.ValidateJWT(jwtToken, secret)
	if err != nil {
		return "", uuid.Nil, err
	}
	return jwtToken, userID, nil
}

func (config *apiConfig) updateDetailsHandler(response http.ResponseWriter, request *http.Request) {
	_, userID, err := getJWTFromHeader(request.Header, config.secret)
	if err != nil {
		respondWithError(response, request, "Something went wrong while validating the JWT", err, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(request.Body)
	incomingjson := IncomingJSON{}
	err = decoder.Decode(&incomingjson)
	if err != nil {
		respondWithError(response, request, "Something went wrong, required format: {'email':'EMAIL', 'password':'PASSWORD', 'expires_in_seconds':'EXPIRES_IN_SECONDS(optional)'}", err, http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(incomingjson.Password)
	if err != nil {
		respondWithError(response, request, "An error occured when hashing the password...", err, http.StatusBadRequest)
		return
	}

	user, err := config.dbQueries.UpdateUserEmailAndPassword(request.Context(), database.UpdateUserEmailAndPasswordParams{
		ID:             userID,
		Email:          incomingjson.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(response, request, "There was an error when updating the password...", err, http.StatusBadRequest)
		return
	}

	respondWithJSON(response, nil, struct {
		Email     string    `json:"email"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		Email:     user.Email,
		UpdatedAt: user.UpdatedAt,
	}, http.StatusOK)
}
