// discovery.go - File discovery and classification
package sitegen

import (
	"os"
	"path/filepath"
	"strings"
)

type FileSet struct {
	MarkdownFiles []string
	AssetFiles    []string
}

// DiscoverFiles walks the input directory and discovers markdown and asset files
// Returns FileSet containing classified files, skipping hidden files/directories
func DiscoverFiles(inputDir string) (*FileSet, error) {
	var markdownFiles []string
	var assetFiles []string

	// Traverse the input directory to find markdown and asset files (skip hidden files/dirs)
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if isHiddenFile(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		fileType := classifyFile(info.Name())
		switch fileType {
		case "markdown":
			markdownFiles = append(markdownFiles, relPath)
		case "asset":
			assetFiles = append(assetFiles, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &FileSet{
		MarkdownFiles: markdownFiles,
		AssetFiles:    assetFiles,
	}, nil
}

// isHiddenFile checks if a file or directory should be skipped based on hidden file rules
func isHiddenFile(relPath string) bool {
	if relPath == "." {
		return false
	}
	// Use filepath.Dir and filepath.Base for better cross-platform handling
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

// classifyFile determines if a file is markdown, asset, or should be skipped
func classifyFile(name string) string {
	ext := filepath.Ext(name)
	switch ext {
	case ".md", ".markdown":
		return "markdown"
	default:
		return "asset"
	}
}
