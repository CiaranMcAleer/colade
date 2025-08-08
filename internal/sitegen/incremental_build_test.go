package sitegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIncrementalBuildCacheFilenames(t *testing.T) {
	dir, err := os.MkdirTemp("", "colade-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	inputDir := filepath.Join(dir, "input")
	outputDir := filepath.Join(dir, "output")
	if err := os.Mkdir(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.Mkdir(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Create multiple markdown files with links to each other
	md1 := `# Page One\n\nGo to [Page Two](two.md) and [Page Three](three.md)`
	md2 := `# Page Two\n\nBack to [Page One](one.md) or forward to [Page Three](three.md)`
	md3 := `# Page Three\n\nBack to [Page One](one.md) and [Page Two](two.md)`
	if err := os.WriteFile(filepath.Join(inputDir, "one.md"), []byte(md1), 0644); err != nil {
		t.Fatalf("failed to write one.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "two.md"), []byte(md2), 0644); err != nil {
		t.Fatalf("failed to write two.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "three.md"), []byte(md3), 0644); err != nil {
		t.Fatalf("failed to write three.md: %v", err)
	}

	// First build
	if err := BuildSite(inputDir, outputDir, 0, false, "", 0, false, "default", "", "", false, false, ""); err != nil {
		t.Fatalf("first build failed: %v", err)
	}
	checkOutputFiles(t, outputDir, []string{"one.html", "two.html", "three.html", ".colade-cache"})

	// Modify one file
	md2mod := `# Page Two\n\nBack to [Page One](one.md) or forward to [Page Three](three.md)\n\nExtra line!`
	if err := os.WriteFile(filepath.Join(inputDir, "two.md"), []byte(md2mod), 0644); err != nil {
		t.Fatalf("failed to modify two.md: %v", err)
	}

	// Second build (incremental)
	if err := BuildSite(inputDir, outputDir, 0, false, "", 0, false, "default", "", "", false, false, ""); err != nil {
		t.Fatalf("second build failed: %v", err)
	}
	checkOutputFiles(t, outputDir, []string{"one.html", "two.html", "three.html", ".colade-cache"})
}

func checkOutputFiles(t *testing.T, outputDir string, expected []string) {
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("failed to read output dir: %v", err)
	}
	found := map[string]bool{}
	for _, f := range files {
		found[f.Name()] = true
	}
	for _, name := range expected {
		if !found[name] {
			t.Errorf("expected output file %q not found", name)
		}
	}
	// Check for unexpected files
	for name := range found {
		if !contains(expected, name) {
			t.Errorf("unexpected output file: %q", name)
		}
	}
	// Optionally, check .colade-cache for filename keys
	cachePath := filepath.Join(outputDir, ".colade-cache")
	if found[".colade-cache"] {
		data, err := os.ReadFile(cachePath)
		if err != nil {
			t.Errorf("failed to read cache: %v", err)
		} else {
			for _, fname := range []string{"one.md", "two.md", "three.md"} {
				if !strings.Contains(string(data), fname) {
					t.Errorf("cache file missing expected filename: %q", fname)
				}
			}
		}
	}

	// Check contents of generated HTML files
	htmlChecks := map[string][]string{
		"one.html": {
			"Page One",
			"Page Two",
			"Page Three",
			"two.html",
			"three.html",
		},
		"two.html": {
			"Page Two",
			"Page One",
			"Page Three",
			"one.html",
			"three.html",
		},
		"three.html": {
			"Page Three",
			"Page One",
			"Page Two",
			"one.html",
			"two.html",
		},
	}
	for fname, substrs := range htmlChecks {
		htmlPath := filepath.Join(outputDir, fname)
		data, err := os.ReadFile(htmlPath)
		if err != nil {
			t.Errorf("failed to read %s: %v", fname, err)
			continue
		}
		content := string(data)
		for _, sub := range substrs {
			if !strings.Contains(content, sub) {
				t.Errorf("%s missing expected content: %q", fname, sub)
			}
		}
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
