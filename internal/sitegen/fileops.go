// fileops.go - File operation utilities
package sitegen

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed templates/*.html templates/style.css
var EmbeddedFiles embed.FS

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

func parseMarkdownFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// renderHTMLPage is a future-proof extension point for templating support.
func renderHTMLPage(html []byte, templateOpt string, headerHTML, footerHTML []byte) []byte {
	// Determine template path
	var templatePath string
	if templateOpt != "" {
		if _, err := os.Stat(templateOpt); err == nil {
			templatePath = templateOpt
		} else if filepath.IsAbs(templateOpt) || filepath.Ext(templateOpt) == ".html" {
			templatePath = templateOpt
		} else {
			templatePath = "templates/" + templateOpt + ".html"
		}
	} else {
		templatePath = "templates/default.html"
	}
	var tmpl *template.Template
	var err error
	if filepath.IsAbs(templatePath) || fileExists(templatePath) {
		tmpl, err = template.ParseFiles(templatePath)
	} else {
		tmpl, err = template.ParseFS(EmbeddedFiles, templatePath)
	}
	if err != nil {
		return html
	}

	data := struct {
		Content    template.HTML
		Meta       map[string]interface{}
		HeaderHTML template.HTML
		FooterHTML template.HTML
	}{
		Content:    template.HTML(html),
		Meta:       map[string]interface{}{},
		HeaderHTML: template.HTML(headerHTML),
		FooterHTML: template.HTML(footerHTML),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return html
	}
	return buf.Bytes()
}

// SimpleMarkdownToHTML provides a minimal Markdown-to-HTML conversion for headers/footers.
func SimpleMarkdownToHTML(md []byte) []byte {
	// Only handle headings, links, emphasis, and lists for minimalism.
	s := string(md)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	var htmlLines []string
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "# "):
			htmlLines = append(htmlLines, "<h1>"+line[2:]+"</h1>")
		case strings.HasPrefix(line, "## "):
			htmlLines = append(htmlLines, "<h2>"+line[3:]+"</h2>")
		case strings.HasPrefix(line, "- "):
			htmlLines = append(htmlLines, "<li>"+line[2:]+"</li>")
		default:
			htmlLines = append(htmlLines, line)
		}
	}
	html := strings.Join(htmlLines, "\n")
	// Replace [text](url) with <a href="url">text</a>
	html = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(html, `<a href="$2">$1</a>`)
	html = strings.ReplaceAll(html, "*", "<em>")
	html = strings.ReplaceAll(html, "_", "<em>")
	return []byte(html)
}

// fileExists checks if a file exists on disk
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
