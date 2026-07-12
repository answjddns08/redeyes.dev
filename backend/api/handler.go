// Package api implements the HTTP API for the blog backend, including caching and data retrieval logic.
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type API struct {
	Logger *log.Logger
}

func New(logger *log.Logger) *API {
	api := &API{
		Logger: logger,
	}

	if err := InitializeCache(); err != nil {
		api.Logger.Printf("FATAL: Failed to initialize post cache: %v", err)
	}

	return api
}

func (api *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/posts", api.handlePosts)
	mux.HandleFunc("/api/posts/", api.handlePost)
	mux.HandleFunc("/api/tags", api.handleTags)
	mux.HandleFunc("/api/tags/", api.handleTags)
	mux.HandleFunc("/api/cache-status", api.handleCacheCheck)

	postsDir, err := postsDirPath()
	if err != nil {
		api.Logger.Printf("failed to resolve posts directory: %v", err)
		return
	}

	fs := http.FileServer(http.Dir(postsDir))
	mux.Handle("/api/posts/images/", http.StripPrefix("/api/posts/images/", fs))
}

// --- Handlers ---

// handleCacheCheck handles the /api/cache-status endpoint, returning the last modified time of the cache file.
func (api *API) handleCacheCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	serverTimestamp := time.Now().Format(time.RFC3339)

	response := map[string]any{
		"isValid":   true,
		"timestamp": serverTimestamp,
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	cacheData, err := readCacheFromJSON()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_read_cache"})
	}

	tags := cacheData.Tags

	WriteJSON(w, http.StatusOK, tags)
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code string) {
	WriteJSON(w, status, map[string]string{"error": code})
}
