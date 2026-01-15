package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/vilebile17/chirpy/internal/auth"
)

func (config *apiConfig) refreshHandler(response http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(response, nil, "There was an error when getting the bearer token, please unsure that you have a header in the format 'Authorization: Bearer TOKENSTRING'", err, 400)
		return
	}

	refreshTokenObj, err := config.dbQueries.GetUserFromRefreshToken(request.Context(), token)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(response, nil, "That refresh Token wasn't found in the database", err, http.StatusUnauthorized)
		} else {
			respondWithError(response, nil, "There was an error checking the refresh token against the database", err, http.StatusBadRequest)
		}
		return
	}

	if refreshTokenObj.RevokedAt.Valid || time.Now().After(refreshTokenObj.ExpiresAt) {
		respondWithError(response, nil, "Oof, you refresh token is out of date or has been revoked", nil, http.StatusUnauthorized)
		return
	}

	jwt, err := auth.MakeJWT(refreshTokenObj.UserID, config.secret)
	if err != nil {
		respondWithError(response, nil, "There was an error creating the JWT access token", err, http.StatusBadRequest)
		return
	}

	respondWithJSON(response, request, struct {
		Token string `json:"token"`
	}{
		jwt,
	}, http.StatusOK)
}

func (config *apiConfig) revokeHandler(response http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(response, nil, "There was an error when getting the bearer token, please unsure that you have a header in the format 'Authorization: Bearer TOKENSTRING'", err, 400)
		return
	}

	if err = config.dbQueries.RevokeRefreshToken(request.Context(), token); err != nil {
		respondWithError(response, nil, "There was an error when trying to revoke the token", err, 400)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
