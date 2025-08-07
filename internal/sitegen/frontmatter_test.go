// frontmatter_test.go - Tests for goldmark-frontmatter handling

package sitegen

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func readHTMLMetaField(html string, field string) string {
	// Looks for: <meta name="field" content="...">
	re := regexp.MustCompile(`<meta name="` + regexp.QuoteMeta(field) + `" content="([^"]*)">`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func TestFrontmatterHandling(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")
	os.MkdirAll(inputDir, 0755)

	// Valid YAML frontmatter
	var validYAML = `---
title: Test Post
date: 2025-08-07
tags: [go, test]
---

# Hello
Content here.
`
	var validYAMLUK = `---
title: UK Date Post
date: 07/08/2025
tags: [go, test]
---

# Hello
Content here.
`
	var validYAMLUS = `---
title: US Date Post
date: 08/07/2025
tags: [go, test]
---

# Hello
Content here.
`
	var validYAMLLong = `---
title: Long Date Post
date: 7 August 2025
tags: [go, test]
---

# Hello
Content here.
`
	// Valid TOML frontmatter
	validTOML := `+++
title = "TOML Post"
date = "2025-08-07"
tags = ["go", "toml"]
+++
# Hello TOML
Content here.
`
	// No frontmatter
	noFM := `# No Frontmatter
Just content.
`
	// Invalid frontmatter (bad YAML)
	invalidYAML := `---
title: Bad YAML
date 2025-08-07
tags: [go, test
---
# Broken
Content.
`
	// Invalid frontmatter (bad TOML)
	invalidTOML := `+++
title = "Bad TOML"
date = 2025-08-07
tags = ["go", "toml"
+++
# Broken TOML
Content.
`

	tests := []struct {
		name     string
		content  string
		wantMeta map[string]string
		wantErr  bool
	}{
		{
			name:    "ValidYAML",
			content: validYAML,
			wantMeta: map[string]string{
				"title": "Test Post",
				"date":  "07 Aug 2025",
			},
			wantErr: false,
		},
		{
			name:    "ValidYAMLUK",
			content: validYAMLUK,
			wantMeta: map[string]string{
				"title": "UK Date Post",
				"date":  "07 Aug 2025",
			},
			wantErr: false,
		},
		{
			name:    "ValidYAMLUS",
			content: validYAMLUS,
			wantMeta: map[string]string{
				"title": "US Date Post",
				"date":  "08 Jul 2025",
			},
			wantErr: false,
		},
		{
			name:    "ValidYAMLLong",
			content: validYAMLLong,
			wantMeta: map[string]string{
				"title": "Long Date Post",
				"date":  "07 Aug 2025",
			},
			wantErr: false,
		},
		{
			name:    "ValidTOML",
			content: validTOML,
			wantMeta: map[string]string{
				"title": "TOML Post",
				"date":  "07 Aug 2025",
			},
			wantErr: false,
		},
		{
			name:     "NoFrontmatter",
			content:  noFM,
			wantMeta: map[string]string{},
			wantErr:  false,
		},
		{
			name:     "InvalidYAML",
			content:  invalidYAML,
			wantMeta: map[string]string{},
			wantErr:  false,
		},
		{
			name:     "InvalidTOML",
			content:  invalidTOML,
			wantMeta: map[string]string{},
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdFile := filepath.Join(inputDir, tc.name+".md")
			os.WriteFile(mdFile, []byte(tc.content), 0644)
			sizeOut := make(chan string, 1)
			proc := NewMarkdownProcessor("default")
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = r.(error)
					}
				}()
				err = proc.ProcessMarkdownFile(inputDir, outputDir, tc.name+".md", 1024*1024, sizeOut, nil, nil)
			}()
			htmlFile := filepath.Join(outputDir, tc.name+".html")
			htmlBytes, htmlErr := os.ReadFile(htmlFile)
			html := string(htmlBytes)
			if tc.name == "ValidYAML" {
				t.Logf("[DEBUG] HTML output for ValidYAML:\n%s", html)
			}
			if tc.wantErr {
				if err == nil && htmlErr == nil {
					t.Errorf("expected error for %s, got none", tc.name)
				}
			} else {
				if err != nil || htmlErr != nil {
					t.Errorf("unexpected error for %s: %v %v", tc.name, err, htmlErr)
				}
				for k, v := range tc.wantMeta {
					if v == "" {
						continue
					}
					switch k {
					case "title":
						if !strings.Contains(html, "<h1>"+v+"</h1>") {
							t.Errorf("expected <h1>%s</h1> in HTML for %s", v, tc.name)
						}
					case "date":
						// Only check for date div if a date is expected
						if v != "" {
							dateDivIdx := strings.Index(html, "<div class=\"date\">")
							if dateDivIdx == -1 {
								t.Errorf("expected <div class=\"date\"> in HTML for %s", tc.name)
							} else {
								endIdx := strings.Index(html[dateDivIdx:], "</div>")
								if endIdx == -1 || !strings.Contains(html[dateDivIdx:dateDivIdx+endIdx], v) {
									t.Errorf("expected date string %q inside <div class=\"date\">...</div> for %s", v, tc.name)
								}
							}
						}
					default:
						if !strings.Contains(html, v) {
							t.Errorf("expected meta %s=%s in HTML for %s", k, v, tc.name)
						}
					}
				}
			}
		})
	}
}
