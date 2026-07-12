// Package api implements the HTTP API for the blog backend, including caching and data retrieval logic.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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
	mux.HandleFunc("/api/upload", api.handleUpload)

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

func (api *API) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(128 << 20) // Limit to 128 MB
	if err != nil {
		fmt.Println("Error parsing multipart form:", err)
		http.Error(w, "failed_to_parse_form", http.StatusBadRequest)
		return
	}

	folderName := r.MultipartForm.Value["folder"][0]
	fmt.Println("folderName:", folderName)

	mdFile := r.MultipartForm.Value["markdown"][0]
	fmt.Println("mdFile:", mdFile)

	// get posts directory path
	postsDir, err := postsDirPath()
	if err != nil {
		api.Logger.Printf("failed to resolve posts directory: %v", err)
		return
	}

	dirPath := path.Join(postsDir, folderName)

	// make directory if not exists
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		api.Logger.Printf("failed to create folder: %v", err)
		WriteError(w, http.StatusInternalServerError, "failed_to_create_folder")
		return
	}

	// write markdown file
	err = os.WriteFile(path.Join(dirPath, "index.md"), []byte(mdFile), 0o644)
	if err != nil {
		api.Logger.Printf("failed to write markdown file: %v", err)
		WriteError(w, http.StatusInternalServerError, "failed_to_write_markdown")
		return
	}

	imageFiles := r.MultipartForm.File["images"]

	for _, fileHeader := range imageFiles {
		file, err := fileHeader.Open()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed_to_open_image")
			return
		}
		defer file.Close()

		dst, err := os.Create(path.Join(dirPath, fileHeader.Filename))
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed_to_create_image_file")
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed_to_save_image")
			return
		}

		// TODO:나중에 이미지들 주소 바꿔치기 해야함
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "file uploaded successfully"})
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code string) {
	WriteJSON(w, status, map[string]string{"error": code})
}
