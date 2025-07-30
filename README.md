     This repo is under active development, most of the features are implemented, but there are still some things to do before the first release.

# colade
Colade is a blog/static site generator written in go, I designed this to learn more about the go language, it is not meant to be a full featured static site generator like Hugo or Jekyll.

Features:

- Fast(like really fast)
- No lock in, uses standard markdown files, you can use any markdown editor you like and generate a website with colade.
- Simple and easy to use CLI
- Aims to keep page size small(<14kb compressed) based on <https://news.ycombinator.com/item?id=44613625>
- Incremental builds to speed up build process for large sites
- RSS feed generation

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
