package sitegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRSSGenerator_Generate(t *testing.T) {
	// Create temporary directories
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create test markdown files
	indexContent := `# My Awesome Blog
This is my personal blog.

Welcome to my corner of the internet!`

	post1Content := `# First Post
This is my first blog post about getting started with static site generators.

I've been exploring different tools and colade seems really fast and simple.`

	post2Content := `# Learning Go
Today I learned about Go's concurrency features.

Goroutines and channels make concurrent programming much easier than in other languages.`

	err := os.WriteFile(filepath.Join(inputDir, "index.md"), []byte(indexContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create index.md: %v", err)
	}

	err = os.WriteFile(filepath.Join(inputDir, "post1.md"), []byte(post1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create post1.md: %v", err)
	}

	err = os.WriteFile(filepath.Join(inputDir, "post2.md"), []byte(post2Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create post2.md: %v", err)
	}

	// Test RSS generation
	rss := NewRSSGenerator("https://example.com", outputDir)
	markdownFiles := []string{"index.md", "post1.md", "post2.md"}

	err = rss.Generate(markdownFiles, inputDir, 20)
	if err != nil {
		t.Fatalf("RSS generation failed: %v", err)
	}

	// Verify feed.xml was created
	feedPath := filepath.Join(outputDir, "feed.xml")
	if _, err := os.Stat(feedPath); err != nil {
		t.Error("feed.xml was not created")
	}

	// Read and verify feed content
	content, err := os.ReadFile(feedPath)
	if err != nil {
		t.Fatalf("Could not read feed.xml: %v", err)
	}

	feedContent := string(content)

	// Check for RSS structure
	if !strings.Contains(feedContent, `<rss version="2.0">`) {
		t.Error("RSS version not found")
	}

	if !strings.Contains(feedContent, `<channel>`) {
		t.Error("Channel element not found")
	}

	// Check for site title (should be inferred from index.md)
	if !strings.Contains(feedContent, `<title>My Awesome Blog</title>`) {
		t.Error("Site title not correctly inferred")
	}

	// Check for base URL
	if !strings.Contains(feedContent, `<link>https://example.com</link>`) {
		t.Error("Site link not found")
	}

	// Check for items
	if !strings.Contains(feedContent, `<item>`) {
		t.Error("No RSS items found")
	}

	// Check for post titles
	if !strings.Contains(feedContent, "First Post") {
		t.Error("Post title not found in RSS")
	}

	if !strings.Contains(feedContent, "Learning Go") {
		t.Error("Post title not found in RSS")
	}

	// Check for proper URLs
	if !strings.Contains(feedContent, "https://example.com/post1.html") {
		t.Error("Post URL not correctly generated")
	}
}

func TestRSSGenerator_ExtractTitle(t *testing.T) {
	rss := NewRSSGenerator("", "")

	tests := []struct {
		content  string
		fallback string
		expected string
	}{
		{"# Hello World\nContent", "test.md", "Hello World"},
		{"## Secondary Heading\nContent", "test.md", "Secondary Heading"},
		{"### Tertiary Heading\nContent", "test.md", "Tertiary Heading"},
		{"No title here", "test.md", "Test"},
		{"", "my-post.md", "My Post"},
		{"", "complex_file-name.md", "Complex File Name"},
	}

	for _, test := range tests {
		result := rss.extractTitle(test.content, test.fallback)
		if result != test.expected {
			t.Errorf("extractTitle(%q, %q) = %q, want %q", test.content, test.fallback, result, test.expected)
		}
	}
}

func TestRSSGenerator_ExtractDescription(t *testing.T) {
	rss := NewRSSGenerator("", "")

	tests := []struct {
		content     string
		title       string
		description string
	}{
		{
			"# Title\nThis is a description",
			"Title",
			"This is a description",
		},
		{
			"# Title\n\nThis is a longer description that should be included in the RSS feed.",
			"Title",
			"This is a longer description that should be included in the RSS feed.",
		},
		{
			"# Title\nNo content",
			"Title",
			"No content", // Will extract "No content" as description
		},
	}

	for _, test := range tests {
		result := rss.extractDescription(test.content, test.title)
		if result != test.description {
			t.Errorf("extractDescription returned %q, want %q", result, test.description)
		}
	}
}

func TestRSSGenerator_MaxItemsConfiguration(t *testing.T) {
	// Create temporary directories
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create multiple test markdown files (more than the max we'll test)
	testPosts := []struct {
		filename string
		title    string
		content  string
	}{
		{"post1.md", "First Post", "# First Post\nThis is the first post."},
		{"post2.md", "Second Post", "# Second Post\nThis is the second post."},
		{"post3.md", "Third Post", "# Third Post\nThis is the third post."},
		{"post4.md", "Fourth Post", "# Fourth Post\nThis is the fourth post."},
		{"post5.md", "Fifth Post", "# Fifth Post\nThis is the fifth post."},
	}

	var markdownFiles []string
	for _, post := range testPosts {
		err := os.WriteFile(filepath.Join(inputDir, post.filename), []byte(post.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create %s: %v", post.filename, err)
		}
		markdownFiles = append(markdownFiles, post.filename)
	}

	// Test with maxItems = 3
	rss := NewRSSGenerator("https://example.com", outputDir)
	err := rss.Generate(markdownFiles, inputDir, 3)
	if err != nil {
		t.Fatalf("RSS generation failed: %v", err)
	}

	// Verify feed.xml was created
	feedPath := filepath.Join(outputDir, "feed.xml")
	content, err := os.ReadFile(feedPath)
	if err != nil {
		t.Fatalf("Could not read feed.xml: %v", err)
	}

	feedContent := string(content)

	// Count the number of <item> elements
	itemCount := strings.Count(feedContent, "<item>")
	if itemCount != 3 {
		t.Errorf("Expected 3 items in RSS feed, got %d", itemCount)
	}

	// Test with maxItems = 0 (should include all items since 0 means no limit)
	err = rss.Generate(markdownFiles, inputDir, 0)
	if err != nil {
		t.Fatalf("RSS generation failed: %v", err)
	}

	content, err = os.ReadFile(feedPath)
	if err != nil {
		t.Fatalf("Could not read feed.xml: %v", err)
	}

	feedContent = string(content)
	itemCount = strings.Count(feedContent, "<item>")
	if itemCount != 5 {
		t.Errorf("Expected 5 items in RSS feed when maxItems=0 (no limit), got %d", itemCount)
	}

	// Test with maxItems = 10 (more than available items)
	err = rss.Generate(markdownFiles, inputDir, 10)
	if err != nil {
		t.Fatalf("RSS generation failed: %v", err)
	}

	content, err = os.ReadFile(feedPath)
	if err != nil {
		t.Fatalf("Could not read feed.xml: %v", err)
	}

	feedContent = string(content)
	itemCount = strings.Count(feedContent, "<item>")
	if itemCount != 5 {
		t.Errorf("Expected 5 items in RSS feed when maxItems > available items, got %d", itemCount)
	}
}

func TestRSSGenerator_DefaultMaxItems(t *testing.T) {
	// Create temporary directories
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create more than 20 test markdown files to test default behavior
	var markdownFiles []string
	for i := 1; i <= 25; i++ {
		filename := fmt.Sprintf("post%d.md", i)
		content := fmt.Sprintf("# Post %d\nThis is post number %d.", i, i)
		err := os.WriteFile(filepath.Join(inputDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
		markdownFiles = append(markdownFiles, filename)
	}

	// Test with default maxItems = 20
	rss := NewRSSGenerator("https://example.com", outputDir)
	err := rss.Generate(markdownFiles, inputDir, 20)
	if err != nil {
		t.Fatalf("RSS generation failed: %v", err)
	}

	// Verify feed.xml was created
	feedPath := filepath.Join(outputDir, "feed.xml")
	content, err := os.ReadFile(feedPath)
	if err != nil {
		t.Fatalf("Could not read feed.xml: %v", err)
	}

	feedContent := string(content)

	// Count the number of <item> elements - should be limited to 20
	itemCount := strings.Count(feedContent, "<item>")
	if itemCount != 20 {
		t.Errorf("Expected 20 items in RSS feed (default limit), got %d", itemCount)
	}
}
