package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type cacheStructure struct {
	Frontmatter map[string]frontMatter `json:"frontmatter"`
	Tags        []string               `json:"tags"`
	Details     map[string]string      `json:"details"`
}

// InitializeCache loads all posts from the filesystem and save them to JSON.
func InitializeCache() error {
	// 맵을 명시적으로 초기화합니다.
	cacheData := cacheStructure{
		Frontmatter: make(map[string]frontMatter),
		Details:     make(map[string]string),
		Tags:        []string{},
	}
	tagsSet := make(map[string]struct{})

	root, err := postsDirPath()
	if err != nil {
		fmt.Printf("FATAL ERROR: Could not resolve posts directory: %v\n", err)
		return err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Printf("FATAL ERROR: Could not read posts directory %s: %v\n", root, err)
		return err
	}

	fmt.Printf("INFO: Found %d directories in posts folder. Starting file processing...\n", len(entries))

	for _, entry := range entries {
		fullPath := path.Join(root, entry.Name(), "index.md")

		rawData, err := os.ReadFile(fullPath)
		if err != nil {
			// 🚨 오류 수정: 오류 발생 시 즉시 반환하지 않고, 로깅 후 다음 항목으로 넘어갑니다.
			fmt.Printf("WARNING: Failed to read index.md for %s: %v. Skipping this post.\n", entry.Name(), err)
			continue
		}

		meta, htmlBody, err := parsePostMarkdown(string(rawData))
		if err != nil {
			// 🚨 오류 수정: 오류 발생 시 즉시 반환하지 않고, 로깅 후 다음 항목으로 넘어갑니다.
			fmt.Printf("WARNING: Failed to parse post markdown for %s: %v. Skipping this post.\n", entry.Name(), err)
			continue
		}

		// 맵에 데이터를 안전하게 할당합니다.
		// meta 구조체에 Folder 정보를 추가합니다.
		meta.Folder = entry.Name()
		cacheData.Frontmatter[entry.Name()] = meta
		for _, tag := range meta.Tag {
			tagsSet[tag] = struct{}{} // add tag in tag set (automatically remove duplication)
		}
		cacheData.Details[entry.Name()] = htmlBody
		fmt.Printf("SUCCESS: Processed post [%s] successfully.\n", entry.Name())
	}

	for tag := range tagsSet {
		cacheData.Tags = append(cacheData.Tags, tag)
	}

	// 모든 포스트 처리가 끝난 후, 캐시를 저장합니다.
	err = saveCacheToJSON(cacheData)
	if err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

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

func saveCacheToJSON(data cacheStructure) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("ERROR: Failed to marshal cache data to JSON: %v", err)
	}

	// 🚨 디렉토리 존재 여부 확인 및 생성 로직 추가
	cacheDir := filepath.Dir(getCachePath())
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, 0o755)
		if err != nil {
			return fmt.Errorf("ERROR: Failed to create cache directory %s: %v", cacheDir, err)
		}
	}

	err = os.WriteFile(getCachePath(), jsonData, 0o644) // permissions: read/write for owner, read for others
	if err != nil {
		return fmt.Errorf("ERROR: Failed to write cache data to file: %v", err)
	}

	fmt.Println("INFO: Cache data successfully saved to cache.json")
	return nil
}

func readCacheFromJSON() (cacheStructure, error) {
	data, err := os.ReadFile(getCachePath())
	if err != nil {
		// 캐시 파일이 없거나 읽을 수 없는 경우 초기 빈 구조체를 반환합니다.
		// 이는 애플리케이션이 처음 시작되거나 캐시 파일이 삭제되었을 때 안정적으로 동작하도록 돕습니다.
		return cacheStructure{
			Frontmatter: make(map[string]frontMatter),
			Details:     make(map[string]string),
			Tags:        []string{},
		}, nil
	}

	var cacheData cacheStructure
	err = json.Unmarshal(data, &cacheData)
	if err != nil {
		return cacheStructure{}, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return cacheData, nil
}