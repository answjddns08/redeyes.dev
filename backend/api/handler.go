// Package api provides the HTTP API handlers for the blog backend.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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
	mux.HandleFunc("/api/tags/", api.handleTags) // for backward compatibility
	mux.HandleFunc("/api/upload", api.handleUpload)
	mux.HandleFunc("/api/postdown/", api.handlePostDown)

	// 나중에 postDown으로 안쓰는 포스트들 지울 때 디렉토리에 있는 참조 안되는 이미지들도 지우도록하면 괜찮을 듯
	// 잠만 근데 slug(id)값 지우면 어케 바꾸지(바꿔야 하는 그 전 값을 모르는데)
	// 아 그럼 프론트에서 slug 값 저장하고 바뀌는지 확인하면 되려나
	//
	// 아 아니면 추가로 서버에 있는 블로그들 모달로 확인할 수 있도록 하고 지우게 하면 되겠네
	// 그럼 대충 문제 해결이겠네
	// 아 이미지도 지워야 하는데
	// 내가 일일이 하거나 서버에서 슬러그 목록 긁어서 그 단어 있는 이미지들 걸러내는 로직 넣어야겠네

	imageDir, err := imageDirPath()
	if err != nil {
		api.Logger.Printf("failed to resolve posts directory: %v", err)
		return
	}

	fs := http.FileServer(http.Dir(imageDir))
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
	if !api.requireAuth(w, r) {
		return
	}

	err := r.ParseMultipartForm(128 << 20) // Limit to 128 MB
	if err != nil {
		fmt.Println("Error parsing multipart form:", err)
		http.Error(w, "failed_to_parse_form", http.StatusBadRequest)
		return
	}

	// TODO: use NextPart to handle files not loading all into memory at once

	slugValues := r.MultipartForm.Value["slug"]
	if len(slugValues) == 0 || strings.TrimSpace(slugValues[0]) == "" {
		WriteError(w, http.StatusBadRequest, "missing_slug")
		return
	}
	slugName := strings.TrimSpace(slugValues[0])

	markdownValues := r.MultipartForm.Value["markdown"]
	if len(markdownValues) == 0 {
		WriteError(w, http.StatusBadRequest, "missing_markdown")
		return
	}
	mdFile := markdownValues[0]

	imageDir, err := imageDirPath()
	if err != nil {
		// make image directory if it doesn't exist
		api.Logger.Printf("failed to resolve image directory: %v\ngenerating new image directory", err)

		err = os.MkdirAll(imageDir, os.ModePerm)
		if err != nil {
			api.Logger.Printf("failed to create image folder: %v", err)
			WriteError(w, http.StatusInternalServerError, "failed_to_create_folder")
			return
		}
	}

	imageFiles := r.MultipartForm.File["images"]
	imageNameMap := make(map[string]string, len(imageFiles))
	for _, fileHeader := range imageFiles {
		//storedName := uniqueImageName(previewPost.Title, slugName, fileHeader.Filename, index, usedNames)
		storedName := slugName + "_" + fileHeader.Filename
		imageNameMap[fileHeader.Filename] = storedName
	}

	finalPost, htmlBody, err := parseMarkdown(mdFile, slugName, imageNameMap)
	if err != nil {
		api.Logger.Printf("failed to render uploaded markdown: %v", err)
		WriteError(w, http.StatusBadRequest, "failed_to_render_markdown")
		return
	}
	finalPost.Slug = slugName

	for _, fileHeader := range imageFiles {
		file, err := fileHeader.Open()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed_to_open_image")
			return
		}
		defer func() {
			err := file.Close()
			if err != nil {
				api.Logger.Printf("failed to close image file %s: %v", fileHeader.Filename, err)
			}
		}()

		storedName := imageNameMap[fileHeader.Filename]
		dst, err := os.Create(path.Join(imageDir, storedName))
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed_to_create_image_file")
			return
		}
		defer func() {
			err := dst.Close()
			if err != nil {
				api.Logger.Printf("failed to close destination image file %s: %v", storedName, err)
			}
		}()

		_, err = io.Copy(dst, file)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed_to_save_image")
			return
		}
	}

	if err := api.Store.SavePost(finalPost, htmlBody); err != nil {
		api.Logger.Printf("failed to persist post %s to sqlite: %v", slugName, err)
		WriteError(w, http.StatusInternalServerError, "failed_to_save_post")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "file uploaded successfully"})
}

func (api *API) handlePostDown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}

	if !api.requireAuth(w, r) {
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/postdown/")
	name = strings.TrimSpace(name)

	if name == "" {
		WriteError(w, http.StatusBadRequest, "missing_post_name")
		return
	}
}

func (api *API) requireAuth(w http.ResponseWriter, r *http.Request) bool {
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
