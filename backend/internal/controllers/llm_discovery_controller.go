package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"workshop-backend/internal/services"
)

type llmProvidersResponse struct {
	Providers []services.LLMProvider `json:"providers"`
}

type llmConnectRequest struct {
	Endpoint string `json:"endpoint"`
}

type llmConnectResponse struct {
	ActiveEndpoint string `json:"active_endpoint"`
}

// HandleLLMProviders handles GET /api/llm/providers.
// It returns the list of detected local AI providers.
func HandleLLMProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	providers := services.DiscoverLLMProviders(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(llmProvidersResponse{
		Providers: providers,
	})
}

// HandleLLMConnect handles POST /api/llm/connect.
// It accepts a JSON body with an "endpoint" string and sets it as the active LLM.
func HandleLLMConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	defer r.Body.Close()

	var req llmConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Endpoint == "" {
		writeError(w, http.StatusBadRequest, "endpoint is required")
		return
	}

	services.SetActiveLLMEndpoint(req.Endpoint)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(llmConnectResponse{
		ActiveEndpoint: req.Endpoint,
	})
}

// HandleLLMDownload handles POST /api/llm/download.
// For the workshop setup, the actual model download and serving are handled
// by the separate mlc-llm service (Docker container). This endpoint streams
// Server-Sent Events (SSE) to guide the user to start mlc-llm and then waits
// until the provider is detected as running.
func HandleLLMDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// In this workshop setup, the backend container does NOT include the
	// mlc_llm binary. Instead, the mlc-llm Docker service is responsible for
	// downloading and serving the model on port 11434.
	writeSSEProgress(
		w,
		flusher,
		"To download and run the recommended model, start the mlc-llm container (e.g. `docker compose --profile mlc up -d mlc-llm`). Waiting for mlc-llm on http://localhost:11434 ...",
	)

	// Poll provider discovery until mlc-llm is reported as running or the
	// request context is cancelled / times out.
	start := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			writeSSEError(w, flusher, "Download was cancelled or timed out while waiting for mlc-llm.")
			time.Sleep(200 * time.Millisecond)
			return
		case <-ticker.C:
			providers := services.DiscoverLLMProviders(r.Context())

			var mlc *services.LLMProvider
			for i := range providers {
				if providers[i].Name == "mlc-llm" {
					mlc = &providers[i]
					break
				}
			}

			if mlc == nil {
				writeSSEProgress(
					w,
					flusher,
					"Still waiting for mlc-llm to be discoverable...",
				)
				continue
			}

			if mlc.Status == "running" {
				elapsed := time.Since(start).Round(time.Second)
				writeSSEProgress(
					w,
					flusher,
					fmt.Sprintf("mlc-llm is running (detected after %s).", elapsed),
				)
				// Short window to let clients receive the final events.
				time.Sleep(200 * time.Millisecond)
				return
			}

			elapsed := time.Since(start).Round(time.Second)
			writeSSEProgress(
				w,
				flusher,
				fmt.Sprintf("mlc-llm not running yet (waited %s)...", elapsed),
			)
		}
	}
}

type sseProgress struct {
	Progress string `json:"progress"`
}

func writeSSEProgress(w http.ResponseWriter, flusher http.Flusher, msg string) {
	payload, err := json.Marshal(sseProgress{Progress: msg})
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", payload)
	flusher.Flush()
}

func writeSSEError(w http.ResponseWriter, flusher http.Flusher, msg string) {
	writeSSEProgress(w, flusher, msg)
}

