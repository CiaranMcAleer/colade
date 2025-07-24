// core logic for building static sites from Markdown files.
package sitegen

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func BuildSite(inputDir, outputDir string) error {
	// Check if input directory exists
	info, err := os.Stat(inputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input directory does not exist: %s", inputDir)
		}
		return fmt.Errorf("error checking input directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("input path is not a directory: %s", inputDir)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("[Stub] Building site from '%s' to '%s'...\n", inputDir, outputDir)

	var markdownFiles []string
	var assetFiles []string

	// Traverse the input directory to find markdown and asset files(mostly images)
	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}
		ext := filepath.Ext(info.Name())
		switch ext {
		case ".md", ".markdown":
			markdownFiles = append(markdownFiles, relPath)
		default:
			assetFiles = append(assetFiles, relPath)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking input directory: %w", err)
	}

	fmt.Printf("Found %d markdown files and %d asset files.\n", len(markdownFiles), len(assetFiles))
	for _, f := range markdownFiles {
		fmt.Printf("  [Markdown] %s\n", f)
	}
	for _, f := range assetFiles {
		fmt.Printf("  [Asset] %s\n", f)
	}

	// Copy asset files to output directory, preserving relative paths
	// TODO add tests for asset copying functionality 
	for _, relPath := range assetFiles {
		src := filepath.Join(inputDir, relPath)
		dst := filepath.Join(outputDir, relPath)
		if err := copyFilePreserveDirs(src, dst); err != nil {
			return fmt.Errorf("failed to copy asset '%s': %w", relPath, err)
		}
	}

	return nil
}

// copyFilePreserveDirs copies a file from src to dst, creating parent directories as needed.
func copyFilePreserveDirs(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
