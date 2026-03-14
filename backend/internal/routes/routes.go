package routes

import (
	"net/http"

	"workshop-backend/internal/controllers"
)

// RegisterRoutes wires all HTTP routes to their handlers.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/commits", controllers.HandleCommits)
	mux.HandleFunc("/api/standup", controllers.HandleStandup)
}

