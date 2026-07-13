package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type sqliteStore struct {
	DB *sql.DB
	Mu sync.Mutex
}

type storedPost struct {
	frontMatter
	RawMarkdown string
	Content     string
	SourceHash  string
	SourceMTime int64
}

type rescanRecord struct {
	post         frontMatter
	rawMarkdown  string
	renderedHTML string
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

func postsDirPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "posts"), nil
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
			title TEXT NOT NULL,
			summary TEXT,
			tags TEXT,
			html_content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		);`}
	// slug: ID and url(ex: my-first-post)
	// created_at: timestamp when the post was created(default to current timestamp)

	for _, stmt := range stmts {
		if _, err := s.DB.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (s *sqliteStore) SavePost(post frontMatter, renderedHTML string) error {
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

	if err = persistPostTx(tx, post, rawMarkdown, renderedHTML); err != nil {
		return err
	}

	if err = pruneUnusedTagsTx(tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func persistPostTx(tx *sql.Tx, post frontMatter, rawMarkdown, renderedHTML string) error {
	hash := sha256.Sum256([]byte(rawMarkdown))
	sourceHash := hex.EncodeToString(hash[:])

	if _, err := tx.Exec(`INSERT INTO posts (folder, title, date, summary, cover_img, raw_markdown, rendered_html, source_hash, source_mtime, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(folder) DO UPDATE SET
			title = excluded.title,
			date = excluded.date,
			summary = excluded.summary,
			cover_img = excluded.cover_img,
			raw_markdown = excluded.raw_markdown,
			rendered_html = excluded.rendered_html,
			source_hash = excluded.source_hash,
			source_mtime = excluded.source_mtime,
			updated_at = excluded.updated_at`,
		post.Slug, post.Title, post.Date, post.Summary, post.Cover, rawMarkdown, renderedHTML, sourceHash, time.Now().UnixNano(), time.Now().UnixNano()); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM post_tags WHERE post_folder = ?`, post.Slug); err != nil {
		return err
	}

	seen := make(map[string]struct{})
	for _, tag := range post.Tag {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}

		if _, err := tx.Exec(`INSERT INTO tags (tag) VALUES (?) ON CONFLICT(tag) DO NOTHING`, tag); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO post_tags (post_folder, tag) VALUES (?, ?)`, post.Slug, tag); err != nil {
			return err
		}
	}

	return nil
}

func pruneUnusedTagsTx(tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM tags WHERE tag NOT IN (SELECT tag FROM post_tags)`)
	return err
}

func (s *sqliteStore) DeletePost(folder string) error {
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

	res, err := tx.Exec(`DELETE FROM posts WHERE folder = ?`, folder)
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

	if err = pruneUnusedTagsTx(tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *sqliteStore) RescanFromFilesystem() (processed int, skipped int, err error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	postsDir, err := postsDirPath()
	if err != nil {
		return 0, 0, err
	}

	entries, err := os.ReadDir(postsDir)
	if err != nil {
		return 0, 0, err
	}

	records := make([]rescanRecord, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folder := entry.Name()
		mdPath := filepath.Join(postsDir, folder, "index.md")
		raw, readErr := os.ReadFile(mdPath)
		if readErr != nil {
			skipped++
			continue
		}

		post, htmlBody, parseErr := parseMarkdown(string(raw), folder, nil)
		if parseErr != nil {
			skipped++
			continue
		}
		post.Slug = folder
		records = append(records, rescanRecord{
			post:         post,
			rawMarkdown:  string(raw),
			renderedHTML: htmlBody,
		})
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM post_tags`); err != nil {
		return 0, 0, err
	}
	if _, err = tx.Exec(`DELETE FROM posts`); err != nil {
		return 0, 0, err
	}
	if _, err = tx.Exec(`DELETE FROM tags`); err != nil {
		return 0, 0, err
	}

	for _, record := range records {
		if err = persistPostTx(tx, record.post, record.rawMarkdown, record.renderedHTML); err != nil {
			return 0, 0, err
		}
		processed++
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}

	return processed, skipped, nil
}

func (s *sqliteStore) ListPosts() ([]frontMatter, error) {
	rows, err := s.DB.Query(`SELECT folder, title, date, summary, cover_img FROM posts ORDER BY date DESC, folder ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]frontMatter, 0)
	for rows.Next() {
		var post frontMatter
		if err := rows.Scan(&post.Slug, &post.Title, &post.Date, &post.Summary, &post.Cover); err != nil {
			return nil, err
		}
		tags, err := s.loadTagsForFolder(post.Slug)
		if err != nil {
			return nil, err
		}
		post.Tag = tags
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *sqliteStore) GetPost(folder string) (storedPost, error) {
	row := s.DB.QueryRow(`SELECT folder, title, date, summary, cover_img, raw_markdown, rendered_html FROM posts WHERE folder = ?`, folder)

	var post storedPost
	if err := row.Scan(&post.Slug, &post.Title, &post.Date, &post.Summary, &post.Cover, &post.RawMarkdown, &post.Content); err != nil {
		return storedPost{}, err
	}

	tags, err := s.loadTagsForFolder(folder)
	if err != nil {
		return storedPost{}, err
	}
	post.Tag = tags

	return post, nil
}

func (s *sqliteStore) ListTags() ([]string, error) {
	rows, err := s.DB.Query(`SELECT tag FROM tags ORDER BY tag ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]string, 0)
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *sqliteStore) loadTagsForFolder(folder string) ([]string, error) {
	rows, err := s.DB.Query(`SELECT tag FROM post_tags WHERE post_folder = ? ORDER BY tag ASC`, folder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]string, 0)
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}
