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

	"encoding/json"

	"github.com/yuin/goldmark"
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

func BuildSite(inputDir, outputDir string, sizeThreshold int, noIncremental bool, rssURL string, rssMaxItems int) error {
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

	cachePath := filepath.Join(outputDir, ".colade-cache")
	var cache *cacheFile
	useIncremental := !noIncremental
	if useIncremental {
		c, err := loadCache(cachePath)
		if err == nil && c.Version == 1 {
			cache = c
			fmt.Printf("[Build] Loaded cache from %s\n", cachePath)
		} else {
			fmt.Printf("[Build] No valid cache found, doing full rebuild\n")
			useIncremental = false
		}
	}

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
	if useIncremental && cache != nil {
		md := goldmark.New()
		sizeOut := make(chan string, len(markdownFiles))
		newCache := &cacheFile{Version: 1, Files: make(map[string]cacheFileEntry)}

		getMtime := func(path string) int64 {
			info, err := os.Stat(path)
			if err != nil {
				return 0
			}
			return info.ModTime().Unix()
		}

		seen := make(map[string]bool)

		for _, relPath := range markdownFiles {
			src := filepath.Join(inputDir, relPath)
			dst := filepath.Join(outputDir, relPath)
			dst = dst[:len(dst)-len(filepath.Ext(dst))] + ".html"
			mtime := getMtime(src)
			seen[relPath] = true

			prev, ok := cache.Files[relPath]
			if !ok || prev.Mtime != mtime {
				fmt.Printf("[IncBuild] %s -> %s (changed/new)\n", relPath, dst)
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
				CheckGzipSize(dst, sizeThreshold, sizeOut)
			} else {
				fmt.Printf("[IncBuild] %s unchanged, skipping\n", relPath)
				sizeOut <- ""
			}
			newCache.Files[relPath] = cacheFileEntry{Mtime: mtime, Output: dst[len(outputDir)+1:]}
		}

		for _, relPath := range assetFiles {
			src := filepath.Join(inputDir, relPath)
			dst := filepath.Join(outputDir, relPath)
			mtime := getMtime(src)
			seen[relPath] = true

			prev, ok := cache.Files[relPath]
			if !ok || prev.Mtime != mtime {
				fmt.Printf("[IncCopy] %s -> %s (changed/new)\n", relPath, dst)
				if err := copyFilePreserveDirs(src, dst); err != nil {
					return fmt.Errorf("failed to copy asset '%s': %w", relPath, err)
				}
			} else {
				fmt.Printf("[IncCopy] %s unchanged, skipping\n", relPath)
			}
			newCache.Files[relPath] = cacheFileEntry{Mtime: mtime, Output: dst[len(outputDir)+1:]}
		}

		for relPath, entry := range cache.Files {
			if !seen[relPath] {
				outPath := filepath.Join(outputDir, entry.Output)
				fmt.Printf("[IncRemove] %s (deleted from input, removing %s)\n", relPath, outPath)
				os.Remove(outPath)
			}
		}

		for i := 0; i < len(markdownFiles); i++ {
			fmt.Fprint(os.Stderr, <-sizeOut)
		}

		// Generate RSS feed if requested
		if rssURL != "" {
			rssGen := NewRSSGenerator(rssURL, outputDir)
			if err := rssGen.Generate(markdownFiles, inputDir, rssMaxItems); err != nil {
				return fmt.Errorf("failed to generate RSS feed: %w", err)
			}
		}

		if err := saveCache(cachePath, newCache); err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}

		fmt.Printf("[Build] Incremental build complete in %v.\n", time.Since(startTime))
		return nil
	}

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
		CheckGzipSize(dst, sizeThreshold, sizeOut)
		fmt.Printf("[Build]  Done in %v\n", time.Since(opStart))
	}
	// Print all size check results(doing it this way to avoid slowing down the build process)
	for i := 0; i < len(markdownFiles); i++ {
		fmt.Fprint(os.Stderr, <-sizeOut)
	}

	// Generate RSS feed if requested
	if rssURL != "" {
		rssGen := NewRSSGenerator(rssURL, outputDir)
		if err := rssGen.Generate(markdownFiles, inputDir, rssMaxItems); err != nil {
			return fmt.Errorf("failed to generate RSS feed: %w", err)
		}
	}

	filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		relPath, err := filepath.Rel(outputDir, path)
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

		expected := false
		for _, f := range markdownFiles {
			out := f[:len(f)-len(filepath.Ext(f))] + ".html"
			if relPath == out {
				expected = true
				break
			}
		}
		for _, f := range assetFiles {
			if relPath == f {
				expected = true
				break
			}
		}
		// Don't clean up RSS feed
		if relPath == "feed.xml" && rssURL != "" {
			expected = true
		}
		if !expected {
			fmt.Printf("[Clean] Removing orphaned output: %s\n", path)
			os.Remove(path)
		}
		return nil
	})

	// Write cache after full rebuild
	newCache := &cacheFile{Version: 1, Files: make(map[string]cacheFileEntry)}
	for _, f := range markdownFiles {
		src := filepath.Join(inputDir, f)
		mtime := int64(0)
		if info, err := os.Stat(src); err == nil {
			mtime = info.ModTime().Unix()
		}
		out := f[:len(f)-len(filepath.Ext(f))] + ".html"
		newCache.Files[f] = cacheFileEntry{Mtime: mtime, Output: out}
	}
	for _, f := range assetFiles {
		src := filepath.Join(inputDir, f)
		mtime := int64(0)
		if info, err := os.Stat(src); err == nil {
			mtime = info.ModTime().Unix()
		}
		newCache.Files[f] = cacheFileEntry{Mtime: mtime, Output: f}
	}
	cachePath = filepath.Join(outputDir, ".colade-cache")
	if err := saveCache(cachePath, newCache); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
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
