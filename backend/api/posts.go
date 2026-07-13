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
	Tags    []string `yaml:"tags" json:"tags"`
	Summary string   `yaml:"summary" json:"summary"`
	Cover   string   `yaml:"coverImg,omitempty" json:"coverImg,omitempty"`
	Slug    string   `json:"folder"`
}

// handlePosts handles the /api/posts endpoint, returning a list of posts with optional search filtering.
func (api *API) handlePosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	posts, err := api.Store.ListPosts()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_list_posts"})
		return
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	if search != "" {
		posts = filterPosts(posts, search)
	}

	WriteJSON(w, http.StatusOK, posts)
}

func (api *API) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	slug = strings.TrimSpace(slug)
	if slug == "" {
		api.handlePosts(w, r)
		return
	}

	if strings.Contains(slug, "/") {
		WriteError(w, http.StatusNotFound, "post_not_found")
		fmt.Printf("WARNING: Invalid post folder requested: %s\n", slug)
		return
	}

	post, err := api.Store.GetPost(slug)
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
		Folder:   post.Slug,
		Title:    post.Title,
		Date:     post.Date,
		Tag:      post.Tags,
		Content:  post.Content,
		CoverImg: post.Cover,
	}

	WriteJSON(w, http.StatusOK, payload)
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

func parseMarkdown(raw string, postDir string, imageMap map[string]string) (frontMatter, string, error) {
	// HTML 태그 입력을 허용하도록 goldmark 설정 (Unsafe 지정을 안 하면 <img> 태그가 렌더링되지 않고 생략됨)
	md := goldmark.New(
		goldmark.WithExtensions(meta.Meta, extension.Strikethrough),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithHardWraps(),
		),
	)

	// TODO: 나중에 goldmark 리졸버를 커스텀하도록 하는 편이 더 좋을 듯?
	// 이 방식은 파일을 2번 스캔해서 비효율적이니

	imagePathProcessed := convertImagePath(raw, postDir, imageMap)

	// 컨택스트 프론트매터 담는 변수
	context := parser.NewContext()
	var htmlBuf bytes.Buffer

	// convert markdown to HTML and extract front matter (and return it as a map)
	if err := md.Convert([]byte(imagePathProcessed), &htmlBuf, parser.WithContext(context)); err != nil {
		return frontMatter{}, "", err
	}

	metaData := meta.Get(context)
	var fm frontMatter

	// If no metadata present, return defaults with content
	if metaData == nil {
		fm.Slug = postDir
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
	fm.Slug = postDir

	// Tags can come in different shapes; normalize to []string
	fm.Tags = []string{}
	if t, exists := metaData["tags"]; exists && t != nil {
		switch tt := t.(type) {
		case []any:
			for _, it := range tt {
				if s, ok := it.(string); ok {
					fm.Tags = append(fm.Tags, s)
				}
			}
		case []string:
			fm.Tags = tt
		case string:
			// single tag as string
			fm.Tags = []string{tt}
		}
	}

	return fm, htmlBuf.String(), nil
}

func convertImagePath(markdown string, postTitle string, imageMap map[string]string) string {
	imageDir := "/api/posts/images" // 이미지가 저장된 디렉토리 이름 (posts/ 폴더 내)

	re := regexp.MustCompile(`!\[\[([^|\]]+)(?:\|([0-9]+))?\]\]`)

	return re.ReplaceAllStringFunc(markdown, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		fileName := submatches[1]                // ex: "image.png"
		width := submatches[2]                   // ex: "350" (if not exist"")
		storedName := postTitle + "-" + fileName // ex: postTitle-image.png

		if imageMap != nil && imageMap[fileName] != "" {
			storedName = imageMap[fileName]
		}

		// url encooding image name to handle special characters and spaces
		escapedName := url.PathEscape(storedName)

		if width != "" {
			return fmt.Sprintf(`<img src="%s" width="%s" />`, path.Join(imageDir, escapedName), width)
		}
		return fmt.Sprintf(`<img src="%s" />`, path.Join(imageDir, escapedName))
	})
}

func filterPosts(posts []frontMatter, search string) []frontMatter {
	if search == "" {
		return posts
	}

	// TODO : seperate tag search and keyword search(keyword affects title, summary)
	// and unite two search results (OR condition) and remove duplicates

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
				for _, postTag := range post.Tags {
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
			for _, tag := range post.Tags {
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
