package markdown

import (
	"github.com/alex/alster/internal/images"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		bad         bool
	}{
		{"plain", "# A project", false},
		{"yaml", "---\r\ntitle: Example\r\ndate: 2026-09-01\r\n---\r\nBody", false},
		{"unclosed", "---\ntitle: Missing", true},
		{"invalid date", "---\ndate: yesterday\n---\n", true},
		{"unknown field", "---\nstatuz: active\n---\n", true},
		{"invalid status", "---\nstatus: finished\n---\n", true},
		{"unsafe url", "---\nlive_url: 'javascript:alert(1)'\n---\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := Parse([]byte(tc.input))
			if (err != nil) != tc.bad {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
func TestRender(t *testing.T) {
	body := []byte("[More](notes.md#details) [Home](/index.html)\n\n![A chart](chart.png)\n\n<script>alert(1)</script>\n\n[Bad](javascript:alert(1))")
	out, err := Render(body, "work/index.md", "../", map[string]images.Variant{"work/chart.png": {WebP: "_images/work/chart.png.webp", Fallback: "_images/work/chart.png", Width: 100, Height: 50}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`href="notes.html#details"`, `href="../index.html"`, `<picture>`, `srcset="../_images/work/chart.png.webp"`, `alt="A chart"`, `width="100"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(string(out), "<script>") || strings.Contains(string(out), `href="javascript:`) {
		t.Fatalf("unsafe HTML: %s", out)
	}
}

func TestEmptyFrontmatter(t *testing.T) {
	_, body, err := Parse([]byte("---\n---\nBody"))
	if err != nil || string(body) != "Body" {
		t.Fatalf("body=%q error=%v", body, err)
	}
}

func TestHeadingFragments(t *testing.T) {
	out, err := Render([]byte("## Details\n\n[Details](#details)"), "work/index.md", "../", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `id="details"`) {
		t.Fatal("heading fragment target missing")
	}
}
