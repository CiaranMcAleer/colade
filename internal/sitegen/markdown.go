// markdown.go - Markdown processing utilities
package sitegen

import (
	"regexp"
	"strings"
)

// replaceMdLinks replaces links to .md/.markdown files with .html in markdown content.
func replaceMdLinks(content []byte) []byte {
	s := string(content)
	// Replace [text](foo.md) and [text](foo.markdown) with [text](foo.html)
	// Use regex to specifically target markdown link syntax
	re := regexp.MustCompile(`\[([^\]]*)\]\(([^)]*\.(?:md|markdown))\)`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		return strings.ReplaceAll(strings.ReplaceAll(match, ".md)", ".html)"), ".markdown)", ".html)")
	})
	return []byte(s)
}
