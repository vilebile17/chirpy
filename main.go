package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	dotenv "github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/vilebile17/chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	coolFunc := func(resp http.ResponseWriter, req *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(resp, req)
	}
	return http.HandlerFunc(coolFunc)
}

func main() {
	if err := dotenv.Load(); err != nil {
		log.Fatal(err)
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	cfg := apiConfig{dbQueries: database.New(db)}
	const port = "8080"

	mux := http.NewServeMux()
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", cfg.middlewareMetricsInc(fileServerHandler))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", http.HandlerFunc(cfg.readServerHits))
	mux.HandleFunc("POST /admin/reset", http.HandlerFunc(cfg.resetHandler))
	mux.HandleFunc("POST /api/chirps", cfg.chirpHandler)
	mux.HandleFunc("POST /api/users", cfg.registerUser)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
