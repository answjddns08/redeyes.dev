package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type API struct {
	Logger      *log.Logger
	pendingJobs atomic.Int64
	jobSeq      atomic.Uint64
}

func New(logger *log.Logger) *API {
	if logger == nil {
		logger = log.Default()
	}

	api := &API{
		Logger: logger,
	}

	if err := InitializeCache(); err != nil {
		api.Logger.Printf("FATAL: Failed to initialize post cache: %v", err)
	}

	return api
}

// --- Handlers ---

func (api *API) handleCacheCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	// *** CRITICAL CHANGE: Returning actual server timestamp ***
	// In a production system, this timestamp would be the last time the underlying data (DB/MD files) was successfully updated.
	serverTimestamp := time.Now().Format(time.RFC3339)

	response := map[string]interface{}{
		"isValid":   true, // Since we are using in-memory cache initialized on startup, we assume it's valid unless a more complex expiry rule is implemented.
		"timestamp": serverTimestamp,
	}

	WriteJSON(w, http.StatusOK, response)
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

	mux.Handle("/api/posts/images/", http.StripPrefix("/api/posts/images/", http.FileServer(http.Dir(postsDir))))
}

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code string) {
	WriteJSON(w, status, map[string]string{"error": code})
}
