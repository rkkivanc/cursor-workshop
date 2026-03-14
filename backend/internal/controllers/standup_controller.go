package controllers

import (
	"encoding/json"
	"net/http"

	"workshop-backend/internal/services"
)

// HandleStandup handles POST /api/standup.
// It accepts the same commit array schema returned by /api/commits and
// returns a structured standup summary.
func HandleStandup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	defer r.Body.Close()

	var commits []services.Commit
	if err := json.NewDecoder(r.Body).Decode(&commits); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	summary := services.GenerateStandupSummary(r.Context(), commits)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summary)
}

