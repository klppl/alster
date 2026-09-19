package build

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/HugoSmits86/nativewebp"
	"github.com/alex/alster/internal/config"
)

func fixture(t *testing.T) config.Config {
	t.Helper()
	root := t.TempDir()
	content := filepath.Join(root, "content")
	if err := os.MkdirAll(filepath.Join(content, "work"), 0755); err != nil {
		t.Fatal(err)
	}
	c := config.Config{Title: "Test", Accent: "#315d49", Content: content, Output: filepath.Join(root, "dist"), Templates: filepath.Join(root, "templates"), ImageWidth: 40, BaseURL: "https://example.com/repo"}
	put(t, filepath.Join(content, "work/index.md"), "---\ntitle: Work\ncover: cover.png\ntags: [Go]\n---\n[Notes](notes.md)\n\n![Cover](cover.png)")
	put(t, filepath.Join(content, "work/notes.md"), "---\ntitle: Notes\n---\n[Home](index.md)")
	im := image.NewNRGBA(image.Rect(0, 0, 80, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			im.Set(x, y, color.NRGBA{uint8(x * 3), 100, 150, 255})
		}
	}
	f, err := os.Create(filepath.Join(content, "work/cover.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	return c
}
func put(t *testing.T, name, body string) {
	t.Helper()
	if err := os.WriteFile(name, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}
func read(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
func TestBuildEndToEnd(t *testing.T) {
	c := fixture(t)
	if err := Run(c); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"index.html", "work/index.html", "work/notes.html", "work/cover.png", "_images/work/cover.png", "_images/work/cover.png.webp", "_theme/style.css", ".nojekyll"} {
		if _, err := os.Stat(filepath.Join(c.Output, file)); err != nil {
			t.Error(err)
		}
	}
	detail := read(t, filepath.Join(c.Output, "work/index.html"))
	for _, want := range []string{`href="../index.html"`, `href="notes.html"`, `href="../_theme/style.css"`, `https://example.com/repo/work/index.html`, `srcset="../_images/work/cover.png.webp"`} {
		if !strings.Contains(detail, want) {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Contains(detail, "ZgotmplZ") {
		t.Fatal("template rejected URL")
	}
	f, err := os.Open(filepath.Join(c.Output, "_images/work/cover.png.webp"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	im, err := nativewebp.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if im.Bounds().Dx() != 40 || im.Bounds().Dy() != 20 {
		t.Fatalf("bounds: %v", im.Bounds())
	}
	// A renamed source removes obsolete pages on the next successful build.
	if err := os.Rename(filepath.Join(c.Content, "work/notes.md"), filepath.Join(c.Content, "work/renamed.md")); err != nil {
		t.Fatal(err)
	}
	if err := Run(c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(c.Output, "work/notes.html")); !os.IsNotExist(err) {
		t.Fatal("stale page retained")
	}
}
func TestFailedBuildPreservesOutput(t *testing.T) {
	c := fixture(t)
	if err := Run(c); err != nil {
		t.Fatal(err)
	}
	before := read(t, filepath.Join(c.Output, "index.html"))
	put(t, filepath.Join(c.Content, "work/notes.md"), "---\nstatus: invalid\n---\n")
	if err := Run(c); err == nil {
		t.Fatal("expected error")
	}
	if read(t, filepath.Join(c.Output, "index.html")) != before {
		t.Fatal("previous site changed")
	}
}
func TestRejectUnsafeOutput(t *testing.T) {
	t.Run("unowned", func(t *testing.T) {
		c := fixture(t)
		if err := os.MkdirAll(c.Output, 0755); err != nil {
			t.Fatal(err)
		}
		put(t, filepath.Join(c.Output, "keep.txt"), "keep")
		if err := Run(c); err == nil {
			t.Fatal("expected error")
		}
		if read(t, filepath.Join(c.Output, "keep.txt")) != "keep" {
			t.Fatal("file changed")
		}
	})
	t.Run("overlap", func(t *testing.T) {
		c := fixture(t)
		c.Output = c.Content
		if err := Run(c); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("symlink", func(t *testing.T) {
		c := fixture(t)
		if err := os.Symlink(filepath.Join(c.Content, "work/index.md"), filepath.Join(c.Content, "work/link.md")); err != nil {
			t.Fatal(err)
		}
		if err := Run(c); err == nil {
			t.Fatal("expected error")
		}
	})
}
func TestOverridesAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name, template string
		bad            bool
	}{
		{"override", `{{define "detail.html"}}Custom {{.Page.Title}}{{end}}`, false},
		{"syntax", `{{define`, true},
		{"execution", `{{define "detail.html"}}{{.Missing}}{{end}}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := fixture(t)
			if err := os.MkdirAll(c.Templates, 0755); err != nil {
				t.Fatal(err)
			}
			put(t, filepath.Join(c.Templates, "detail.html"), tc.template)
			err := Run(c)
			if (err != nil) != tc.bad {
				t.Fatalf("error=%v", err)
			}
			if !tc.bad && read(t, filepath.Join(c.Output, "work/index.html")) != "Custom Work" {
				t.Fatal("override not used")
			}
		})
	}
}
func TestBrokenImageAndMissingCover(t *testing.T) {
	t.Run("corrupt", func(t *testing.T) {
		c := fixture(t)
		put(t, filepath.Join(c.Content, "work/cover.png"), "not an image")
		if err := Run(c); err == nil {
			t.Fatal("expected image error")
		}
	})
	t.Run("missing", func(t *testing.T) {
		c := fixture(t)
		put(t, filepath.Join(c.Content, "work/index.md"), "---\ncover: absent.png\n---\n")
		if err := Run(c); err == nil {
			t.Fatal("expected cover error")
		}
	})
}

func TestInvalidBaseOverride(t *testing.T) {
	c := fixture(t)
	c.BaseURL = "javascript:alert(1)"
	if err := Run(c); err == nil {
		t.Fatal("expected base URL error")
	}
}

func TestSpecialCharactersInPaths(t *testing.T) {
	c := fixture(t)
	if err := os.Rename(filepath.Join(c.Content, "work"), filepath.Join(c.Content, "work #1")); err != nil {
		t.Fatal(err)
	}
	if err := Run(c); err != nil {
		t.Fatal(err)
	}
	projects := read(t, filepath.Join(c.Output, "projects/index.html"))
	if !strings.Contains(projects, `href="../work%20%231/index.html"`) {
		t.Fatal("gallery URL must escape literal spaces and fragment characters")
	}
	if !strings.Contains(projects, `srcset="../_images/work%20%231/cover.png.webp"`) {
		t.Fatal("image URL must escape literal spaces and fragment characters")
	}
}

func TestEmptyContent(t *testing.T) {
	root := t.TempDir()
	c := config.Config{Title: "Empty", Accent: "#315d49", Content: filepath.Join(root, "content"), Output: filepath.Join(root, "dist"), Templates: filepath.Join(root, "templates"), ImageWidth: 40}
	if err := os.Mkdir(c.Content, 0755); err != nil {
		t.Fatal(err)
	}
	if err := Run(c); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, filepath.Join(c.Output, "projects/index.html")), "No projects yet.") {
		t.Fatal("empty state missing")
	}
}

func TestOutputAliasCannotOverlapContent(t *testing.T) {
	c := fixture(t)
	alias := filepath.Join(filepath.Dir(c.Content), "alias")
	if err := os.Symlink(c.Content, alias); err != nil {
		t.Fatal(err)
	}
	c.Output = filepath.Join(alias, "generated")
	if err := Run(c); err == nil {
		t.Fatal("expected overlapping output error through symlinked parent")
	}
}

func TestStaticSubpathLinks(t *testing.T) {
	c := fixture(t)
	if err := Run(c); err != nil {
		t.Fatal(err)
	}
	handler := http.StripPrefix("/portfolio/", http.FileServer(http.Dir(c.Output)))
	refs := regexp.MustCompile(`(?:href|src|srcset)="([^"]+)"`)
	for _, file := range []string{"index.html", "work/index.html", "work/notes.html"} {
		base, _ := url.Parse("https://static.test/portfolio/" + file)
		html := read(t, filepath.Join(c.Output, file))
		for _, match := range refs.FindAllStringSubmatch(html, -1) {
			ref, err := url.Parse(match[1])
			if err != nil {
				t.Fatal(err)
			}
			if ref.IsAbs() || ref.Host != "" || strings.HasPrefix(match[1], "#") {
				continue
			}
			target := base.ResolveReference(ref)
			if !strings.HasPrefix(target.Path, "/portfolio/") {
				t.Fatalf("escaped subpath: %s", target)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", target.String(), nil))
			if response.Code != http.StatusOK && response.Code != http.StatusMovedPermanently {
				t.Errorf("%s: %s returned %d", file, target, response.Code)
			}
		}
	}
}

func TestMultiSectionSite(t *testing.T) {
	c := fixture(t)
	blogPostDir := filepath.Join(c.Content, "blog/first-post")
	if err := os.MkdirAll(blogPostDir, 0755); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(blogPostDir, "index.md"), "---\ntitle: First Post\ndate: 2026-09-10\ntags: [Design]\n---\nHello world from post.")

	aboutDir := filepath.Join(c.Content, "about")
	if err := os.MkdirAll(aboutDir, 0755); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(aboutDir, "index.md"), "---\ntitle: About Me\n---\nPersonal introduction.")

	put(t, filepath.Join(c.Content, "frontpage-about.md"), "---\ntitle: Custom Frontpage Title\n---\nCustom frontpage bio text.")

	if err := Run(c); err != nil {
		t.Fatal(err)
	}

	for _, wantFile := range []string{
		"index.html",
		"projects/index.html",
		"blog/index.html",
		"blog/first-post/index.html",
		"about/index.html",
	} {
		if _, err := os.Stat(filepath.Join(c.Output, wantFile)); err != nil {
			t.Errorf("missing expected output file %s: %v", wantFile, err)
		}
	}

	if _, err := os.Stat(filepath.Join(c.Output, "frontpage-about.html")); err == nil {
		t.Error("frontpage-about.md must not generate a standalone HTML page")
	}

	indexHTML := read(t, filepath.Join(c.Output, "index.html"))
	if !strings.Contains(indexHTML, "Custom Frontpage Title") {
		t.Error("index.html missing custom frontpage title from frontpage-about.md")
	}
	if !strings.Contains(indexHTML, "Custom frontpage bio text.") {
		t.Error("index.html missing custom bio text from frontpage-about.md")
	}
	if !strings.Contains(indexHTML, "nav-index") {
		t.Error("index.html missing nav-index section")
	}
	if !strings.Contains(indexHTML, "Projects") || !strings.Contains(indexHTML, "Writing") {
		t.Error("index.html missing navigation sections")
	}

	projectsHTML := read(t, filepath.Join(c.Output, "projects/index.html"))
	if !strings.Contains(projectsHTML, "filter-btn") {
		t.Error("projects/index.html missing filter buttons")
	}

	blogHTML := read(t, filepath.Join(c.Output, "blog/index.html"))
	if !strings.Contains(blogHTML, "First Post") {
		t.Error("blog/index.html missing post title")
	}
	if !strings.Contains(blogHTML, "filter-btn") {
		t.Error("blog/index.html missing filter buttons")
	}
	if !strings.Contains(blogHTML, `class="writing-item" data-tags="Design"`) {
		t.Errorf("blog/index.html missing writing-item with data-tags: %s", blogHTML)
	}

	postHTML := read(t, filepath.Join(c.Output, "blog/first-post/index.html"))
	if !strings.Contains(postHTML, "Hello world from post.") {
		t.Error("blog post HTML missing content")
	}
	if !strings.Contains(postHTML, `href="../../blog/index.html"`) {
		t.Errorf("blog post back link must point to ../../blog/index.html: %s", postHTML)
	}

	aboutHTML := read(t, filepath.Join(c.Output, "about/index.html"))
	if !strings.Contains(aboutHTML, "Personal introduction.") {
		t.Error("about page HTML missing content")
	}
}

func TestSwedishSite(t *testing.T) {
	c := fixture(t)
	c.Lang = "sv"
	c.Nav = []config.Link{
		{Label: "Projekt", URL: "/projects/index.html"},
		{Label: "Skrivande", URL: "/blog/index.html"},
		{Label: "Om", URL: "/about/index.html"},
	}
	if err := Run(c); err != nil {
		t.Fatal(err)
	}

	indexHTML := read(t, filepath.Join(c.Output, "index.html"))
	if !strings.Contains(indexHTML, `<html lang="sv">`) {
		t.Errorf("index.html missing lang=sv: %s", indexHTML[:200])
	}
	if !strings.Contains(indexHTML, "Projekt") || !strings.Contains(indexHTML, "Skrivande") {
		t.Error("index.html missing Swedish navigation items")
	}
	if !strings.Contains(indexHTML, "Hoppa till innehåll") {
		t.Error("index.html missing Swedish skip link")
	}
	if !strings.Contains(indexHTML, "https://github.com/klppl") || !strings.Contains(indexHTML, "klppl") {
		t.Error("index.html missing footer klppl link")
	}

	projectsHTML := read(t, filepath.Join(c.Output, "projects/index.html"))
	if !strings.Contains(projectsHTML, ">Alla</button>") {
		t.Error("projects/index.html missing Swedish filter button 'Alla'")
	}
	if !strings.Contains(projectsHTML, `aria-label="Filtrera projekt efter tagg"`) {
		t.Error("projects/index.html missing Swedish filter aria label")
	}

	workHTML := read(t, filepath.Join(c.Output, "work/index.html"))
	if !strings.Contains(workHTML, "← Projekt") {
		t.Error("work detail page missing Swedish back link '← Projekt'")
	}
}

func TestAboutProfilePhoto(t *testing.T) {
	c := fixture(t)
	if err := os.MkdirAll(filepath.Join(c.Content, "about"), 0755); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(c.Content, "about/index.md"), "---\ntitle: Om\nphoto: profile.png\n---\nBio text.")

	im := image.NewNRGBA(image.Rect(0, 0, 80, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 80; x++ {
			im.Set(x, y, color.NRGBA{100, 100, 100, 255})
		}
	}
	f, err := os.Create(filepath.Join(c.Content, "about/profile.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := Run(c); err != nil {
		t.Fatal(err)
	}

	aboutHTML := read(t, filepath.Join(c.Output, "about/index.html"))
	if !strings.Contains(aboutHTML, "page-profile-photo") {
		t.Errorf("about page missing page-profile-photo: %s", aboutHTML)
	}
	if !strings.Contains(aboutHTML, `class="page-layout has-photo"`) {
		t.Errorf("about page missing has-photo class on page-layout: %s", aboutHTML)
	}
	if !strings.Contains(aboutHTML, "_images/about/profile.png.webp") {
		t.Errorf("about page missing webp variant: %s", aboutHTML)
	}
}

func TestAboutSidebar(t *testing.T) {
	c := fixture(t)
	c.Social = []config.Link{
		{Label: "GitHub", URL: "https://github.com/alex"},
		{Label: "Bluesky", URL: "https://bsky.app/profile/alex.bsky.social"},
		{Label: "E-post", URL: "mailto:alex@example.com"},
	}
	if err := os.MkdirAll(filepath.Join(c.Content, "about"), 0755); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(c.Content, "about/index.md"), "---\ntitle: Om\n---\nBio text.")

	if err := Run(c); err != nil {
		t.Fatal(err)
	}

	aboutHTML := read(t, filepath.Join(c.Output, "about/index.html"))
	if !strings.Contains(aboutHTML, "page-sidebar") {
		t.Errorf("about page missing page-sidebar: %s", aboutHTML)
	}
	if !strings.Contains(aboutHTML, "with-sidebar") {
		t.Errorf("about page missing with-sidebar class: %s", aboutHTML)
	}
	if !strings.Contains(aboutHTML, "https://github.com/alex") {
		t.Errorf("about page missing GitHub link: %s", aboutHTML)
	}
	if !strings.Contains(aboutHTML, "https://bsky.app/profile/alex.bsky.social") {
		t.Errorf("about page missing Bluesky link: %s", aboutHTML)
	}
	if !strings.Contains(aboutHTML, `rel="noopener me"`) {
		t.Errorf("about page external links missing rel=noopener me: %s", aboutHTML)
	}
}

func TestFrontpageMarkdownIndex(t *testing.T) {
	c := fixture(t)
	put(t, filepath.Join(c.Content, "index.md"), "---\ntitle: Welcome to My Studio\n---\nHere is my custom bio and introduction.")

	if err := Run(c); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	indexHTML := read(t, filepath.Join(c.Output, "index.html"))
	if !strings.Contains(indexHTML, "Welcome to My Studio") {
		t.Error("index.html missing custom frontpage title from index.md")
	}
	if !strings.Contains(indexHTML, "Here is my custom bio and introduction.") {
		t.Error("index.html missing custom bio text from index.md")
	}
}

func TestFrontpageMarkdownHeadingWithoutFrontmatter(t *testing.T) {
	c := fixture(t)
	put(t, filepath.Join(c.Content, "index.md"), "# Heading From Markdown\n\nBio text without frontmatter.")

	if err := Run(c); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	indexHTML := read(t, filepath.Join(c.Output, "index.html"))
	if !strings.Contains(indexHTML, "Heading From Markdown") {
		t.Error("index.html missing custom frontpage title from markdown H1")
	}
	if !strings.Contains(indexHTML, "Bio text without frontmatter.") {
		t.Error("index.html missing custom bio text from markdown")
	}
}

