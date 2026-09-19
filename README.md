# Alster

A minimalist, zero-dependency static site generator and Scandinavian portfolio theme written in Go. Compiles Markdown and assets into self-contained, high-performance static HTML without frontend frameworks.

## Requirements

- **Go**: 1.24+ (tested on Go 1.26+)
- **Runtime dependencies**: None (templates, styles, Markdown parsing, and WebP encoding are embedded in the binary)

## Quick Start

```sh
# Build binary
go build -o alster ./cmd/alster

# Build static output to dist/
./alster build

# Run local preview server with live reload (http://127.0.0.1:8080)
./alster serve

# Run test suite
go test ./...
```

Flags:
- `./alster build -config path/to/alster.yaml` — use custom config file
- `./alster build -base-url https://example.com` — override canonical base URL
- `./alster serve -addr 127.0.0.1:9000` — bind preview server to custom address

## Content Structure

```text
content/
├── index.md                 # Homepage bio and intro text
├── about/
│   ├── index.md             # About page bio, principles, colophon
│   └── profile.jpg          # Profile portrait (auto-converted to WebP)
├── blog/
│   ├── building-alster/
│   │   └── index.md         # Blog post: dist/blog/building-alster/index.html
│   └── on-minimalism/
│       └── index.md         # Blog post: dist/blog/on-minimalism/index.html
├── teslatracker/
│   ├── index.md             # Project page: dist/teslatracker/index.html
│   └── cover.png            # Project card cover (auto-converted to WebP)
└── ghost-themes/
    ├── index.md             # Project overview: dist/ghost-themes/index.html
    └── fieldnotes.md        # Project sub-article: dist/ghost-themes/fieldnotes.html
```

## Adding Content

### Add a Project Entry

1. Create a directory under `content/<project-name>/` (e.g., `content/audio-dsp/`).
2. Add a cover image (e.g. `cover.png`).
3. Create `content/<project-name>/index.md` with frontmatter:

```yaml
---
title: AudioDSP
description: Low-latency audio filtering utility built in Go.
date: 2026-09-15
tags: [Go, Audio, DSP]
status: active
cover: cover.png
live_url: https://audiodsp.example.com
source_url: https://github.com/alex/audiodsp
---

Write project documentation and notes here in standard Markdown.
```

- **Status values**: `active` (Aktiv), `wip` (Pågående), `archived` (Arkiverad).
- **Images**: Local images referenced in Markdown or `cover` are automatically resized and converted to lossless WebP with fallbacks.

### Add a Blog / Writing Entry

1. Create a directory under `content/blog/<slug>/` (e.g., `content/blog/static-first/`).
2. Create `content/blog/<slug>/index.md` with frontmatter:

```yaml
---
title: The Case for Static Software
description: Why self-contained files outlive complex database-driven backends.
date: 2026-09-18
tags: [Filosofi, Webb]
---

Write article body here. Subheadings (###) and lists format cleanly into the editorial layout.
```

- Tag filtering on `/blog/` updates automatically when new tags are introduced.

## Configuration

Site settings, navigation, and author details live in `alster.yaml`:

```yaml
title: Alster
author: Alex
description: Små verktyg, teman och saker jag bygger på fritiden.
lang: sv                   # Language code: sv or en
base_url: ""               # Canonical URL (e.g. https://alex.github.io/alster)
accent: "#315d49"          # Accent hex color
content: content           # Content directory
output: dist               # Target static build directory
templates: templates       # Optional local template overrides
image_width: 1600          # Maximum image downscale width

nav:
  - label: Projekt
    url: /projects/index.html
  - label: Skrivande
    url: /blog/index.html
  - label: Om
    url: /about/index.html

social:
  - label: E-post
    url: mailto:alex@example.com
  - label: GitHub
    url: https://github.com/alex
  - label: Bluesky
    url: https://bsky.app/profile/alex.bsky.social
  - label: LinkedIn
    url: https://linkedin.com/in/alex
  - label: X
    url: https://x.com/alex
```

- Set `lang: sv` for Swedish UI strings or `lang: en` for English.
- The `accent` hex value generates the CSS theme highlight dynamically.

## Deployment

Alster compiles everything into pure, static HTML/CSS/JS inside the `dist/` directory.

### GitHub Pages

Add `.github/workflows/pages.yml`:

```yaml
name: Deploy to GitHub Pages
on:
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Build static site
        run: |
          go run ./cmd/alster build
      - uses: actions/upload-pages-artifact@v3
        with:
          path: dist/
      - id: deployment
        uses: actions/deploy-pages@v4
```

### Cloudflare Pages

1. Create a Cloudflare Pages project linked to your Git repository.
2. Configure build settings:
   - **Framework preset**: None
   - **Build command**: `go run ./cmd/alster build`
   - **Build output directory**: `dist`
   - **Environment variable**: `GO_VERSION` = `1.24.0`

### Static Server / VPS

Sync the `dist/` folder to any static web server (Nginx, Caddy, Apache):

```sh
rsync -avz --delete dist/ user@server:/var/www/alster/
```

## License

MIT
