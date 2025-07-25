package sitegen

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSite_InputDirValidation(t *testing.T) {
	t.Run("nonexistent input dir", func(t *testing.T) {
		err := BuildSite("/unlikely/to/exist/colade_test_input", t.TempDir())
		if err == nil || err.Error() == "" {
			t.Error("expected error for nonexistent input directory, got nil")
		}
	})

	t.Run("input path is file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "file.md")
		os.WriteFile(file, []byte("# test"), 0644)
		err := BuildSite(file, t.TempDir())
		if err == nil || err.Error() == "" {
			t.Error("expected error for input path as file, got nil")
		}
	})

	t.Run("valid input dir", func(t *testing.T) {
		inputDir := t.TempDir()
		outputDir := t.TempDir()
		if err := BuildSite(inputDir, outputDir); err != nil {
			t.Errorf("expected no error for valid input/output dirs, got: %v", err)
		}
	})
}

func TestBuildSite_MarkdownAndAssetDiscovery(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	os.WriteFile(filepath.Join(inputDir, "file.md"), []byte("# Title"), 0644)
	os.WriteFile(filepath.Join(inputDir, "file.txt"), []byte("asset"), 0644)
	os.Mkdir(filepath.Join(inputDir, ".hidden"), 0755)
	os.WriteFile(filepath.Join(inputDir, ".hidden", "skip.md"), []byte("# Hidden"), 0644)
	err := BuildSite(inputDir, outputDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "file.html")); err != nil {
		t.Error("expected HTML output for markdown file")
	}
	if _, err := os.Stat(filepath.Join(outputDir, "file.txt")); err != nil {
		t.Error("expected asset file to be copied")
	}
	if _, err := os.Stat(filepath.Join(outputDir, ".hidden", "skip.html")); err == nil {
		t.Error("hidden markdown file should not be processed")
	}
}

func TestBuildSite_AssetCopyError(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	assetPath := filepath.Join(inputDir, "asset.txt")
	os.WriteFile(assetPath, []byte("asset"), 0000) // unreadable
	err := BuildSite(inputDir, outputDir)
	if err == nil {
		t.Error("expected error when asset file is unreadable")
	}
}

func TestBuildSite_MarkdownConversion(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	mdPath := filepath.Join(inputDir, "doc.md")
	os.WriteFile(mdPath, []byte("# Hello World"), 0644)
	err := BuildSite(inputDir, outputDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	htmlPath := filepath.Join(outputDir, "doc.html")
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("expected HTML file, got error: %v", err)
	}
	if !bytes.Contains(html, []byte("Hello World")) {
		t.Error("HTML output missing expected content")
	}
}

func TestBuildSite_MarkdownReadError(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	mdPath := filepath.Join(inputDir, "bad.md")
	os.WriteFile(mdPath, []byte("# Bad"), 0000) // unreadable
	err := BuildSite(inputDir, outputDir)
	if err == nil {
		t.Error("expected error when markdown file is unreadable")
	}
}
