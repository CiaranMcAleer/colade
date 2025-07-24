package sitegen

import (
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
