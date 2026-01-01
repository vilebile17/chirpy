package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"

	mux := http.NewServeMux()
	fileServerHandler := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", fileServerHandler))
	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	healthzHandler := func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte("OK"))
	}
	mux.HandleFunc("/healthz", healthzHandler)

	log.Fatal(server.ListenAndServe())
}
