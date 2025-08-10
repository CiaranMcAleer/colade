package sitegen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper: create many markdown files
func createManyMarkdownFiles(dir string, count int) error {
	for i := 0; i < count; i++ {
		filename := filepath.Join(dir, fmt.Sprintf("page%d.md", i))
		content := fmt.Sprintf("# Page %d\n\nThis is page %d.", i, i)
		if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// Helper: create many asset files (simulate images)
func createManyAssets(dir string, count int) error {
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		filename := filepath.Join(assetsDir, fmt.Sprintf("img%d.jpg", i))
		if err := ioutil.WriteFile(filename, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 0644); err != nil {
			return err
		}
	}
	return nil
}

// Helper: create deeply nested markdown files
func createDeeplyNestedMarkdown(dir string, depth int, breadth int) error {
	var create func(path string, d int) error
	create = func(path string, d int) error {
		if d == 0 {
			return nil
		}
		for i := 0; i < breadth; i++ {
			subdir := filepath.Join(path, fmt.Sprintf("level%d_%d", d, i))
			if err := os.MkdirAll(subdir, 0755); err != nil {
				return err
			}
			filename := filepath.Join(subdir, fmt.Sprintf("page%d.md", d))
			content := fmt.Sprintf("# Nested Page d=%d i=%d", d, i)
			if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
				return err
			}
			if err := create(subdir, d-1); err != nil {
				return err
			}
		}
		return nil
	}
	return create(dir, depth)
}

func BenchmarkSitegenPerformance(b *testing.B) {
	scenarios := []struct {
		name   string
		setup  func(inputDir string) error
		change func(inputDir string) error
	}{
		{
			name: "ManyMarkdownPages",
			setup: func(inputDir string) error {
				return createManyMarkdownFiles(inputDir, 1000)
			},
			change: func(inputDir string) error {
				// Edit one file
				return ioutil.WriteFile(filepath.Join(inputDir, "page0.md"), []byte("# Updated\n\nChanged."), 0644)
			},
		},
		{
			name: "ManyAssets",
			setup: func(inputDir string) error {
				if err := createManyMarkdownFiles(inputDir, 10); err != nil {
					return err
				}
				return createManyAssets(inputDir, 1000)
			},
			change: func(inputDir string) error {
				// Add a new asset
				return ioutil.WriteFile(filepath.Join(inputDir, "assets", "img_new.jpg"), []byte{1, 2, 3, 4, 5}, 0644)
			},
		},
		{
			name: "DeeplyNestedPages",
			setup: func(inputDir string) error {
				return createDeeplyNestedMarkdown(inputDir, 6, 3) // 6 levels, 3 branches each
			},
			change: func(inputDir string) error {
				// Edit a leaf file
				leaf := filepath.Join(inputDir, "level6_0", "level5_0", "level4_0", "level3_0", "level2_0", "level1_0", "page1.md")
				return ioutil.WriteFile(leaf, []byte("# Updated Nested\n\nChanged."), 0644)
			},
		},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			inputDir := b.TempDir()
			outputDir := b.TempDir()
			if err := scenario.setup(inputDir); err != nil {
				b.Fatalf("setup failed: %v", err)
			}
			// Initial build
			start := time.Now()
			err := BuildSite(inputDir, outputDir, 14*1024, false, "", 20, false, "default", "", "", false, false, "")
			if err != nil {
				b.Fatalf("initial build failed: %v", err)
			}
			fullBuildTime := time.Since(start)
			// Make a change
			if err := scenario.change(inputDir); err != nil {
				b.Fatalf("change failed: %v", err)
			}
			// Incremental build
			start = time.Now()
			err = BuildSite(inputDir, outputDir, 14*1024, false, "", 20, false, "default", "", "", false, false, "")
			if err != nil {
				b.Fatalf("incremental build failed: %v", err)
			}
			incrBuildTime := time.Since(start)
			b.ReportMetric(float64(fullBuildTime.Milliseconds()), "full_build_ms")
			b.ReportMetric(float64(incrBuildTime.Milliseconds()), "incr_build_ms")
		})
	}
}
