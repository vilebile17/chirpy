package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	coolFunc := func(resp http.ResponseWriter, req *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(resp, req)
	}
	return http.HandlerFunc(coolFunc)
}

func (cfg *apiConfig) readServerHits(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())))
}

func (cfg *apiConfig) resetServerHits(resp http.ResponseWriter, req *http.Request) {
	cfg.fileServerHits.Store(0)
	resp.Header().Add("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(http.StatusOK)

	if cfg.fileServerHits.Load() == 0 {
		resp.Write([]byte("Successfully reset fileServerHits"))
	} else {
		resp.Write([]byte("Err, that didn't seem to work..."))
	}
}

func main() {
	const port = "8080"
	cfg := apiConfig{}

	mux := http.NewServeMux()
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	healthzHandler := func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte("OK"))
	}

	mux.Handle("/app/", cfg.middlewareMetricsInc(fileServerHandler))
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/metrics", http.HandlerFunc(cfg.readServerHits))
	mux.HandleFunc("/reset", http.HandlerFunc(cfg.resetServerHits))

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
