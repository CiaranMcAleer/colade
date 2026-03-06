# colade
Colade is a blog/static site generator written in go, I designed this to learn more about the go language, it is not meant to be a full featured static site generator like Hugo or Jekyll.

Features:

- Fast(like really fast)
- No lock in, uses standard markdown files, you can use any markdown editor you like and generate a website with colade.
- Simple and easy to use CLI
- Aims to keep page size small(<14kb compressed) based on <https://news.ycombinator.com/item?id=44613625>
- Incremental builds to speed up build process for large sites
- RSS feed generation

## Adding Headers and Footers

To add a header or footer to your site, create `header.md` and/or `footer.md` files in your input directory. These files will be converted to HTML and included at the top and bottom of every generated page. You can also use cli flags to specify custom header and footer files:

## Using YAML Frontmatter

You can add YAML frontmatter to your markdown files to specify metadata such as title, date, and tags. Example:

```markdown
---
title: My Post Title
date: 2025-08-07
tags: [go, static-site]
---
# My Post

Content goes here.
```

Supported date formats include `yyyy-mm-dd`, `dd/mm/yyyy`, `mm/dd/yyyy`, and long-form dates like `7 August 2025`.

## Building Templates and Styling Your Site

.Colade provides a straightforward template system allowing you to create custom HTML for your site layout, control CSS, and tailor site appearance to your needs.

### 1. Template Directory Structure
- Create a directory called `templates/` (or use an existing one) in your project root.
- Place HTML template files inside: e.g., `default.html`, `minimal.html`, `dark.html`.

### 2. Template Design and Variables
Templates are standard Go HTML templates. You control every aspect of the layout. Available template variables:

- `.Content`: Rendered HTML from your markdown file
- `.Title`: Title from frontmatter
- `.Date`: Date from frontmatter (normalized)
- `.Tags`: Tags from frontmatter (as a list)
- `.HeaderHTML` / `.FooterHTML`: HTML for header/footer (from header.md/footer.md)
- `.Meta`: Full parsed frontmatter as a map

**Example template (`default.html`):**

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>{{ .Title }}</title>
  <link rel="stylesheet" href="/style.css">
</head>
<body>
  {{ .HeaderHTML }}
  <h1>{{ .Title }}</h1>
  {{ if .Date }}<div class="date">{{ .Date }}</div>{{ end }}
  {{ if .Tags }}<div class="tags">{{ range .Tags }}<span class="tag">{{ . }}</span>{{ end }}</div>{{ end }}
  {{ .Content }}
  {{ .FooterHTML }}
</body>
</html>
```

You may use Go template logic (if, range, index) to customize output. E.g., `{{ if .Tags }}` for conditional tags, or `{{ index .Meta "title" }}` for advanced frontmatter access.

### 3. Selecting a Template
- By default, colade uses `default.html`.
- To choose a different template, pass `--template <name>` (without extension) when building: e.g.,
  ```
  colade build input/ output/ --template dark
  ```
- Or set a template in your markdown frontmatter:
  ```yaml
  ---
  title: My Post
  template: minimal
  ---
  ```

### 4. Customizing CSS and Styles
- Default styling is defined in `style.css` (see sample below).
- To use your own stylesheet, pass `--css path/to/your.css` when building. Colade will copy it to the output directory as `style.css` automatically.


**Sample `style.css`:**
```css
body {
  font-family: system-ui, sans-serif;
  margin: 2rem auto;
  max-width: 700px;
  background: #f9f9f9;
  color: #222;
}
h1, h2, h3, h4, h5, h6 {
  color: #2a5d9f;
}
a {
  color: #2a5d9f;
  text-decoration: none;
}
a:hover {
  text-decoration: underline;
}
code, pre {
  background: #f4f4f4;
  border-radius: 4px;
  padding: 2px 4px;
  font-family: "Fira Mono", monospace;
}
```


You can freely extend or replace this stylesheet to change colors, layout, spacing, or add new classes. Any template referring to `/style.css` will receive your styling.


### 5. Header and Footer
To add a site-wide header and footer, create `header.md` and `footer.md` in your input directory. These will be converted to HTML and injected as `.HeaderHTML` and `.FooterHTML`.

**Example header.md/footer.md usage:**
```markdown
<!-- Portfolio Header -->
```


### 6. Workflow: Building Your Site with Custom Template and CSS
1. Create your markdown pages in your input directory (`input/` or similar), using YAML frontmatter for metadata:
   ```yaml
   ---
   title: My Portfolio
   date: 2026-03-06
   tags: [portfolio, ai]
   template: dark
   ---
   # Welcome to My Portfolio
   ```
2. Design your templates in `templates/` (edit `default.html`, add new HTML files as needed).
3. Make your `style.css` (or use default, or copy/extend sample above).
4. Build your site with colade:
   ```
   colade build input/ output/ --template dark --css templates/style.css
   ```
5. Grab your generated site from the output directory—ready to deploy.
6. Iterate and adjust as needed: change templates, tweak CSS, rebuild.

### 7. Download Latest Release and Start Working
- Download the latest colade release from Github.
- Create your markdown, templates, and CSS.
- Build your site with one command—no complex setup, no lock-in.

**Quick-start:**
```sh
colade build input/ output/ --template default --css templates/style.css
```

### Troubleshooting
If pages don't look as expected:
- Double-check template file names and YAML frontmatter.
- Ensure the template is referenced correctly (without `.html`).
- Confirm your CSS is copied as `style.css` in the output.
- Review build flags (`--template`, `--css`) for typos.
- Look for errors/warnings in the colade build output.

---
This expanded guide should allow anyone to build, style, and launch a Colade-powered site quickly.
### 8. Interlinking Pages and Site Navigation
Colade generates HTML from your Markdown files, but site-wide navigation and interlinking are managed inside your templates. Here’s how to link pages, create blog indexes, and build basic navigation:

- **Linking between pages:** Use standard HTML `<a href="">` tags. For example, in `index.md`, link to a post:
  ```markdown
  [Read my post](post.html)
  ```

- **Dynamic navigation/template lists:** While Colade does not auto-create a page list, you can manually build an index (`blog.md` or `index.md`) that links to each post.
  ```markdown
  - [Post One](post-one.html)
  - [Post Two](post-two.html)
  ```

- **Displaying metadata in templates:** Use template variables to show post metadata:
  ```html
  <h1>{{ .Title }}</h1>
  <div class="date">{{ .Date }}</div>
  <div class="tags">{{ range .Tags }}<span class="tag">{{ . }}</span>{{ end }}</div>
  ```

- **Creating a tag index or categories:** Add a tag page manually (e.g., `tags.md`), or use Go template logic to reference tags. You can link all pages with a certain tag yourself.

- **Navigation bar example (in header.md, footer.md, or template):**
  ```html
  <nav>
    <a href="index.html">Home</a> |
    <a href="blog.html">Blog</a> |
    <a href="about.html">About</a>
  </nav>
  ```

**Note:** All links are relative to your site’s output structure. The simplest workflow is:
1. Write your Markdown files and assemble manual lists (or copy/paste links).
2. Reference the intended output filenames (e.g., `about.html`, `post.html`).
3. Use template variables for metadata, styling, and per-page titles.

For advanced link generation between posts, custom scripting is possible – but for most users, manual lists and simple navigation suffice for robust static sites.

## Incremental Build Cache Format

`.colade-cache` (JSON example):

```json
{
  "version": 1,// To ensure compatibility with future versions
  "files": {
    "index.md": {// file name and path relative to the input directory
      "mtime": 1722172800,// last modified time in Unix timestamp
      "output": "index.html"// relative path in the output directory
    },
    "test.md": {
      "mtime": 1722172801,
      "output": "test.html"
    },
    "assets/logo.png": {
      "mtime": 1722172802,
      "output": "assets/logo.png"
    }
  }
}
```

## Incremental Build Usage

By default, Colade uses incremental builds to speed up site generation. Only changed, added, or deleted files are processed.

- To force a full rebuild (disable incremental), use the `--no-incremental` flag:

```
colade build input/ output/ --no-incremental
```

- The build system maintains a `.colade-cache` file in the output directory to track file changes as part of the incremental build process.

## Benchmark Coalde Performance
```go
go test -bench=BenchmarkSitegenPerformance ./internal/sitegen/
```
