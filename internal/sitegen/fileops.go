// fileops.go - File operation utilities
package sitegen

import (
	"io"
	"os"
	"path/filepath"
)

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

// parseMarkdownFile is a future-proof extension point for frontmatter support.
func parseMarkdownFile(path string) ([]byte, error) {
	//TODO when implementing frontmatter, this will need to parse the file
	return os.ReadFile(path)
}

// renderHTMLPage is a future-proof extension point for templating support.
func renderHTMLPage(html []byte) []byte {
	//TODO when implementing templating, this will need to render the HTML
	// For now, just return the HTML as-is.
	return html
}
