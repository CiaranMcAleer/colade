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

## Custom Templates

You can define custom HTML templates in the `templates/` directory. To use a custom template, specify its name (without extension) in your build command or frontmatter.

- Template variables available:
  - `.Content`: Rendered HTML content of the markdown file
  - `.Title`: Title from frontmatter
  - `.Date`: Date from frontmatter (normalized)
  - `.Tags`: Tags from frontmatter (as a list)
  - `.HeaderHTML` / `.FooterHTML`: Rendered header/footer HTML
  - `.Meta`: Full frontmatter as a map

Example usage in a template:

```html
<!DOCTYPE html>
<html>
<head>
  <title>{{ .Title }}</title>
</head>
<body>
  {{ .HeaderHTML }}
  <h1>{{ .Title }}</h1>
  {{ if .Date }}<div class="date">{{ .Date }}</div>{{ end }}
  {{ .Content }}
  {{ .FooterHTML }}
</body>
</html>
```

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
