package main

import (
	"fmt"
	"net/http"
	"os"
)

func healthzHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("OK"))
}

func (cfg *apiConfig) readServerHits(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`,
		cfg.fileServerHits.Load())))
}

// accessAllowed: Returns true if access is allowed, otherwise false
func accessAllowed(resp http.ResponseWriter) bool {
	fmt.Println(os.Getenv("PLATFORM"))
	if plat := os.Getenv("PLATFORM"); plat == "dev" {
		return true
	}

	resp.WriteHeader(http.StatusForbidden)
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.Write([]byte("Unable to reset the system: Access Forbidden"))
	return false
}

func (cfg *apiConfig) resetHandler(resp http.ResponseWriter, req *http.Request) {
	if !accessAllowed(resp) {
		return
	}

	cfg.fileServerHits.Store(0)
	if err := cfg.dbQueries.ResetUsers(req.Context()); err != nil {
		resp.WriteHeader(400)
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.Write([]byte("There was an error clearing the users table"))
		fmt.Println(err)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.Write([]byte("Successfully removed all entries from the users table and reset the fileServerHits"))
}
