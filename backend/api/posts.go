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
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
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
		Tags     []string `json:"tags"`
		Content  string   `json:"content"`
		CoverImg string   `json:"coverImg,omitempty"`
	}{
		Folder:   post.Slug,
		Title:    post.Title,
		Date:     post.Date,
		Tags:     post.Tags,
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

// --- Obsidian-style image extension (![[file|width]]) ---
//
// Renders embeds like ![[image.png|350]] as
// <img src="/api/posts/images/..." width="..." /> during markdown parsing,
// instead of regex-replacing the raw text before parsing (which scanned the
// document twice). The parser, AST node and renderer are wired up via the
// wikiImageExtension below.

var wikiImageKind = ast.NewNodeKind("WikiImage")

var wikiImageRe = regexp.MustCompile(`^!\[\[([^|\]]+)(?:\|([0-9]+))?\]\]`)

// wikiImage is a custom inline AST node for ![[file|width]] embeds.
type wikiImage struct {
	ast.BaseInline
	Destination string
	Width       string
}

// Kind implements ast.Node.
func (n *wikiImage) Kind() ast.NodeKind { return wikiImageKind }

// Dump implements ast.Node.
func (n *wikiImage) Dump(source []byte, level int) {
	m := map[string]string{"Destination": n.Destination}
	if n.Width != "" {
		m["Width"] = n.Width
	}
	ast.DumpHelper(n, source, level, m, nil)
}

// wikiImageParser matches the wiki syntax at the current reader position and
// produces a wikiImage node, resolving image names via the upload's image map.
type wikiImageParser struct {
	postDir  string
	imageMap map[string]string
}

// Trigger implements parser.InlineParser.
func (p *wikiImageParser) Trigger() []byte { return []byte{'!'} }

// Parse implements parser.InlineParser.
func (p *wikiImageParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	m := wikiImageRe.FindSubmatch(line)
	if m == nil {
		// Not a wiki image; let the other parsers (e.g. the link parser for
		// standard ![alt](url) images) handle this position.
		return nil
	}

	fileName := string(m[1])
	width := ""
	if len(m[2]) > 0 {
		width = string(m[2])
	}

	storedName := p.postDir + "-" + fileName
	if mapped, ok := p.imageMap[fileName]; ok && mapped != "" {
		storedName = mapped
	}

	img := &wikiImage{
		Destination: path.Join("/api/posts/images", url.PathEscape(storedName)),
		Width:       width,
	}
	block.Advance(len(m[0]))
	return img
}

// wikiImageRenderer writes the wikiImage node as an <img> tag.
type wikiImageRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *wikiImageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(wikiImageKind, r.renderWikiImage)
}

func (r *wikiImageRenderer) renderWikiImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	img := node.(*wikiImage)

	_, _ = w.WriteString(`<img src="`)
	_, _ = w.Write(util.EscapeHTML([]byte(img.Destination)))
	_, _ = w.WriteString(`"`)
	if img.Width != "" {
		fmt.Printf(" width: %s\n", img.Width)
	}
	_, _ = w.WriteString(" />")
	return ast.WalkSkipChildren, nil
}

// wikiImageExtension wires the parser and renderer into goldmark.
type wikiImageExtension struct {
	postDir  string
	imageMap map[string]string
}

// Extend implements goldmark.Extender.
func (e *wikiImageExtension) Extend(md goldmark.Markdown) {
	// Priority 150 < the link parser's 200, so wiki images are matched before
	// the standard image/link syntax.
	md.Parser().AddOptions(
		parser.WithInlineParsers(util.Prioritized(&wikiImageParser{postDir: e.postDir, imageMap: e.imageMap}, 150)),
	)
	md.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(&wikiImageRenderer{}, 0)),
	)
}

func parseMarkdown(raw string, postDir string, imageMap map[string]string) (frontMatter, string, error) {
	// HTML 태그 입력을 허용하도록 goldmark 설정 (Unsafe 지정을 안 하면 <img> 태그가 렌더링되지 않고 생략됨)
	md := goldmark.New(
		goldmark.WithExtensions(meta.Meta, extension.Strikethrough, &wikiImageExtension{postDir: postDir, imageMap: imageMap}),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithHardWraps(),
		),
	)

	// 컨택스트 프론트매터 담는 변수
	context := parser.NewContext()
	var htmlBuf bytes.Buffer

	// convert markdown to HTML and extract front matter (and return it as a map)
	if err := md.Convert([]byte(raw), &htmlBuf, parser.WithContext(context)); err != nil {
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

func filterPosts(posts []frontMatter, search string) []frontMatter {
	search = strings.TrimSpace(search)
	if search == "" {
		return posts
	}

	rawTokens := strings.Split(search, ",") // split by comma to allow multiple search terms
	var targetTags []string
	var targetKeywords []string

	for _, token := range rawTokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if tag, ok := strings.CutPrefix(token, "#"); ok {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				targetTags = append(targetTags, tag)
			}
		} else {
			// if normal keyword, convert to lowercase for case-insensitive matching
			targetKeywords = append(targetKeywords, strings.ToLower(token))
		}
	}

	if len(targetTags) == 0 && len(targetKeywords) == 0 {
		return posts
	}

	filtered := []frontMatter{}
	seen := make(map[string]bool) // map by post slug to avoid duplicates

	for _, post := range posts {
		isMatch := false

		// 태그 검색 조건 (OR 결합: 검색 태그 중 하나라도 일치하면 통과)
		if len(targetTags) > 0 {
			for _, searchTag := range targetTags {
				for _, postTag := range post.Tags {
					if strings.EqualFold(postTag, searchTag) {
						isMatch = true
						break
					}
				}
				if isMatch {
					break
				}
			}
		}

		// 키워드 검색 조건 (OR 결합: 제목, 요약, 또는 태그 텍스트 부분 포함 여부)
		if !isMatch && len(targetKeywords) > 0 {
			titleLower := strings.ToLower(post.Title)
			summaryLower := strings.ToLower(post.Summary)

			for _, keyword := range targetKeywords {
				// 제목이나 요약에 포함되는지 확인
				if strings.Contains(titleLower, keyword) || strings.Contains(summaryLower, keyword) {
					isMatch = true
					break
				}

				// 태그 자체에 키워드가 부분 포함되는지 확인
				for _, postTag := range post.Tags {
					if strings.Contains(strings.ToLower(postTag), keyword) {
						isMatch = true
						break
					}
				}
				if isMatch {
					break
				}
			}
		}

		// 결과 수집 및 중복 제거
		if isMatch {
			if !seen[post.Slug] {
				seen[post.Slug] = true
				filtered = append(filtered, post)
			}
		}
	}

	return filtered
}
