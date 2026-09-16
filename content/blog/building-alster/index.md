---
title: "Building Alster: A Pragmatic Static Site Generator"
description: Designing a lightweight portfolio and blog generator using Go standard library, semantic HTML, and zero frontend frameworks.
date: 2026-09-02
tags: [Go, Web, Tools]
---

Static website generators frequently begin as simple scripts and gradually metastasize into complex framework wrappers requiring node runtimes, hydration hydration layers, and multi-gigabyte package trees.

Alster was built on an alternate thesis: an entire portfolio, blog, and showcase site can be compiled from Markdown into portable, self-hosted static HTML in milliseconds using standard Go libraries.

### Core Architectural Decisions

1. **Embedded Assets:** Templates and stylesheets are embedded directly into the Go binary using `embed.FS`. The generator runs anywhere as a single portable binary.
2. **Local Portability:** All output paths, stylesheets, and image variants use relative linking. Sites render identically on local file previewers, GitHub Pages subpaths, or custom root domains.
3. **Automated Image Optimization:** Cover images and illustrations are resized and encoded to modern WebP alongside standard fallbacks during compilation.

### The Standard Library is Enough

When you pair `net/http`, `html/template`, and Goldmark with intentional CSS, you eliminate hundreds of dependencies while delivering faster page loads and simpler operational maintenance.
