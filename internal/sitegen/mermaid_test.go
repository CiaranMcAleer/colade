package sitegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSite_MermaidChart(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	mdContent := "# Mermaid Test\n\n" +
		"\x60\x60\x60mermaid\n" +
		"graph TD\n" +
		"  A[Start] --> B{Is it working?}\n" +
		"  B -- Yes --> C[Celebrate!]\n" +
		"  B -- No --> D[Debug]\n" +
		"  D --> B\n" +
		"\x60\x60\x60\n"
	if err := os.WriteFile(filepath.Join(inputDir, "mermaid.md"), []byte(mdContent), 0644); err != nil {
		t.Fatalf("failed to write markdown: %v", err)
	}

	tplPath := "templates/default.html" // Use the default template with mermaid.js
	err := BuildSite(inputDir, outputDir, 0, true, "", 0, false, tplPath, "", "", false, false)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	outFile := filepath.Join(outputDir, "mermaid.html")
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	t.Logf("[DEBUG] Output file content:\n%s", data)

	// Check for mermaid block in output
	if !strings.Contains(string(data), "class=\"mermaid\"") {
		t.Errorf("expected mermaid block in output, got: %s", data)
	}
	if !strings.Contains(string(data), "graph TD") {
		t.Errorf("expected mermaid graph content in output, got: %s", data)
	}
	if !strings.Contains(string(data), "mermaid.min.js") {
		t.Errorf("expected mermaid.js script in output, got: %s", data)
	}
}
