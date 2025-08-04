// processor.go - File processing and conversion logic
package sitegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
	mermaid "go.abhg.dev/goldmark/mermaid"
)

// MarkdownProcessor handles markdown file conversion
type MarkdownProcessor struct {
	md          goldmark.Markdown
	templateOpt string
}

// NewMarkdownProcessor creates a new markdown processor
func NewMarkdownProcessor(templateOpt string) *MarkdownProcessor {
	return &MarkdownProcessor{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				meta.Meta,
				&mermaid.Extender{},
				&frontmatter.Extender{},
			),
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
			),
		),
		templateOpt: templateOpt,
	}
}

// ProcessMarkdownFile converts a single markdown file to HTML
func (mp *MarkdownProcessor) ProcessMarkdownFile(inputDir, outputDir, relPath string, sizeThreshold int, sizeOut chan<- string) error {
	src := filepath.Join(inputDir, relPath)
	dst := filepath.Join(outputDir, relPath)
	dst = dst[:len(dst)-len(filepath.Ext(dst))] + ".html"

	content, err := parseMarkdownFile(src)
	if err != nil {
		return fmt.Errorf("failed to read markdown file '%s': %w", relPath, err)
	}

	content = replaceMdLinks(content)
	var buf bytes.Buffer
	if err := mp.md.Convert(content, &buf); err != nil {
		return fmt.Errorf("failed to convert markdown '%s': %w", relPath, err)
	}

	// Frontmatter is automatically stripped from content by the extension
	htmlOut := renderHTMLPage(buf.Bytes(), mp.templateOpt)
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create output dir for '%s': %w", relPath, err)
	}

	if err := os.WriteFile(dst, htmlOut, 0644); err != nil {
		return fmt.Errorf("failed to write HTML file '%s': %w", relPath, err)
	}

	CheckGzipSize(dst, sizeThreshold, sizeOut)
	return nil
}

// ProcessAssetFile copies a single asset file
func ProcessAssetFile(inputDir, outputDir, relPath string) error {
	src := filepath.Join(inputDir, relPath)
	dst := filepath.Join(outputDir, relPath)
	return copyFilePreserveDirs(src, dst)
}

// IncrementalBuilder handles incremental build logic
type IncrementalBuilder struct {
	processor     *MarkdownProcessor
	inputDir      string
	outputDir     string
	sizeThreshold int
	cache         *cacheFile
	newCache      *cacheFile
	seen          map[string]bool
	templateOpt   string
}

// NewIncrementalBuilder creates a new incremental builder
func NewIncrementalBuilder(inputDir, outputDir string, sizeThreshold int, cache *cacheFile, templateOpt string) *IncrementalBuilder {
	return &IncrementalBuilder{
		processor:     NewMarkdownProcessor(templateOpt),
		inputDir:      inputDir,
		outputDir:     outputDir,
		sizeThreshold: sizeThreshold,
		cache:         cache,
		newCache:      newCache(),
		seen:          make(map[string]bool),
		templateOpt:   templateOpt,
	}
}

// ProcessMarkdownFiles processes all markdown files incrementally
func (ib *IncrementalBuilder) ProcessMarkdownFiles(markdownFiles []string, sizeOut chan<- string) error {
	for _, relPath := range markdownFiles {
		src := filepath.Join(ib.inputDir, relPath)
		dst := filepath.Join(ib.outputDir, relPath)
		dst = dst[:len(dst)-len(filepath.Ext(dst))] + ".html"
		mtime := getMtime(src)
		ib.seen[relPath] = true

		prev, ok := ib.cache.Files[relPath]
		if !ok || prev.Mtime != mtime {
			fmt.Printf("[IncBuild] %s -> %s (changed/new)\n", relPath, dst)
			if err := ib.processor.ProcessMarkdownFile(ib.inputDir, ib.outputDir, relPath, ib.sizeThreshold, sizeOut); err != nil {
				return err
			}
		} else {
			fmt.Printf("[IncBuild] %s unchanged, skipping\n", relPath)
			sizeOut <- ""
		}
		outputPath, err := filepath.Rel(ib.outputDir, dst)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		ib.newCache.Files[relPath] = cacheFileEntry{Mtime: mtime, Output: outputPath}
	}
	return nil
}

// ProcessAssetFiles processes all asset files incrementally
func (ib *IncrementalBuilder) ProcessAssetFiles(assetFiles []string) error {
	for _, relPath := range assetFiles {
		src := filepath.Join(ib.inputDir, relPath)
		dst := filepath.Join(ib.outputDir, relPath)
		mtime := getMtime(src)
		ib.seen[relPath] = true

		prev, ok := ib.cache.Files[relPath]
		if !ok || prev.Mtime != mtime {
			fmt.Printf("[IncCopy] %s -> %s (changed/new)\n", relPath, dst)
			if err := ProcessAssetFile(ib.inputDir, ib.outputDir, relPath); err != nil {
				return fmt.Errorf("failed to copy asset '%s': %w", relPath, err)
			}
		} else {
			fmt.Printf("[IncCopy] %s unchanged, skipping\n", relPath)
		}
		outputPath, err := filepath.Rel(ib.outputDir, dst)
		if err != nil {
			return fmt.Errorf("failed to get relative path for asset: %w", err)
		}
		ib.newCache.Files[relPath] = cacheFileEntry{Mtime: mtime, Output: outputPath}
	}
	return nil
}

// CleanupRemovedFiles removes files that no longer exist in input
func (ib *IncrementalBuilder) CleanupRemovedFiles() {
	for relPath, entry := range ib.cache.Files {
		if !ib.seen[relPath] {
			outPath := filepath.Join(ib.outputDir, entry.Output)
			fmt.Printf("[IncRemove] %s (deleted from input, removing %s)\n", relPath, outPath)
			os.Remove(outPath)
		}
	}
}

// GetNewCache returns the updated cache
func (ib *IncrementalBuilder) GetNewCache() *cacheFile {
	return ib.newCache
}

// FullBuilder handles full build logic
type FullBuilder struct {
	processor     *MarkdownProcessor
	inputDir      string
	outputDir     string
	sizeThreshold int
	templateOpt   string
}

// NewFullBuilder creates a new full builder
func NewFullBuilder(inputDir, outputDir string, sizeThreshold int, templateOpt string) *FullBuilder {
	return &FullBuilder{
		processor:     NewMarkdownProcessor(templateOpt),
		inputDir:      inputDir,
		outputDir:     outputDir,
		sizeThreshold: sizeThreshold,
		templateOpt:   templateOpt,
	}
}

// ProcessAssetFiles processes all asset files in full build mode
func (fb *FullBuilder) ProcessAssetFiles(assetFiles []string) error {
	for _, relPath := range assetFiles {
		opStart := time.Now()
		fmt.Printf("[Copy]   %s -> %s\n", relPath, filepath.Join(fb.outputDir, relPath))
		if err := ProcessAssetFile(fb.inputDir, fb.outputDir, relPath); err != nil {
			return fmt.Errorf("failed to copy asset '%s': %w", relPath, err)
		}
		fmt.Printf("[Copy]   Done in %v\n", time.Since(opStart))
	}
	return nil
}

// ProcessMarkdownFiles processes all markdown files in full build mode
func (fb *FullBuilder) ProcessMarkdownFiles(markdownFiles []string, sizeOut chan<- string) error {
	for _, relPath := range markdownFiles {
		opStart := time.Now()
		fmt.Printf("[Build]  %s -> %s\n", relPath, filepath.Join(fb.outputDir, relPath[:len(relPath)-len(filepath.Ext(relPath))]+".html"))

		if err := fb.processor.ProcessMarkdownFile(fb.inputDir, fb.outputDir, relPath, fb.sizeThreshold, sizeOut); err != nil {
			return err
		}
		fmt.Printf("[Build]  Done in %v\n", time.Since(opStart))
	}
	return nil
}
