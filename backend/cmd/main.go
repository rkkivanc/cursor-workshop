package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"workshop-backend/internal/routes"
)

func main() {
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Println("Starting HTTP server on :8080")

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

