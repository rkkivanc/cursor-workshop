package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Commit represents a simplified Git commit returned to callers.
type Commit struct {
	SHA        string    `json:"sha"`
	Message    string    `json:"message"`
	AuthorName string    `json:"author_name"`
	Date       time.Time `json:"date"`
}

// FetchCommits fetches commits from the GitHub API for the given repository.
// It requests commits since 24 hours ago (UTC) and additionally filters
// results in Go to keep the time-window logic testable.
//
// The returned statusCode is the HTTP status that should be surfaced to
// the API caller (e.g. 200, 401, 404, 500).
func FetchCommits(ctx context.Context, githubToken, owner, repo string) ([]Commit, int, error) {
	if githubToken == "" || owner == "" || repo == "" {
		return nil, http.StatusBadRequest, errors.New("github_token, owner, and repo are required")
	}

	sinceTime := time.Now().Add(-24 * time.Hour).UTC()
	sinceParam := url.QueryEscape(sinceTime.Format(time.RFC3339))

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?since=%s", owner, repo, sinceParam)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("User-Agent", "standup-bot-backend")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusUnauthorized:
		return nil, http.StatusUnauthorized, errors.New("invalid GitHub token")
	case http.StatusNotFound:
		return nil, http.StatusNotFound, errors.New("repository not found")
	default:
		return nil, http.StatusInternalServerError, fmt.Errorf("unexpected status from GitHub: %s", resp.Status)
	}

	var ghCommits []githubCommitResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghCommits); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	filtered := make([]Commit, 0, len(ghCommits))
	for _, c := range ghCommits {
		commitTime := c.Commit.Author.Date
		// Ensure we only keep commits within the last 24 hours window.
		if commitTime.Before(sinceTime) {
			continue
		}

		filtered = append(filtered, Commit{
			SHA:        c.SHA,
			Message:    c.Commit.Message,
			AuthorName: c.Commit.Author.Name,
			Date:       commitTime,
		})
	}

	return filtered, http.StatusOK, nil
}

// githubCommitResponse matches the subset of fields we need from the
// GitHub commits API response.
type githubCommitResponse struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string    `json:"name"`
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

