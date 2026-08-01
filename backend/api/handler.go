// Package api provides the HTTP API handlers for the blog backend.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	fmt.Printf("Admin token: %s\n", api.adminToken) // For debugging purposes, remove in production

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

func (api *API) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}
	if !api.requireAuth(r) {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Stream multipart parts one by one (via NextPart) instead of loading the
	// whole form into memory with ParseMultipartForm. Image files are written
	// straight to temp files on disk, never buffered in memory. The request
	// body is still hard-capped at 128 MB.
	r.Body = http.MaxBytesReader(w, r.Body, 128<<20)

	mr, err := r.MultipartReader()
	if err != nil {
		api.Logger.Printf("failed to parse multipart form: %v", err)
		WriteError(w, http.StatusBadRequest, "failed_to_parse_form")
		return
	}

	imageDir, err := imageDirPath()
	if err != nil {
		api.Logger.Printf("failed to resolve image directory: %v", err)
		WriteError(w, http.StatusInternalServerError, "failed_to_resolve_image_dir")
		return
	}
	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		api.Logger.Printf("failed to create image folder: %v", err)
		WriteError(w, http.StatusInternalServerError, "failed_to_create_folder")
		return
	}

	// multpartData structure
	// - slug: string
	// - markdown: file
	// - images: file[]

	// Image parts are buffered into temp files first so the images can be
	// streamed to disk in any order, then renamed to their final names once
	// the slug and markdown fields have all been read.
	type pendingImage struct {
		origName   string
		tmpPath    string
		storedName string
	}
	var pendingImages []pendingImage
	cleanupPending := func() {
		for _, img := range pendingImages {
			if err := os.Remove(img.tmpPath); err != nil && !os.IsNotExist(err) {
				api.Logger.Printf("failed to clean up temp image %s: %v", img.tmpPath, err)
			}
		}
	}
	defer cleanupPending()

	var slugName string
	var mdFile string
	sawMarkdown := false

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			api.Logger.Printf("failed to read multipart part: %v", err)
			WriteError(w, http.StatusBadRequest, "failed_to_parse_form")
			return
		}

		switch part.FormName() {
		case "slug":
			data, readErr := io.ReadAll(part)
			_ = part.Close()
			if readErr != nil {
				api.Logger.Printf("failed to read slug field: %v", readErr)
				WriteError(w, http.StatusBadRequest, "failed_to_parse_form")
				return
			}
			slugName = strings.TrimSpace(string(data))

		case "markdown":
			data, readErr := io.ReadAll(part)
			_ = part.Close()
			if readErr != nil {
				api.Logger.Printf("failed to read markdown field: %v", readErr)
				WriteError(w, http.StatusBadRequest, "failed_to_parse_form")
				return
			}
			sawMarkdown = true
			mdFile = string(data)

		case "images":
			fileName := part.FileName()
			if fileName == "" {
				_, _ = io.Copy(io.Discard, part)
				_ = part.Close()
				continue
			}

			tmpFile, createErr := os.CreateTemp(imageDir, ".upload-*")
			if createErr != nil {
				_, _ = io.Copy(io.Discard, part)
				_ = part.Close()
				api.Logger.Printf("failed to create temp image file: %v", createErr)
				WriteError(w, http.StatusInternalServerError, "failed_to_create_image_file")
				return
			}
			_, copyErr := io.Copy(tmpFile, part)
			_ = part.Close()
			closeErr := tmpFile.Close()
			if copyErr != nil || closeErr != nil {
				_ = os.Remove(tmpFile.Name())
				api.Logger.Printf("failed to stream image %s: %v", fileName, copyErr)
				WriteError(w, http.StatusInternalServerError, "failed_to_save_image")
				return
			}
			pendingImages = append(pendingImages, pendingImage{origName: fileName, tmpPath: tmpFile.Name()})

		default:
			_, _ = io.Copy(io.Discard, part)
			_ = part.Close()
		}
	}

	if slugName == "" {
		WriteError(w, http.StatusBadRequest, "missing_slug")
		return
	}
	if !sawMarkdown {
		WriteError(w, http.StatusBadRequest, "missing_markdown")
		return
	}

	imageNameMap := make(map[string]string, len(pendingImages))
	for i := range pendingImages {
		storedName := fmt.Sprintf("%s_%02d_%s", slugName, i+1, filepath.Base(pendingImages[i].origName))
		pendingImages[i].storedName = storedName
		imageNameMap[pendingImages[i].origName] = storedName
	}

	finalPost, htmlBody, err := parseMarkdown(mdFile, slugName, imageNameMap)
	if err != nil {
		api.Logger.Printf("failed to render uploaded markdown: %v", err)
		WriteError(w, http.StatusBadRequest, "failed_to_render_markdown")
		return
	}
	finalPost.Slug = slugName

	// Move streamed temp files to their final names, then persist the post.
	for _, img := range pendingImages {
		if err := os.Rename(img.tmpPath, filepath.Join(imageDir, img.storedName)); err != nil {
			api.Logger.Printf("failed to finalize image %s: %v", img.origName, err)
			WriteError(w, http.StatusInternalServerError, "failed_to_save_image")
			return
		}
	}

	if err := api.Store.SavePost(finalPost, mdFile, htmlBody); err != nil {
		api.Logger.Printf("failed to persist post %s to sqlite: %v", slugName, err)
		WriteError(w, http.StatusInternalServerError, "failed_to_save_post")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "file uploaded successfully"})
}

func (api *API) handlePostDown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}

	// only need param(slug) from URL, no body needed

	if !api.requireAuth(r) {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/postdown/")
	name = strings.TrimSpace(name)

	if name == "" {
		WriteError(w, http.StatusBadRequest, "missing_post_name")
		return
	}

	err := api.Store.DeletePost(name)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed_to_delete_post")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "post deleted successfully"})
}

func (api *API) requireAuth(r *http.Request) bool {
	if api.adminToken == "" {
		return false
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return false
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	if token == "" || token != api.adminToken {
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
