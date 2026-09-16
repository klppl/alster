# Alster

A small Go static site generator for personal projects. Markdown and assets go in
project folders; Alster builds a gallery and a detail page for each write-up.
The output needs only a static file host.

## Run it

Install Go 1.26.3 or newer, then run from this repository:

```sh
go build -o alster ./cmd/alster
./alster build
./alster serve
```

Open http://127.0.0.1:8080. The sample site has two project folders and three
write-ups. `serve` watches content, configuration, and template overrides every
500 ms. Browser pages reload after a successful rebuild; errors appear in the
terminal and the last successful output stays available. Press Ctrl+C to stop.

```sh
./alster build -config path/to/alster.yaml
./alster build -base-url https://username.github.io/reponame
./alster serve -addr 127.0.0.1:9000
go test ./...
```

Dependencies are downloaded when compiling. The resulting binary has no runtime
dependencies: templates, styles, Markdown parsing, and WebP encoding are included.
For a portable Linux binary, use `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
-o alster ./cmd/alster`.

The module path is `github.com/alex/alster` as a scaffold placeholder. When you
publish the generator as a Go module, replace it and the corresponding imports
with your actual repository path. Local builds work as provided.

## Source layout

```text
cmd/alster/main.go           CLI commands and flags
internal/
  config/                   Strict YAML config and URL validation
  markdown/                 YAML frontmatter, GFM, links, picture rendering
  images/                   PNG/JPEG resize and pure-Go WebP encoding
  build/                    Content discovery, gallery model, staged output
  preview/                  Local HTTP server and live reload
  theme/
    theme.go                Embedded theme filesystem
    templates/
      base.html             Shared document head, header, footer
      index.html            Gallery grouped by project folder
      detail.html           Project write-up
    static/
      tokens.css            Default colors, fonts, spacing
      style.css             Responsive layouts and typography
      filter.js             Optional progressive tag filtering
content/                    Sample projects
alster.yaml                 Site settings
.github/workflows/pages.yml Build, test, and deploy on main
```

`tokens.css` at the repository root is a portable copy of the starter tokens.
The embedded theme uses `internal/theme/static/tokens.css`; edit that file and
recompile to change the binary's defaults.

## Content layout and routes

```text
content/
  teslatracker/
    index.md                → dist/teslatracker/index.html
    cover.png               → copied original + optimized variants
  ghost-themes/
    index.md                → dist/ghost-themes/index.html
    fieldnotes.md           → dist/ghost-themes/fieldnotes.html
    cover.png               → shared by either write-up
  scrapers/                 Add your own folders
    example.md
    downloads/example.csv
```

- A top-level folder is a project collection. Every `.md` file gets a detail
  page and a gallery card under that folder. An `index.md` is the optional
  collection overview; it is rendered like any other write-up.
- Folders sort alphabetically; write-ups within a folder sort by date descending,
  then filename. Undated write-ups appear last. This is a showcase, not a blog.
- Nested folders work. Use `[Notes](notes.md#details)` for sibling write-ups and
  `[Other project](../other/index.md)` across folders. Local `.md` links become
  `.html` links, preserving queries and fragments.
- Assets resolve relative to their Markdown file. `/folder/asset.png` refers to
  the content root. All generated internal links are relative, so the same
  output works under a repository subpath or at a domain root.
- Hidden files are omitted. Symlinks and HTML assets are rejected. `_images` and
  `_theme` are reserved folder names. Put content files inside project folders.
- An empty collection renders an empty-state message.

### Frontmatter

```yaml
---
title: My project
description: A short description of what it does.
date: 2026-09-15
tags: [Go, Tools]
status: active
live_url: https://example.com
source_url: https://github.com/example/project
cover: screenshots/cover.png
# template: special.html
---

## Why I built it

Write your project here.
```

All metadata is optional. Without a title, Alster uses the filename (or the
folder name for `index.md`). Dates use `YYYY-MM-DD`; status is `active`, `archived`,
or `wip`. Unknown YAML fields, malformed dates, and unsupported URL schemes fail
with an error. Raw HTML and unsafe links in Markdown are disabled by Goldmark.
Templates are trusted local code.

### Images

PNG and JPEG assets are resized to at most `image_width` pixels wide, without
upscaling. Alster creates a lossless WebP and a resized PNG/JPEG fallback under
`dist/_images/`. Both Markdown images and cover images use `<picture>`. Original
files remain at their original paths for downloads. Inline images load lazily.

[Native WebP](https://github.com/HugoSmits86/nativewebp) provides encoding entirely
in Go; no `cwebp`, C compiler, or ImageMagick is needed at runtime. Lossless WebP
can be larger than JPEG for photographic content. V1 produces one optimized
size, not responsive `srcset` variants. Inputs above 40 megapixels are rejected
to limit decode memory. EXIF orientation is not applied; export photos upright.
SVG, GIF, existing WebP, and other formats are copied unchanged. Remote images
are linked directly, never downloaded or optimized; their availability is not
checked during a build.

## Site settings

Paths are resolved relative to the config file, not the current shell directory.
See `alster.yaml` for the working starter configuration.

```yaml
title: Alster
author: Alex
description: Small tools, themes, and things I make in my spare time.
base_url: https://username.github.io/reponame
accent: "#315d49"
content: content
output: dist
templates: templates
image_width: 1600
nav:
  - label: Projects
    url: /index.html
social:
  - label: GitHub
    url: https://github.com/example
```

`base_url` is optional and controls canonical metadata only. It does not prefix
asset URLs. The build flag overrides it, which is useful for deployment.
Use `/`-prefixed site links in nav/social settings so they resolve from every
page. `accent` accepts a six-digit hex color and generates `--color-accent` in a
small stylesheet. The theme uses local font stacks and makes no font requests.

## Override templates

Create a `templates/` folder next to `alster.yaml`. Alster parses every `.html`
file there after loading the embedded defaults. Copy a default template and edit
it, or replace a named block:

```html
{{define "footer"}}
<footer class="site-footer">{{.Site.Author}} · Personal projects</footer>
</body></html>
{{end}}
```

Override `index.html` or `detail.html` to change page layouts. To style a specific
write-up, create `templates/special.html` and set `template: special.html` in its
frontmatter. Reuse the shared blocks:

```html
{{template "head" .}}
<main id="main" class="detail">
  <h1>{{.Page.Title}}</h1>
  <article class="prose">{{.Page.HTML}}</article>
</main>
{{template "footer" .}}
```

Templates receive `.Site`, `.Prefix`, `.Title`, `.Description`, `.Canonical`, and
`.Page` on detail pages, or `.Groups` and `.Tags` on the gallery. `.Page` includes
frontmatter, `.Group`, `.URL`, `.CoverURL`, `.CoverImage`, and rendered `.HTML`.
Helpers: `link URL prefix` resolves site-root URLs; `asset path prefix` handles
local asset paths or remote URLs; `pathURL path` escapes filesystem names;
`join strings separator` joins a list. Use these helpers for portable links.
For custom CSS, override `head` and link to an asset in a content project folder.
Editing embedded defaults requires recompiling; local overrides reload in `serve`.

## Build safety

Alster stages a complete build before replacing `dist`. Existing output must
contain Alster's `.alster-output` marker; otherwise it refuses to overwrite it.
Output cannot overlap content or templates. Only generated output should live
in that directory: a successful rebuild removes stale files. Do not run two
builds against the same output directory concurrently.

## GitHub Pages

The included workflow tests the code, compiles with CGO disabled, builds `dist`,
and deploys it on push to `main` or manual dispatch. It obtains the canonical base
URL from `actions/configure-pages`, including the repository subpath.

1. Push this repository to GitHub.
2. In **Settings → Pages → Build and deployment**, select **GitHub Actions**.
3. Push to `main` or run **Deploy portfolio to GitHub Pages** manually.

The workflow follows GitHub's [custom Pages workflow](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages).
No credentials belong in the config. GitHub provides a scoped deployment token.

## Cloudflare Pages

Deploy the same `dist/` as a static directory. With a Git-connected Pages project,
use a build environment with the Go version in `go.mod`, a build command of
`go run ./cmd/alster build`, and output directory `dist`. Alternatively, compile
and build locally, then upload `dist` through Pages Direct Upload. Set `base_url`
to your public URL if you want canonical tags. No functions, rewrites, or SPA
fallback are required. See Cloudflare’s [static HTML guide](https://developers.cloudflare.com/pages/framework-guides/deploy-anything/).

## V1 boundaries

The preview server is a local development tool. Production output has no server
or live-reload script. Builds do not contact project URLs or remote images.
Local Markdown links are rewritten but are not comprehensively checked for broken
targets or missing fragments. RSS/Atom, image caching, custom tag pages, and
incremental builds are left for later; the content model and theme are separate
so these can be added without changing project folders.
