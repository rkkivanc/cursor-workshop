package controllers

import (
	"encoding/json"
	"net/http"

	"workshop-backend/internal/services"
)

type commitsRequest struct {
	GitHubToken string `json:"github_token"`
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
}

type errorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// HandleCommits handles POST /api/commits.
// It accepts a JSON body with github_token, owner, and repo,
// delegates to the GitHub service, and returns either a commit array
// or a structured error.
func HandleCommits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	defer r.Body.Close()

	var req commitsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.GitHubToken == "" || req.Owner == "" || req.Repo == "" {
		writeError(w, http.StatusBadRequest, "github_token, owner, and repo are required")
		return
	}

	commits, statusCode, err := services.FetchCommits(r.Context(), req.GitHubToken, req.Owner, req.Repo)
	if err != nil {
		writeError(w, statusCode, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(commits)
}

func writeError(w http.ResponseWriter, statusCode int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Error: msg,
		Code:  statusCode,
	})
}

