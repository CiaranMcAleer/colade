// core logic for building static sites from Markdown files.
package sitegen

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
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

	startTime := time.Now() // Used to measure build time
	fmt.Printf("[Build] Starting site build from '%s' to '%s'...\n", inputDir, outputDir)

	var markdownFiles []string
	var assetFiles []string

	// Traverse the input directory to find markdown and asset files (skip hidden files/dirs)
	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}
		// Skip hidden files and directories
		if relPath != "." {
			// Use filepath.Dir and filepath.Base for better cross-platform handling
			parts := strings.Split(filepath.ToSlash(relPath), "/")
			for _, part := range parts {
				if strings.HasPrefix(part, ".") {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}
		if info.IsDir() {
			return nil
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

	fmt.Printf("[Build] Found %d markdown files and %d asset files.\n", len(markdownFiles), len(assetFiles))
	for _, f := range markdownFiles {
		fmt.Printf("    [Markdown] %s\n", f)
	}
	for _, f := range assetFiles {
		fmt.Printf("    [Asset] %s\n", f)
	}

	// Copy asset files to output directory, preserving relative paths, TODO add tests for this
	for _, relPath := range assetFiles {
		src := filepath.Join(inputDir, relPath)
		dst := filepath.Join(outputDir, relPath)
		opStart := time.Now()
		fmt.Printf("[Copy]   %s -> %s\n", relPath, dst)
		if err := copyFilePreserveDirs(src, dst); err != nil {
			return fmt.Errorf("failed to copy asset '%s': %w", relPath, err)
		}
		fmt.Printf("[Copy]   Done in %v\n", time.Since(opStart))
	}

	// Convert markdown files to HTML and write to output directory
	md := goldmark.New()
	sizeOut := make(chan string, len(markdownFiles))
	for _, relPath := range markdownFiles {
		src := filepath.Join(inputDir, relPath)
		dst := filepath.Join(outputDir, relPath)
		dst = dst[:len(dst)-len(filepath.Ext(dst))] + ".html"
		opStart := time.Now()
		fmt.Printf("[Build]  %s -> %s\n", relPath, dst)

		content, err := parseMarkdownFile(src)
		if err != nil {
			return fmt.Errorf("failed to read markdown file '%s': %w", relPath, err)
		}
		var buf bytes.Buffer
		if err := md.Convert(content, &buf); err != nil {
			return fmt.Errorf("failed to convert markdown '%s': %w", relPath, err)
		}
		htmlOut := renderHTMLPage(buf.Bytes())
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("failed to create output dir for '%s': %w", relPath, err)
		}
		if err := os.WriteFile(dst, htmlOut, 0644); err != nil {
			return fmt.Errorf("failed to write HTML file '%s': %w", relPath, err)
		}
		CheckGzipSize(dst, 14*1024, sizeOut)
		fmt.Printf("[Build]  Done in %v\n", time.Since(opStart))
	}
	// Print all size check results(doing it this way to avoid slowing down the build process)
	for i := 0; i < len(markdownFiles); i++ {
		fmt.Fprint(os.Stderr, <-sizeOut)
	}

	fmt.Printf("[Build] Site build complete in %v.\n", time.Since(startTime))
	return nil
}

// parseMarkdownFile is a future-proof extension point for frontmatter support.
func parseMarkdownFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// renderHTMLPage is a future-proof extension point for templating support.
func renderHTMLPage(html []byte) []byte {
	// For now, just return the HTML as-is.
	return html
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
