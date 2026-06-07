// Package api provides functions to initialize and manage a cache of post metadata and details.
package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type cacheStructure struct {
	// Maps must be explicitly initialized to prevent 'assignment to entry in nil map' panic.
	frontmatter map[string]frontMatter
	tags        []string
	details     map[string]string
}

// InitializeCache loads all posts from the filesystem and save them to JSON.
func InitializeCache() error {
	// 맵을 명시적으로 초기화합니다.
	cacheData := cacheStructure{
		frontmatter: make(map[string]frontMatter),
		details:     make(map[string]string),
		tags:        []string{},
	}
	tagsSet := make(map[string]struct{})

	root, err := postsDirPath()
	if err != nil {
		fmt.Printf("FATAL ERROR: Could not resolve posts directory: %v\n", err)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Printf("FATAL ERROR: Could not read posts directory %s: %v\n", root, err)
	}

	fmt.Printf("INFO: Found %d directories in posts folder. Starting file processing...\n", len(entries))

	for _, entry := range entries {
		rawData, err := os.ReadFile(path.Join(root, entry.Name(), "index.md"))
		if err != nil {
			return fmt.Errorf("failed to read index.md for %s: %w", entry.Name(), err)
		}

		meta, mainPostsData, err := parsePostMarkdown(string(rawData))
		if err != nil {
			return fmt.Errorf("failed to parse post markdown for %s: %w", entry.Name(), err)
		}

		// 맵에 데이터를 안전하게 할당합니다.
		cacheData.frontmatter[entry.Name()] = meta
		for _, tag := range meta.Tag {
			tagsSet[tag] = struct{}{} // add tag in tag set (automatically remove duplication)
		}
		cacheData.details[entry.Name()] = mainPostsData
	}

	saveCacheToJsoon(cacheData)

	return nil
}

func postsDirPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "posts"), nil
}

func getCachePath() string {
	wd, err := os.Getwd()
	if err != nil {
		return "cache.json" // fallback to current directory
	}
	return filepath.Join(wd, "cache", "cache.json")
}

func saveCacheToJsoon(data cacheStructure) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Errorf("ERROR: Failed to marshal cache data to JSON: %v\n", err)
		return err
	}

	err = os.WriteFile(getCachePath(), jsonData, 0o644) // permissions: read/write for owner, read for others
	if err != nil {
		fmt.Errorf("ERROR: Failed to write cache data to file: %v\n", err)
	}

	fmt.Println("INFO: Cache data successfully saved to cache.json")
	return nil
}

func readCacheFromJSON() (cacheStructure, error) {
	data, err := os.ReadFile(getCachePath())
	if err != nil {
		// 캐시 파일이 없거나 읽을 수 없는 경우 초기 빈 구조체를 반환합니다.
		return cacheStructure{
			frontmatter: make(map[string]frontMatter),
			details:     make(map[string]string),
		}, nil // 오류를 반환하지 않고 빈 캐시를 반환하여 프로그램이 계속 진행되게 합니다.
	}

	var cacheData cacheStructure
	err = json.Unmarshal(data, &cacheData)
	if err != nil {
		return cacheStructure{}, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return cacheData, nil
}