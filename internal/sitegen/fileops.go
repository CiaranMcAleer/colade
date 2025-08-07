// fileops.go - File operation utilities
package sitegen

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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
func renderHTMLPage(html []byte, templateOpt string, headerHTML, footerHTML []byte, meta map[string]interface{}) []byte {
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

	// Flatten common meta fields for easier template access
	var title, date string
	var tags []interface{}
	if meta != nil {
		if v, ok := meta["title"].(string); ok {
			title = v
		}
		// Accept date as string or time.Time
		switch v := meta["date"].(type) {
		case string:
			fmt.Printf("[DEBUG] meta[\"date\"] = %q\n", v)
			fmt.Printf("[DEBUG] meta = %#v\n", meta)
			formats := []string{
				"2006-01-02",      // ISO
				"02/01/2006",      // UK/EU
				"01/02/2006",      // US
				"02 Jan 2006",     // 07 Aug 2025
				"2 January 2006",  // 7 August 2025
				"January 2, 2006", // August 7, 2025
			}
			var parsed time.Time
			for _, f := range formats {
				t, err := time.Parse(f, v)
				if err == nil {
					fmt.Printf("[DEBUG] Parsed date %q with format %q\n", v, f)
					parsed = t
					break
				}
			}
			if !parsed.IsZero() {
				date = parsed.Format("02 Jan 2006")
			} else {
				fmt.Printf("[DEBUG] Could not parse date %q, using as-is\n", v)
				date = v // fallback to original
			}
		case time.Time:
			fmt.Printf("[DEBUG] meta[\"date\"] is time.Time: %v\n", v)
			date = v.Format("02 Jan 2006")
		}
		if v, ok := meta["tags"].([]interface{}); ok {
			tags = v
		}
	}

	data := struct {
		Content    template.HTML
		Meta       map[string]interface{}
		HeaderHTML template.HTML
		FooterHTML template.HTML
		Title      string
		Date       string
		Tags       []interface{}
	}{
		Content:    template.HTML(html),
		Meta:       meta,
		HeaderHTML: template.HTML(headerHTML),
		FooterHTML: template.HTML(footerHTML),
		Title:      title,
		Date:       date,
		Tags:       tags,
	}

	// DEBUG: Print the final date value passed to the template
	fmt.Printf("[DEBUG] FINAL data.Date = %q for title %q\n", data.Date, data.Title)

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
