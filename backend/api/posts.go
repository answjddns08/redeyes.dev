package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v3"
)

// --- Data Structures ---
type frontMatter struct {
	Title   string   `yaml:"title" json:"title"`
	Date    string   `yaml:"date" json:"date"`
	Tag     []string `yaml:"tag" json:"tag"`
	Summary string   `yaml:"summary" json:"summary"`
	Cover   string   `yaml:"coverImg" json:"coverImg,omitempty"`
	Folder  string   `json:"folder"`
}

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

func parsePostMarkdown(raw string) (frontMatter, string, error) {
	parts := strings.Split(raw, "---")
	if len(parts) < 3 {
		return frontMatter{}, "", fmt.Errorf("invalid front matter: at least 3 parts expected")
	}

	yamlPart := strings.TrimSpace(parts[1])
	markdownBody := strings.TrimSpace(strings.Join(parts[2:], "---"))

	var meta frontMatter
	if err := yaml.Unmarshal([]byte(yamlPart), &meta); err != nil {
		return frontMatter{}, "", fmt.Errorf("yaml unmarshal error: %w", err)
	}

	if meta.Tag == nil {
		meta.Tag = []string{}
	}

	// ========================================================
	// 💡 커스텀 로직: Obsidian 이미지 문법을 표준 Markdown으로 변환 (Pre-processing)
	// [[ImageName]] -> ![ImageName](images/ImageName.jpg)
	// ========================================================

	// 정규식: [[캡처그룹1]] 패턴을 찾습니다. (.*?)는 비탐욕적 매칭입니다.
	re := regexp.MustCompile(`\[\[(.*?)\]\]`)

	// 치환 함수: 찾은 그룹(ImageName)을 사용하여 표준 Markdown 형식으로 변환합니다.
	processedBody := re.ReplaceAllString(markdownBody, "![\\1](images/\\1.jpg)")

	// Markdown을 HTML로 파싱

	// 1. Parser 초기화 및 확장 설정
	// parser.NewWithExtensions를 사용하여 원하는 확장 기능을 활성화합니다.
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	// 2. AST 파싱
	doc := p.Parse([]byte(processedBody))

	// 3. HTML 렌더러 설정
	htmlFlags := html.CommonFlags | html.Smartypants
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// 4. AST를 HTML로 렌더링
	htmlBody := markdown.Render(doc, renderer)

	return meta, string(htmlBody), nil
}

func filterPosts(posts []frontMatter, search string) []frontMatter {
	if search == "" {
		return posts
	}

	// 1. 태그 검색 모드 확인
	if strings.HasPrefix(search, "#") {
		tagSearchString := strings.TrimPrefix(search, "#")
		tags := strings.Split(tagSearchString, ",")

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
