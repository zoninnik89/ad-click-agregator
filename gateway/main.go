package main

import (
	"log"
	"net/http"
)

const (
	httpAddress = ":8080"
)

func main() {
	mux := http.NewServeMux()
	handler := NewHandler()
	handler.registerRoutes(mux)

	log.Printf("Starting HTTP server at %s", httpAddress)

	if err := http.ListenAndServe(httpAddress, mux); err != nil {
		log.Fatal("Failed to start http server")
	}
}
