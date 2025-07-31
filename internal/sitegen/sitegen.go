// core logic for building static sites from Markdown files.
package sitegen

import (
	"fmt"
	"os"
	"time"
)

func BuildSite(inputDir, outputDir string, sizeThreshold int, noIncremental bool, rssURL string, rssMaxItems int, keepOrphaned bool) error {
	// Validate inputs and create output directory
	if err := validateInputsAndCreateOutput(inputDir, outputDir); err != nil {
		return err
	}

	startTime := time.Now()
	fmt.Printf("[Build] Starting site build from '%s' to '%s'...\n", inputDir, outputDir)

	// Discover files
	fileSet, err := DiscoverFiles(inputDir)
	if err != nil {
		return fmt.Errorf("error discovering files: %w", err)
	}

	logDiscoveredFiles(fileSet)

	// Try incremental build first
	if !noIncremental {
		if completed, err := tryIncrementalBuild(inputDir, outputDir, sizeThreshold, rssURL, rssMaxItems, fileSet, startTime, keepOrphaned); err != nil {
			return err
		} else if completed {
			return nil
		}
	}

	// Fall back to full build
	return performFullBuild(inputDir, outputDir, sizeThreshold, rssURL, rssMaxItems, fileSet, startTime, keepOrphaned)
}

// validateInputsAndCreateOutput validates input directory and creates output directory
func validateInputsAndCreateOutput(inputDir, outputDir string) error {
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

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	return nil
}

// logDiscoveredFiles logs the discovered files
func logDiscoveredFiles(fileSet *FileSet) {
	fmt.Printf("[Build] Found %d markdown files and %d asset files.\n", len(fileSet.MarkdownFiles), len(fileSet.AssetFiles))
	for _, f := range fileSet.MarkdownFiles {
		fmt.Printf("    [Markdown] %s\n", f)
	}
	for _, f := range fileSet.AssetFiles {
		fmt.Printf("    [Asset] %s\n", f)
	}
}

// tryIncrementalBuild attempts an incremental build, returns (completed, error)
func tryIncrementalBuild(inputDir, outputDir string, sizeThreshold int, rssURL string, rssMaxItems int, fileSet *FileSet, startTime time.Time, keepOrphaned bool) (bool, error) {
	cachePath := getCachePath(outputDir)
	cache, err := loadCache(cachePath)
	if err != nil || cache.Version != 1 {
		fmt.Printf("[Build] No valid cache found, doing full rebuild\n")
		return false, nil
	}

	fmt.Printf("[Build] Loaded cache from %s\n", cachePath)

	// Perform incremental build
	builder := NewIncrementalBuilder(inputDir, outputDir, sizeThreshold, cache)
	sizeOut := make(chan string, len(fileSet.MarkdownFiles))

	// Process files incrementally
	if err := builder.ProcessMarkdownFiles(fileSet.MarkdownFiles, sizeOut); err != nil {
		return false, err
	}
	if err := builder.ProcessAssetFiles(fileSet.AssetFiles); err != nil {
		return false, err
	}

	// Cleanup removed files (if not keeping orphaned files)
	if !keepOrphaned {
		builder.CleanupRemovedFiles()
	}

	// Print size check results
	for i := 0; i < len(fileSet.MarkdownFiles); i++ {
		fmt.Fprint(os.Stderr, <-sizeOut)
	}

	// Generate RSS feed and save cache
	if err := generateRSSFeed(rssURL, outputDir, fileSet.MarkdownFiles, inputDir, rssMaxItems); err != nil {
		return false, err
	}

	cacheManager := NewCacheManager(inputDir, outputDir)
	if err := cacheManager.SaveCache(builder.GetNewCache()); err != nil {
		return false, fmt.Errorf("failed to save cache: %w", err)
	}

	fmt.Printf("[Build] Incremental build complete in %v.\n", time.Since(startTime))
	return true, nil
}

// performFullBuild performs a complete rebuild
func performFullBuild(inputDir, outputDir string, sizeThreshold int, rssURL string, rssMaxItems int, fileSet *FileSet, startTime time.Time, keepOrphaned bool) error {
	builder := NewFullBuilder(inputDir, outputDir, sizeThreshold)

	// Process asset files
	if err := builder.ProcessAssetFiles(fileSet.AssetFiles); err != nil {
		return err
	}

	// Process markdown files
	sizeOut := make(chan string, len(fileSet.MarkdownFiles))
	if err := builder.ProcessMarkdownFiles(fileSet.MarkdownFiles, sizeOut); err != nil {
		return err
	}

	// Print size check results
	for i := 0; i < len(fileSet.MarkdownFiles); i++ {
		fmt.Fprint(os.Stderr, <-sizeOut)
	}

	// Generate RSS feed
	if err := generateRSSFeed(rssURL, outputDir, fileSet.MarkdownFiles, inputDir, rssMaxItems); err != nil {
		return err
	}

	// Cleanup orphaned files (if not keeping orphaned files)
	if !keepOrphaned {
		cleaner := NewOutputCleaner(outputDir, rssURL)
		if err := cleaner.CleanupOrphanedFiles(fileSet); err != nil {
			return err
		}
	}

	// Create and save cache
	cacheManager := NewCacheManager(inputDir, outputDir)
	newCache, err := cacheManager.CreateCacheFromFileSet(fileSet)
	if err != nil {
		return err
	}
	if err := cacheManager.SaveCache(newCache); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	fmt.Printf("[Build] Site build complete in %v.\n", time.Since(startTime))
	return nil
}

// generateRSSFeed generates RSS feed if requested
func generateRSSFeed(rssURL, outputDir string, markdownFiles []string, inputDir string, rssMaxItems int) error {
	if rssURL != "" {
		rssGen := NewRSSGenerator(rssURL, outputDir)
		if err := rssGen.Generate(markdownFiles, inputDir, rssMaxItems); err != nil {
			return fmt.Errorf("failed to generate RSS feed: %w", err)
		}
	}
	return nil
}
