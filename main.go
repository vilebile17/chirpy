package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"

	m := http.NewServeMux()
	h := http.FileServer(http.Dir("."))
	m.Handle("/", h)
	server := http.Server{
		Addr:    ":" + port,
		Handler: m,
	}

	log.Fatal(server.ListenAndServe())
}
