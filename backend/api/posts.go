package api

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type frontMatter struct {
	Title   string      `yaml:"title"`
	Date    string      `yaml:"date"`
	Tag     stringSlice `yaml:"tag"`
	Summary string      `yaml:"summary"`
	Cover   *string     `yaml:"coverImg"`
}

type stringSlice []string

func (s *stringSlice) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		if strings.TrimSpace(value.Value) == "" {
			*s = []string{}
			return nil
		}
		*s = []string{value.Value}
		return nil
	case yaml.SequenceNode:
		var out []string
		if err := value.Decode(&out); err != nil {
			return err
		}
		*s = out
		return nil
	default:
		return fmt.Errorf("unsupported tag format")
	}
}

type postListItem struct {
	Folder   string   `json:"folder"`
	Title    string   `json:"title"`
	Date     string   `json:"date"`
	Tag      []string `json:"tag"`
	Summary  string   `json:"summary"`
	CoverImg *string  `json:"coverImg"`
}

type postDetail struct {
	Folder  string   `json:"folder"`
	Title   string   `json:"title"`
	Date    string   `json:"date"`
	Tag     []string `json:"tag"`
	Content string   `json:"content"`
}

func (api *API) handlePosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	posts, err := readAllPostListItems()
	if err != nil {
		api.Logger.Printf("failed to read posts: %v", err)
		WriteError(w, http.StatusInternalServerError, "failed_to_read_posts")
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
		return
	}

	post, err := readSinglePost(folder)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			WriteError(w, http.StatusNotFound, "post_not_found")
			return
		}
		api.Logger.Printf("failed to read post %s: %v", folder, err)
		WriteError(w, http.StatusInternalServerError, "failed_to_read_post")
		return
	}

	WriteJSON(w, http.StatusOK, post)
}

func filterPosts(posts []postListItem, rawSearch string) []postListItem {
	terms := strings.Split(rawSearch, ",")
	cleanTerms := make([]string, 0, len(terms))
	for _, term := range terms {
		trimmed := strings.TrimSpace(term)
		if trimmed == "" {
			continue
		}
		cleanTerms = append(cleanTerms, trimmed)
	}

	if len(cleanTerms) == 0 {
		return posts
	}

	filtered := make([]postListItem, 0, len(posts))
	for _, post := range posts {
		if matchesAnyTerm(post, cleanTerms) {
			filtered = append(filtered, post)
		}
	}
	return filtered
}

func matchesAnyTerm(post postListItem, terms []string) bool {
	for _, term := range terms {
		if strings.HasPrefix(term, "#") {
			needle := strings.ToLower(strings.TrimPrefix(term, "#"))
			for _, tag := range post.Tag {
				if strings.Contains(strings.ToLower(tag), needle) {
					return true
				}
			}
			continue
		}

		if strings.Contains(strings.ToLower(post.Title), strings.ToLower(term)) {
			return true
		}
	}
	return false
}

func readAllPostListItems() ([]postListItem, error) {
	root, err := postsDirPath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	result := make([]postListItem, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		item, err := readListItem(root, entry.Name())
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

func readSinglePost(folder string) (postDetail, error) {
	root, err := postsDirPath()
	if err != nil {
		return postDetail{}, err
	}

	path := filepath.Join(root, folder, "index.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return postDetail{}, err
	}

	meta, body, err := parsePostMarkdown(string(content))
	if err != nil {
		return postDetail{}, err
	}

	return postDetail{
		Folder:  folder,
		Title:   meta.Title,
		Date:    meta.Date,
		Tag:     []string(meta.Tag),
		Content: body,
	}, nil
}

func readListItem(root, folder string) (postListItem, error) {
	path := filepath.Join(root, folder, "index.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return postListItem{}, err
	}

	meta, _, err := parsePostMarkdown(string(content))
	if err != nil {
		return postListItem{}, err
	}

	return postListItem{
		Folder:   folder,
		Title:    meta.Title,
		Date:     meta.Date,
		Tag:      []string(meta.Tag),
		Summary:  meta.Summary,
		CoverImg: meta.Cover,
	}, nil
}

func parsePostMarkdown(raw string) (frontMatter, string, error) {
	parts := strings.Split(raw, "---")
	if len(parts) < 3 {
		return frontMatter{}, "", errors.New("invalid front matter")
	}

	yamlPart := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(strings.Join(parts[2:], "---"))

	var meta frontMatter
	if err := yaml.Unmarshal([]byte(yamlPart), &meta); err != nil {
		return frontMatter{}, "", err
	}
	if meta.Tag == nil {
		meta.Tag = []string{}
	}

	return meta, body, nil
}

func postsDirPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "posts"), nil
}
