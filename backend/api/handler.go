// Package api implements the HTTP API for the blog backend, including caching and data retrieval logic.
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type API struct {
	Logger     *log.Logger
	Store      *sqliteStore
	adminToken string
}

func New(logger *log.Logger) *API {
	store, err := openSQLiteStore()
	if err != nil {
		logger.Fatalf("FATAL: Failed to open SQLite store: %v", err)
	}

	api := &API{
		Logger:     logger,
		Store:      store,
		adminToken: strings.TrimSpace(os.Getenv("BLOG_ADMIN_TOKEN")),
	}

	return api
}

func (api *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/posts", api.handlePosts)
	mux.HandleFunc("/api/posts/", api.handlePost)
	mux.HandleFunc("/api/tags", api.handleTags)
	mux.HandleFunc("/api/tags/", api.handleTags)
	mux.HandleFunc("/api/upload", api.handleUpload)
	mux.HandleFunc("/api/rescan", api.handleRescan)
	mux.HandleFunc("/api/reindex", api.handleRescan)

	postsDir, err := postsDirPath()
	if err != nil {
		api.Logger.Printf("failed to resolve posts directory: %v", err)
		return
	}

	fs := http.FileServer(http.Dir(postsDir))
	mux.Handle("/api/posts/images/", http.StripPrefix("/api/posts/images/", fs))
}

func (api *API) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	tags, err := api.Store.ListTags()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_read_tags"})
		return
	}

	WriteJSON(w, http.StatusOK, tags)
}

func (api *API) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}
	if !api.requireAdminAuth(w, r) {
		return
	}

	err := r.ParseMultipartForm(128 << 20) // Limit to 128 MB
	if err != nil {
		fmt.Println("Error parsing multipart form:", err)
		http.Error(w, "failed_to_parse_form", http.StatusBadRequest)
		return
	}

	folderValues := r.MultipartForm.Value["folder"]
	if len(folderValues) == 0 || strings.TrimSpace(folderValues[0]) == "" {
		WriteError(w, http.StatusBadRequest, "missing_folder")
		return
	}
	folderName := strings.TrimSpace(folderValues[0])

	markdownValues := r.MultipartForm.Value["markdown"]
	if len(markdownValues) == 0 {
		WriteError(w, http.StatusBadRequest, "missing_markdown")
		return
	}
	mdFile := markdownValues[0]

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

	previewPost, _, err := parsePostMarkdown(mdFile, folderName, nil)
	if err != nil {
		api.Logger.Printf("failed to parse uploaded markdown: %v", err)
		WriteError(w, http.StatusBadRequest, "failed_to_parse_markdown")
		return
	}

	imageFiles := r.MultipartForm.File["images"]
	imageNameMap := make(map[string]string, len(imageFiles))
	usedNames := make(map[string]struct{}, len(imageFiles))
	for index, fileHeader := range imageFiles {
		storedName := uniqueImageName(previewPost.Title, folderName, fileHeader.Filename, index, usedNames)
		usedNames[storedName] = struct{}{}
		imageNameMap[fileHeader.Filename] = storedName
	}

	rewrittenMarkdown := rewriteImageReferences(mdFile, imageNameMap)
	finalPost, htmlBody, err := parsePostMarkdown(rewrittenMarkdown, folderName, imageNameMap)
	if err != nil {
		api.Logger.Printf("failed to render uploaded markdown: %v", err)
		WriteError(w, http.StatusBadRequest, "failed_to_render_markdown")
		return
	}
	finalPost.Folder = folderName

	// write markdown file
	err = os.WriteFile(path.Join(dirPath, "index.md"), []byte(rewrittenMarkdown), 0o644)
	if err != nil {
		api.Logger.Printf("failed to write markdown file: %v", err)
		WriteError(w, http.StatusInternalServerError, "failed_to_write_markdown")
		return
	}

	for _, fileHeader := range imageFiles {
		file, err := fileHeader.Open()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed_to_open_image")
			return
		}
		defer file.Close()

		storedName := imageNameMap[fileHeader.Filename]
		dst, err := os.Create(path.Join(dirPath, storedName))
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
	}

	if err := api.Store.SavePost(finalPost, rewrittenMarkdown, htmlBody); err != nil {
		api.Logger.Printf("failed to persist post %s to sqlite: %v", folderName, err)
		WriteError(w, http.StatusInternalServerError, "failed_to_save_post")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "file uploaded successfully"})
}

func (api *API) handleRescan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if !api.requireAdminAuth(w, r) {
		return
	}

	processed, skipped, err := api.Store.RescanFromFilesystem()
	if err != nil {
		api.Logger.Printf("failed to rescan filesystem into sqlite: %v", err)
		WriteError(w, http.StatusInternalServerError, "failed_to_rescan")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"message":   "rescan_completed",
		"processed": processed,
		"skipped":   skipped,
	})
}

func (api *API) handlePost(w http.ResponseWriter, r *http.Request) {
	folder := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	folder = strings.TrimSpace(folder)
	if folder == "" {
		api.handlePosts(w, r)
		return
	}

	if strings.Contains(folder, "/") {
		WriteError(w, http.StatusNotFound, "post_not_found")
		fmt.Printf("WARNING: Invalid post folder requested: %s\n", folder)
		return
	}

	switch r.Method {
	case http.MethodGet:
		post, err := api.Store.GetPost(folder)
		if err != nil {
			WriteError(w, http.StatusNotFound, "post_not_found")
			return
		}

		payload := struct {
			Folder   string   `json:"folder"`
			Title    string   `json:"title"`
			Date     string   `json:"date"`
			Tag      []string `json:"tag"`
			Content  string   `json:"content"`
			CoverImg string   `json:"coverImg,omitempty"`
		}{
			Folder:   post.Folder,
			Title:    post.Title,
			Date:     post.Date,
			Tag:      post.Tag,
			Content:  post.Content,
			CoverImg: post.Cover,
		}

		WriteJSON(w, http.StatusOK, payload)
	case http.MethodDelete:
		if !api.requireAdminAuth(w, r) {
			return
		}

		if err := api.Store.DeletePost(folder); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				WriteError(w, http.StatusNotFound, "post_not_found")
				return
			}
			api.Logger.Printf("failed to delete post from sqlite: %v", err)
			WriteError(w, http.StatusInternalServerError, "failed_to_delete_post")
			return
		}

		postsDir, err := postsDirPath()
		if err != nil {
			api.Logger.Printf("failed to resolve posts directory during delete: %v", err)
			WriteError(w, http.StatusInternalServerError, "failed_to_delete_files")
			return
		}

		if err := os.RemoveAll(filepath.Join(postsDir, folder)); err != nil {
			api.Logger.Printf("failed to remove post files for %s: %v", folder, err)
			WriteError(w, http.StatusInternalServerError, "failed_to_delete_files")
			return
		}

		processed, skipped, err := api.Store.RescanFromFilesystem()
		if err != nil {
			api.Logger.Printf("failed to rescan after deleting %s: %v", folder, err)
			WriteError(w, http.StatusInternalServerError, "failed_to_reindex")
			return
		}

		WriteJSON(w, http.StatusOK, map[string]any{
			"message":   "file deleted successfully",
			"processed": processed,
			"skipped":   skipped,
		})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
	}
}

func (api *API) requireAdminAuth(w http.ResponseWriter, r *http.Request) bool {
	if api.adminToken == "" {
		return true
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	if token == "" || token != api.adminToken {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}

	return true
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code string) {
	WriteJSON(w, status, map[string]string{"error": code})
}
