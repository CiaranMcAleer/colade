package sitegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSite_WithDifferentTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	// Setup input directory and markdown file
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	mdContent := "# Hello World\nThis is a test post."
	if err := os.WriteFile(filepath.Join(inputDir, "test.md"), []byte(mdContent), 0644); err != nil {
		t.Fatalf("failed to write markdown: %v", err)
	}

	// Setup two templates
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}
	defaultTpl := "<html><body><h1>{{ .Title }}</h1>{{ .Content }}</body></html>"
	minimalTpl := "<html><main>{{ .Content }}</main></html>"
	if err := os.WriteFile(filepath.Join(templatesDir, "default.html"), []byte(defaultTpl), 0644); err != nil {
		t.Fatalf("failed to write default template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templatesDir, "minimal.html"), []byte(minimalTpl), 0644); err != nil {
		t.Fatalf("failed to write minimal template: %v", err)
	}

	t.Run("DefaultTemplate", func(t *testing.T) {
		tplOpt := filepath.Join(templatesDir, "default.html")
		t.Logf("[DEBUG] Using template: %s", tplOpt)
		if _, err := os.Stat(tplOpt); err != nil {
			t.Fatalf("[DEBUG] Template file missing: %v", err)
		}
		err := BuildSite(inputDir, outputDir, 0, true, "", 0, false, tplOpt, "", "", false, false, "")
		if err != nil {
			t.Fatalf("BuildSite failed: %v", err)
		}
		outFile := filepath.Join(outputDir, "test.html")
		data, err := os.ReadFile(outFile)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}
		t.Logf("[DEBUG] Output file content:\n%s", data)
		if string(data) == "" || !strings.Contains(string(data), "Hello World") {
			t.Errorf("output missing expected content: %s", data)
		}
		if !strings.Contains(string(data), "<h1>") {
			t.Errorf("default template not applied: %s", data)
		}
	})

	t.Run("MinimalTemplate", func(t *testing.T) {
		tplOpt := filepath.Join(templatesDir, "minimal.html")
		t.Logf("[DEBUG] Using template: %s", tplOpt)
		if _, err := os.Stat(tplOpt); err != nil {
			t.Fatalf("[DEBUG] Template file missing: %v", err)
		}
		err := BuildSite(inputDir, outputDir, 0, true, "", 0, false, tplOpt, "", "", false, false, "")
		if err != nil {
			t.Fatalf("BuildSite failed: %v", err)
		}
		outFile := filepath.Join(outputDir, "test.html")
		data, err := os.ReadFile(outFile)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}
		t.Logf("[DEBUG] Output file content:\n%s", data)
		if !strings.Contains(string(data), "<main>") {
			t.Errorf("minimal template not applied: %s", data)
		}
	})
}
