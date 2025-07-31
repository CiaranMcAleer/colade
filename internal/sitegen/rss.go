// RSS feed generation
package sitegen

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type RSSGenerator struct {
	baseURL   string
	outputDir string
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language,omitempty"`
	LastBuildDate string `xml:"lastBuildDate,omitempty"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// NewRSSGenerator creates a new RSS generator
func NewRSSGenerator(baseURL, outputDir string) *RSSGenerator {
	return &RSSGenerator{
		baseURL:   baseURL,
		outputDir: outputDir,
	}
}

// Generate creates an RSS feed from the provided markdown files
func (rg *RSSGenerator) Generate(markdownFiles []string, inputDir string, maxItems int) error {
	if rg.baseURL == "" {
		return nil // No RSS generation if base URL is not set
	}

	fmt.Printf("[RSS] Generating RSS feed...\n")

	items, err := rg.collectItems(markdownFiles, inputDir)
	if err != nil {
		return fmt.Errorf("failed to collect RSS items: %w", err)
	}

	if len(items) == 0 {
		fmt.Printf("[RSS] No items found for RSS feed\n")
		return nil
	}

	// Sort by modification time (newest first)
	sort.Slice(items, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC1123Z, items[i].PubDate)
		tj, _ := time.Parse(time.RFC1123Z, items[j].PubDate)
		return ti.After(tj)
	})

	// Use the configurable max items parameter (0 means all items should be included)
	if maxItems > 0 && len(items) > maxItems {
		items = items[:maxItems]
	}

	// Create RSS structure
	rss := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:         rg.inferSiteTitle(inputDir),
			Link:          strings.TrimSuffix(rg.baseURL, "/"),
			Description:   rg.inferSiteDescription(inputDir),
			Language:      "en-gb",
			LastBuildDate: time.Now().Format(time.RFC1123Z),
			Items:         items,
		},
	}

	return rg.writeRSSFile(rss, len(items))
}

// collectItems extracts RSS items from markdown files
func (rg *RSSGenerator) collectItems(markdownFiles []string, inputDir string) ([]Item, error) {
	var items []Item

	for _, relPath := range markdownFiles {
		fullPath := filepath.Join(inputDir, relPath)

		// Read file to extract title and content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("[RSS] Warning: Could not read %s for RSS: %v\n", relPath, err)
			continue // Skip files we can't read
		}

		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		title := rg.extractTitle(string(content), relPath)
		description := rg.extractDescription(string(content), title)
		htmlPath := strings.TrimSuffix(relPath, filepath.Ext(relPath)) + ".html"

		// Ensure proper URL formation
		link := strings.TrimSuffix(rg.baseURL, "/") + "/" + strings.ReplaceAll(htmlPath, "\\", "/")

		items = append(items, Item{
			Title:       title,
			Link:        link,
			Description: description,
			PubDate:     info.ModTime().Format(time.RFC1123Z),
			GUID:        link,
		})
	}

	return items, nil
}

// extractTitle extracts the title from markdown content or falls back to filename
func (rg *RSSGenerator) extractTitle(content, fallback string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Extract title from first heading
			title := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			title = strings.TrimSpace(strings.TrimPrefix(title, "#")) // Handle ## headings too
			title = strings.TrimSpace(strings.TrimPrefix(title, "#")) // Handle ### headings too
			if title != "" {
				return title
			}
		}
	}
	// Fallback to filename without extension, make it more readable
	filename := strings.TrimSuffix(filepath.Base(fallback), filepath.Ext(fallback))
	// Convert kebab-case or snake_case to readable title
	filename = strings.ReplaceAll(filename, "-", " ")
	filename = strings.ReplaceAll(filename, "_", " ")
	return cases.Title(language.Und).String(filename)
}

// extractDescription extracts a description from the content
func (rg *RSSGenerator) extractDescription(content, title string) string {
	lines := strings.Split(content, "\n")
	var description strings.Builder
	foundTitle := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip the title line
		if strings.HasPrefix(line, "#") {
			foundTitle = true
			continue
		}

		// If we found the title, look for the first substantial paragraph
		if foundTitle && line != "" && !strings.HasPrefix(line, "#") {
			// Stop at next heading or after 200 characters
			if description.Len() > 0 && description.Len() < 200 {
				description.WriteString(" ")
			}
			description.WriteString(line)
			if description.Len() >= 200 {
				break
			}
		}
	}

	result := description.String()
	if len(result) > 200 {
		// Truncate at word boundary
		words := strings.Fields(result)
		truncated := ""
		for _, word := range words {
			if len(truncated)+len(word)+1 > 200 {
				break
			}
			if truncated != "" {
				truncated += " "
			}
			truncated += word
		}
		result = truncated + "..."
	}

	// Fallback if no description found
	if result == "" {
		result = title
	}

	return result
}

// inferSiteTitle tries to infer the site title from common patterns
func (rg *RSSGenerator) inferSiteTitle(inputDir string) string {
	// Try to read from index.md or README.md first
	candidates := []string{"index.md", "README.md", "readme.md"}

	for _, candidate := range candidates {
		indexPath := filepath.Join(inputDir, candidate)
		if content, err := os.ReadFile(indexPath); err == nil {
			if title := rg.extractTitle(string(content), candidate); title != "" && title != "Index" && title != "Readme" {
				return title
			}
		}
	}

	// Fallback to directory name
	dirName := filepath.Base(inputDir)
	if dirName == "." || dirName == "/" {
		return "Site Feed"
	}

	// Make directory name more readable
	dirName = strings.ReplaceAll(dirName, "-", " ")
	dirName = strings.ReplaceAll(dirName, "_", " ")
	return cases.Title(language.Und).String(dirName)
}

// inferSiteDescription tries to infer a site description
func (rg *RSSGenerator) inferSiteDescription(inputDir string) string {
	// Try to read description from index.md or README.md
	candidates := []string{"index.md", "README.md", "readme.md"}

	for _, candidate := range candidates {
		indexPath := filepath.Join(inputDir, candidate)
		if content, err := os.ReadFile(indexPath); err == nil {
			title := rg.extractTitle(string(content), candidate)
			if desc := rg.extractDescription(string(content), title); desc != "" && desc != title {
				return desc
			}
		}
	}

	return "Latest posts and updates"
}

// writeRSSFile writes the RSS feed to feed.xml
func (rg *RSSGenerator) writeRSSFile(rss RSS, itemCount int) error {
	rssPath := filepath.Join(rg.outputDir, "feed.xml")
	file, err := os.Create(rssPath)
	if err != nil {
		return fmt.Errorf("error creating RSS file: %w", err)
	}
	defer file.Close()

	// Write XML header and RSS content
	file.WriteString(xml.Header)
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(rss); err != nil {
		return fmt.Errorf("error encoding RSS: %w", err)
	}

	fmt.Printf("[RSS] Generated feed.xml with %d items\n", itemCount)
	return nil
}
