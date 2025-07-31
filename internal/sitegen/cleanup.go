// cleanup.go - Output directory cleanup and cache management
package sitegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type OutputCleaner struct {
	outputDir string
	rssURL    string
}

func NewOutputCleaner(outputDir, rssURL string) *OutputCleaner {
	return &OutputCleaner{
		outputDir: outputDir,
		rssURL:    rssURL,
	}
}

func (oc *OutputCleaner) CleanupOrphanedFiles(fileSet *FileSet) error {
	return filepath.Walk(oc.outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		relPath, err := filepath.Rel(oc.outputDir, path)
		if err != nil || relPath == "." {
			return nil
		}
		if relPath == ".colade-cache" || strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		expected := oc.isExpectedFile(relPath, fileSet)
		if !expected {
			fmt.Printf("[Clean] Removing orphaned output: %s\n", path)
			os.Remove(path)
		}
		return nil
	})
}

func (oc *OutputCleaner) isExpectedFile(relPath string, fileSet *FileSet) bool {
	for _, f := range fileSet.MarkdownFiles {
		out := f[:len(f)-len(filepath.Ext(f))] + ".html"
		if relPath == out {
			return true
		}
	}

	for _, f := range fileSet.AssetFiles {
		if relPath == f {
			return true
		}
	}

	// Don't clean up generated RSS feed
	if relPath == "feed.xml" && oc.rssURL != "" {
		return true
	}

	return false
}

// CacheManager handles cache operations for full builds
type CacheManager struct {
	inputDir  string
	outputDir string
}

func NewCacheManager(inputDir, outputDir string) *CacheManager {
	return &CacheManager{
		inputDir:  inputDir,
		outputDir: outputDir,
	}
}

func (cm *CacheManager) CreateCacheFromFileSet(fileSet *FileSet) (*cacheFile, error) {
	newCache := newCache()

	// Add markdown files to cache
	for _, f := range fileSet.MarkdownFiles {
		src := filepath.Join(cm.inputDir, f)
		mtime := int64(0)
		if info, err := os.Stat(src); err == nil {
			mtime = info.ModTime().Unix()
		}
		out := f[:len(f)-len(filepath.Ext(f))] + ".html"
		newCache.Files[f] = cacheFileEntry{Mtime: mtime, Output: out}
	}

	// Add asset files to cache
	for _, f := range fileSet.AssetFiles {
		src := filepath.Join(cm.inputDir, f)
		mtime := int64(0)
		if info, err := os.Stat(src); err == nil {
			mtime = info.ModTime().Unix()
		}
		newCache.Files[f] = cacheFileEntry{Mtime: mtime, Output: f}
	}

	return newCache, nil
}

func (cm *CacheManager) SaveCache(cache *cacheFile) error {
	cachePath := getCachePath(cm.outputDir)
	return saveCache(cachePath, cache)
}
