// cache.go - Incremental build cache management
package sitegen

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type cacheFile struct {
	Version int                       `json:"version"`
	Files   map[string]cacheFileEntry `json:"files"`
}

type cacheFileEntry struct {
	Mtime  int64  `json:"mtime"`
	Output string `json:"output"`
}

func loadCache(path string) (*cacheFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var c cacheFile
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func saveCache(path string, c *cacheFile) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(c)
}

func newCache() *cacheFile {
	return &cacheFile{
		Version: 1,
		Files:   make(map[string]cacheFileEntry),
	}
}

func getCachePath(outputDir string) string {
	return filepath.Join(outputDir, ".colade-cache")
}

func getMtime(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.ModTime().Unix()
}
