// fileops.go - File operation utilities
package sitegen

import (
	"bytes"
	"html/template"
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
func renderHTMLPage(html []byte, templateOpt string) []byte {
	// Determine template path
	var templatePath string
	// Use templateOpt directly if it exists as a file (absolute or relative)
	if templateOpt != "" {
		if _, err := os.Stat(templateOpt); err == nil {
			templatePath = templateOpt
		} else if filepath.IsAbs(templateOpt) || filepath.Ext(templateOpt) == ".html" {
			templatePath = templateOpt
		} else {
			templatePath = filepath.Join("templates", templateOpt+".html")
		}
	} else {
		templatePath = filepath.Join("templates", "default.html")
	}
	// Fallback: if template doesn't exist, use default
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templatePath = filepath.Join("templates", "default.html")
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		// fallback to raw HTML if template fails
		return html
	}

	data := struct {
		Content template.HTML
		Meta    map[string]interface{}
	}{
		Content: template.HTML(html),
		Meta:    map[string]interface{}{}, // TODO: pass real meta if available
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return html
	}
	return buf.Bytes()
}
