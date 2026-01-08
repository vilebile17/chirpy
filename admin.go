package main

import (
	"fmt"
	"net/http"
	"os"
)

func healthzHandler(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("OK"))
}

func (config *apiConfig) readServerHits(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	HTMLpage := []byte{}
	HTMLpage = fmt.Appendf(HTMLpage, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`,
		config.fileServerHits.Load(),
	)
	response.Write(HTMLpage)
}

func accessAllowed(response http.ResponseWriter) bool {
	if platform := os.Getenv("PLATFORM"); platform == "dev" {
		return true
	}

	response.WriteHeader(http.StatusForbidden)
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.Write([]byte("Unable to reset the system: Access Forbidden"))
	return false
}

func (config *apiConfig) resetHandler(response http.ResponseWriter, request *http.Request) {
	if !accessAllowed(response) {
		return
	}

	config.fileServerHits.Store(0)
	if err := config.dbQueries.ResetUsers(request.Context()); err != nil {
		respondWithError(response, request, "There was an error reseting the users table", err, http.StatusBadRequest)
		return
	}

	response.WriteHeader(http.StatusOK)
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.Write([]byte("Successfully removed all entries from the users table and reset the fileServerHits"))
}
