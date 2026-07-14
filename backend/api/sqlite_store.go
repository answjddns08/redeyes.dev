package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

type sqliteStore struct {
	DB *sql.DB
	Mu sync.Mutex
}

type storedPost struct {
	frontMatter
	Content string
}

func openSQLiteStore() (*sqliteStore, error) {
	dbPath, err := sqliteDBPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	store := &sqliteStore{DB: db}
	if _, err := store.DB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func sqliteDBPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "cache", "blog.db"), nil
}

func imageDirPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "images"), nil
}

// migrate creates the necessary tables in the SQLite database if they do not already exist.
func (s *sqliteStore) migrate() error {
	stmts := []string{`
		CREATE TABLE IF NOT EXISTS posts (
			slug TEXT PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '',
			date TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '[]',
			cover_img TEXT NOT NULL DEFAULT '',
			summary TEXT NOT NULL DEFAULT '',
			raw_markdown TEXT NOT NULL DEFAULT '',
			html_content TEXT NOT NULL DEFAULT ''
		);`}
	// slug: ID and url(ex: my-first-post)
	// created_at: timestamp when the post was created(default to current timestamp)

	for _, stmt := range stmts {
		if _, err := s.DB.Exec(stmt); err != nil {
			return err
		}
	}

	if err := s.ensureColumn("posts", "raw_markdown", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumn("posts", "html_content", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumn("posts", "tags", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		return err
	}

	hasRenderedHTML, err := s.columnExists("posts", "rendered_html")
	if err != nil {
		return err
	}
	hasHTMLContent, err := s.columnExists("posts", "html_content")
	if err != nil {
		return err
	}
	if hasRenderedHTML && hasHTMLContent {
		if _, err := s.DB.Exec(`UPDATE posts SET html_content = rendered_html WHERE (html_content = '' OR html_content IS NULL) AND rendered_html IS NOT NULL AND rendered_html != ''`); err != nil {
			return err
		}
	}

	return nil
}

func (s *sqliteStore) ensureColumn(tableName, columnName, columnDefinition string) error {
	exists, err := s.columnExists(tableName, columnName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = s.DB.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, tableName, columnName, columnDefinition))
	return err
}

func (s *sqliteStore) columnExists(tableName, columnName string) (bool, error) {
	rows, err := s.DB.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, tableName))
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notnull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notnull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}

	return false, nil
}

func (s *sqliteStore) SavePost(post frontMatter, rawMarkdown, renderedHTML string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback() // rollback
		}
	}()

	if err = insertPostData(tx, post, rawMarkdown, renderedHTML); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func insertPostData(tx *sql.Tx, post frontMatter, rawMarkdown, renderedHTML string) error {
	tagsJSON, err := json.Marshal(post.Tags)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`INSERT INTO posts (slug, title, date, tags, cover_img, summary, raw_markdown, html_content) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(slug) DO UPDATE SET
		title = excluded.title,
		date = excluded.date,
		tags = excluded.tags,
		cover_img = excluded.cover_img,
		summary = excluded.summary,
		raw_markdown = excluded.raw_markdown,
		html_content = excluded.html_content
		`, post.Slug, post.Title, post.Date, string(tagsJSON), post.Cover, post.Summary, rawMarkdown, renderedHTML); err != nil {
		return err
	}

	return nil
}

func (s *sqliteStore) DeletePost(slug string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.Exec(`DELETE FROM posts WHERE slug = ?`, slug)
	if err != nil {
		return err
	}

	deletedRows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if deletedRows == 0 {
		return sql.ErrNoRows
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	// deleting associated images from the image directory
	imageDir, err := imageDirPath()
	if err != nil {
		return err
	}

	images, err := os.ReadDir(imageDir) // image lilst (ex: my-first-post_image1.png, my-first-post_image2.jpg)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var deleteErr error

	for _, img := range images {
		if strings.HasPrefix(img.Name(), slug+"_") {
			err := os.Remove(filepath.Join(imageDir, img.Name())) // delete image file
			if err != nil {
				fmt.Printf("failed to delete image file %s: %v\n", img.Name(), err)
				if deleteErr == nil {
					deleteErr = err
				}
			}
		}
	}

	return deleteErr
}

func (s *sqliteStore) ListPosts() ([]frontMatter, error) {
	rows, err := s.DB.Query(`SELECT slug, title, date, tags, cover_img, summary FROM posts ORDER BY date DESC, slug ASC`)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}()

	posts := make([]frontMatter, 0)
	for rows.Next() {
		var post frontMatter
		var tagsJSON string
		if err := rows.Scan(&post.Slug, &post.Title, &post.Date, &tagsJSON, &post.Cover, &post.Summary); err != nil {
			return nil, err
		}
		post.Tags = decodeTags(tagsJSON)
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *sqliteStore) GetPost(slug string) (storedPost, error) {
	row := s.DB.QueryRow(`SELECT slug, title, date, tags, cover_img, summary, html_content FROM posts WHERE slug = ?`, slug)

	var post storedPost
	var tagsJSON string
	if err := row.Scan(&post.Slug, &post.Title, &post.Date, &tagsJSON, &post.Cover, &post.Summary, &post.Content); err != nil {
		return storedPost{}, err
	}
	post.Tags = decodeTags(tagsJSON)

	return post, nil
}

// ListTags retrieves a list of unique tags from the posts in the SQLite database.
func (s *sqliteStore) ListTags() ([]string, error) {
	rows, err := s.DB.Query(`SELECT tags FROM posts`)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}()

	tags := make(map[string]struct{})
	for rows.Next() {
		var tagsJSON string
		if err := rows.Scan(&tagsJSON); err != nil {
			return nil, err
		}
		for _, tag := range decodeTags(tagsJSON) {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			tags[tag] = struct{}{}
		}
	}

	tagArr := make([]string, 0, len(tags))

	for tag := range tags {
		tagArr = append(tagArr, tag)
	}

	return tagArr, nil
}

func decodeTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}

	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err == nil {
		return tags
	}

	parts := strings.Split(raw, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(strings.Trim(part, "[]\""))
	}
	return parts
}
