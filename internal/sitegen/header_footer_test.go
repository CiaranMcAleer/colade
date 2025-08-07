// header_footer_test.go - Tests for header and footer injection

package sitegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHeaderFooterInjection(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")
	os.MkdirAll(inputDir, 0755)

	// Write header.md, footer.md, and a content file
	headerContent := "# Site Header\n\n[Home](/)\n"
	footerContent := "_Footer text_"
	pageContent := "# Hello World\n\nThis is the main content."

	os.WriteFile(filepath.Join(inputDir, "header.md"), []byte(headerContent), 0644)
	os.WriteFile(filepath.Join(inputDir, "footer.md"), []byte(footerContent), 0644)
	os.WriteFile(filepath.Join(inputDir, "index.md"), []byte(pageContent), 0644)

	// Use default template and build site
	err := BuildSite(inputDir, outputDir, 14*1024, false, "", 20, false, "default", "", "", false, false)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	// Read generated HTML
	htmlBytes, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatalf("Failed to read generated HTML: %v", err)
	}
	html := string(htmlBytes)

	// Check header and footer content present
	if !strings.Contains(html, "Site Header") || !strings.Contains(html, "Home") {
		t.Errorf("Header content not found in output HTML")
	}
	if !strings.Contains(html, "Footer text") {
		t.Errorf("Footer content not found in output HTML")
	}
}

func TestHeaderFooterMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")
	os.MkdirAll(inputDir, 0755)

	// Only write content file, no header.md/footer.md
	pageContent := "# Hello World\n\nThis is the main content."
	os.WriteFile(filepath.Join(inputDir, "index.md"), []byte(pageContent), 0644)

	err := BuildSite(inputDir, outputDir, 14*1024, false, "", 20, false, "default", "", "", false, false)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	htmlBytes, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatalf("Failed to read generated HTML: %v", err)
	}
	html := string(htmlBytes)

	// Should not contain header/footer content
	if strings.Contains(html, "Site Header") || strings.Contains(html, "Footer text") {
		t.Errorf("Unexpected header/footer content found in output HTML")
	}
}

func TestHeaderFooterInvalidMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")
	os.MkdirAll(inputDir, 0755)

	// Write invalid markdown to header/footer
	os.WriteFile(filepath.Join(inputDir, "header.md"), []byte("<<<<<"), 0644)
	os.WriteFile(filepath.Join(inputDir, "footer.md"), []byte(">>>>>"), 0644)
	os.WriteFile(filepath.Join(inputDir, "index.md"), []byte("# Main"), 0644)

	err := BuildSite(inputDir, outputDir, 14*1024, false, "", 20, false, "default", "", "", false, false)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	htmlBytes, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatalf("Failed to read generated HTML: %v", err)
	}
	html := string(htmlBytes)

	// Should still include the invalid content as raw HTML
	if !strings.Contains(html, "<<<<<") && !strings.Contains(html, ">>>>>") {
		t.Errorf("Invalid markdown in header/footer not rendered as HTML")
	}
}

func TestHeaderFooterDisabledFlags(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")
	os.MkdirAll(inputDir, 0755)

	// Write header.md, footer.md, and a content file
	os.WriteFile(filepath.Join(inputDir, "header.md"), []byte("# Header"), 0644)
	os.WriteFile(filepath.Join(inputDir, "footer.md"), []byte("Footer!"), 0644)
	os.WriteFile(filepath.Join(inputDir, "index.md"), []byte("# Main"), 0644)

	// Build with header/footer disabled
	err := BuildSite(inputDir, outputDir, 14*1024, false, "", 20, false, "default", "", "", true, true)
	if err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	htmlBytes, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatalf("Failed to read generated HTML: %v", err)
	}
	html := string(htmlBytes)

	if strings.Contains(html, "Header") || strings.Contains(html, "Footer!") {
		t.Errorf("Header/footer should not be present when disabled by flags")
	}
}
