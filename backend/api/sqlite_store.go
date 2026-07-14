package api

import (
	"database/sql"
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
			title TEXT NOT NULL,
			date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			tags TEXT,
			cover_img TEXT,
			summary TEXT,
			html_content TEXT NOT NULL,
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
			_ = tx.Rollback() // rollback
		}
	}()

	if err = insertPostData(tx, post, renderedHTML); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func insertPostData(tx *sql.Tx, post frontMatter, renderedHTML string) error {
	if _, err := tx.Exec(`INSERT INTO posts (slug, title, date, tags, summary, cover_img, rendered_html) VALUES (?, ?, ?, ?, ?, ?, ?) 
	ON CONFLICT(slug) DO UPDATE SET
		slug = excluded.slug,
		title = excluded.title,
		date = excluded.date,
		tags = excluded.tags,
		cover_img = excluded.cover_img,
		summary = excluded.summary,
		html_content = excluded.html_content
		`, post.Slug, post.Title, post.Date, post.Tags, post.Cover, post.Summary, renderedHTML); err != nil {
		return err
	}

	return nil
}

func (s *sqliteStore) DeletePost(slug string) error {
	s.Mu.Lock()

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
	s.Mu.Unlock()

	// deleting associated images from the image directory
	imageDir, err := imageDirPath()
	if err != nil {
		fmt.Printf("failed to resolve image directory: %v\n", err)
		return nil
	}

	images, err := os.ReadDir(imageDir) // image lilst (ex: my-first-post_image1.png, my-first-post_image2.jpg)
	if err != nil {
		fmt.Printf("failed to read image directory: %v\n", err)
		return nil
	}

	for _, img := range images {
		if strings.HasPrefix(img.Name(), slug+"_") {
			err := os.Remove(filepath.Join(imageDir, img.Name())) // delete image file
			if err != nil {
				fmt.Printf("failed to delete image file %s: %v\n", img.Name(), err)
				continue
			}
		}
	}

	return nil
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
		if err := rows.Scan(&post.Slug, &post.Title, &post.Date, &post.Tags, &post.Summary, &post.Cover); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *sqliteStore) GetPost(slug string) (storedPost, error) {
	row := s.DB.QueryRow(`SELECT slug, title, date, tags, cover_img, summary, html_content FROM posts WHERE slug = ?`, slug)

	var post storedPost
	if err := row.Scan(&post.Slug, &post.Title, &post.Date, &post.Tags, &post.Cover, &post.Summary, &post.Content); err != nil {
		return storedPost{}, err
	}

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
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags[tag] = struct{}{}
	}

	tagArr := make([]string, 0, len(tags))

	for tag := range tags {
		tagArr = append(tagArr, tag)
	}

	return tagArr, nil
}
