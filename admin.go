package main

import (
	"fmt"
	"net/http"
)

func healthzHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("OK"))
}

func (cfg *apiConfig) resetServerHits(resp http.ResponseWriter, req *http.Request) {
	cfg.fileServerHits.Store(0)
	resp.Header().Add("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("Successfully reset fileServerHits\n"))
}

func (cfg *apiConfig) readServerHits(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "text/html; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`,
		cfg.fileServerHits.Load())))
}
