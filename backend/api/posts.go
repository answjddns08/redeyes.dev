package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// --- Data Structures ---
type frontMatter struct {
	Title   string   `yaml:"title" json:"title"`
	Date    string   `yaml:"date" json:"date"`
	Tag     []string `yaml:"tags" json:"tags"`
	Summary string   `yaml:"summary" json:"summary"`
	Cover   string   `yaml:"coverImg,omitempty" json:"coverImg,omitempty"`
	Folder  string   `json:"folder"`
}

// handlePosts handles the /api/posts endpoint, returning a list of posts with optional search filtering.
func (api *API) handlePosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	cacheData, err := readCacheFromJSON()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_read_cache"})
	}

	posts := make([]frontMatter, 0, len(cacheData.Frontmatter))
	for _, meta := range cacheData.Frontmatter {
		posts = append(posts, meta)
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	if search != "" {
		posts = filterPosts(posts, search)
	}

	WriteJSON(w, http.StatusOK, posts)
}

// handlePost handles the /api/posts/{folder} endpoint, returning the details of a specific post.
func (api *API) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

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

	cacheData, err := readCacheFromJSON()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_read_cache"})
	}

	payload := struct {
		Folder   string   `json:"folder"`
		Title    string   `json:"title"`
		Date     string   `json:"date"`
		Tag      []string `json:"tag"`
		Content  string   `json:"content"`
		CoverImg string   `json:"coverImg,omitempty"`
	}{
		Folder:   folder,
		Title:    cacheData.Frontmatter[folder].Title,
		Date:     cacheData.Frontmatter[folder].Date,
		Tag:      cacheData.Frontmatter[folder].Tag,
		Content:  cacheData.Details[folder],
		CoverImg: cacheData.Frontmatter[folder].Cover,
	}

	WriteJSON(w, http.StatusOK, payload)
}

func parsePostMarkdown(raw string, postDir string) (frontMatter, string, error) {
	// HTML 태그 입력을 허용하도록 goldmark 설정 (Unsafe 지정을 안 하면 <img> 태그가 렌더링되지 않고 생략됩니다)
	md := goldmark.New(
		goldmark.WithExtensions(meta.Meta, extension.Strikethrough),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithHardWraps(),
		),
	)

	imagePathProcessed := convertImagePath(raw, postDir)

	// 컨택스트 프론트매터 담는 변수
	context := parser.NewContext()
	var htmlBuf bytes.Buffer

	// Convert를 실행하면 내부적으로 프론트매터는 잘라내서 context에 저장하고,
	// 남은 본문은 HTML로 변환하여 htmlBuf에 담아줍니다.
	if err := md.Convert([]byte(imagePathProcessed), &htmlBuf, parser.WithContext(context)); err != nil {
		return frontMatter{}, "", err
	}

	metaData := meta.Get(context)
	var fm frontMatter

	// If no metadata present, return defaults with content
	if metaData == nil {
		fm.Folder = postDir
		return fm, htmlBuf.String(), nil
	}

	// Safe extraction of string fields
	if v, ok := metaData["title"].(string); ok {
		fm.Title = v
	}
	if v, ok := metaData["coverImg"].(string); ok {
		fm.Cover = v
	}
	if v, ok := metaData["date"].(string); ok {
		fm.Date = v
	}
	if v, ok := metaData["summary"].(string); ok {
		fm.Summary = v
	}
	fm.Folder = postDir

	// Tags can come in different shapes; normalize to []string
	fm.Tag = []string{}
	if t, exists := metaData["tag"]; exists && t != nil {
		switch tt := t.(type) {
		case []any:
			for _, it := range tt {
				if s, ok := it.(string); ok {
					fm.Tag = append(fm.Tag, s)
				}
			}
		case []string:
			fm.Tag = tt
		case string:
			// single tag as string
			fm.Tag = []string{tt}
		}
	}

	return fm, htmlBuf.String(), nil
}

func convertImagePath(markdown string, postDir string) string {
	imageDir := "/api/posts/images" // 이미지가 저장된 디렉토리 이름 (posts/ 폴더 내)

	re := regexp.MustCompile(`!\[\[([^|\]]+)(?:\|([0-9]+))?\]\]`)

	return re.ReplaceAllStringFunc(markdown, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		fileName := submatches[1] // 예: "image.png"
		width := submatches[2]    // 예: "350" (없으면 "")

		// 파일명 URL 인코딩 (공백은 %20 등으로 변환됨)
		escapedName := url.PathEscape(fileName)

		if width != "" {
			return fmt.Sprintf(`<img src="%s" width="%s" />`, path.Join(imageDir, postDir, escapedName), width)
		}
		return fmt.Sprintf(`<img src="%s" />`, path.Join(imageDir, postDir, escapedName))
	})
}

func filterPosts(posts []frontMatter, search string) []frontMatter {
	if search == "" {
		return posts
	}

	// 1. 태그 검색 모드 확인
	if tagStrings, isFound := strings.CutPrefix(search, "#"); isFound {
		tags := strings.Split(tagStrings, ",")

		filtered := []frontMatter{}
		for _, post := range posts {
			// 각 태그가 포스트의 태그 배열에 포함되는지 확인 (AND 조건)
			match := true
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if len(tag) == 0 {
					continue
				}
				found := false
				for _, postTag := range post.Tag {
					if strings.EqualFold(postTag, tag) {
						found = true
						break
					}
				}
				if !found {
					match = false
					break
				}
			}
			if match {
				filtered = append(filtered, post)
			}
		}
		return filtered
	}

	// 2. 키워드 검색 모드 (일반 검색)
	keyword := strings.ToLower(search)
	filtered := []frontMatter{}
	for _, post := range posts {
		// 제목, 요약, 태그에 키워드가 포함되는지 확인 (OR 조건)
		titleLower := strings.ToLower(post.Title)
		summaryLower := strings.ToLower(post.Summary)

		match := strings.Contains(titleLower, keyword) ||
			strings.Contains(summaryLower, keyword)

		// 태그 검색도 키워드로 수행할 수 있도록 처리
		if !match {
			for _, tag := range post.Tag {
				tagLower := strings.ToLower(tag)
				if strings.Contains(tagLower, keyword) {
					match = true
					break
				}
			}
		}

		if match {
			filtered = append(filtered, post)
		}
	}
	return filtered
}
