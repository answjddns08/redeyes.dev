package api

import (
	"encoding/json"
	"errors"

	"log"
	"net/http"
	"sync/atomic"
)

type API struct {
	Logger      *log.Logger
	pendingJobs atomic.Int64
	jobSeq      atomic.Uint64
}

var errDownloadQueueFull = errors.New("download_queue_full")

func New(logger *log.Logger) *API {
	if logger == nil {
		logger = log.Default()
	}

	api := &API{
		Logger: logger,
	}

	return api
}

func (api *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", api.handleHealth)
	mux.HandleFunc("/api/posts", api.handlePosts)
	mux.HandleFunc("/api/posts/", api.handlePost)

	postsDir, err := postsDirPath()
	if err != nil {
		api.Logger.Printf("failed to resolve posts directory: %v", err)
		return
	}

	mux.Handle("/api/posts/images/", http.StripPrefix("/api/posts/images/", http.FileServer(http.Dir(postsDir))))
}
func (api *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"test":   "test",
	})
}

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code string) {
	WriteJSON(w, status, map[string]string{"error": code})
}
