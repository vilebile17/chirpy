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

func (config *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	coolFunc := func(response http.ResponseWriter, request *http.Request) {
		config.fileServerHits.Add(1)
		next.ServeHTTP(response, request)
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
	mux.Handle("/", cfg.middlewareMetricsInc(http.FileServer(http.Dir("./website/"))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", http.HandlerFunc(cfg.readServerHits))
	mux.HandleFunc("POST /admin/reset", http.HandlerFunc(cfg.resetHandler))
	mux.HandleFunc("POST /api/chirps", cfg.createChirpHandler)
	mux.HandleFunc("GET /api/chirps", cfg.getAllChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{ChirpID}", cfg.getChirpHandler)
	mux.HandleFunc("POST /api/users", cfg.registerUser)
	mux.HandleFunc("POST /api/login", cfg.loginHandler)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
