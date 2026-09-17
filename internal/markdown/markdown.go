package markdown

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/alex/alster/internal/config"
	"github.com/alex/alster/internal/images"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
)

type Meta struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Date        string   `yaml:"date"`
	Tags        []string `yaml:"tags"`
	LiveURL     string   `yaml:"live_url"`
	SourceURL   string   `yaml:"source_url"`
	Status      string   `yaml:"status"`
	Cover       string   `yaml:"cover"`
	Photo       string   `yaml:"photo"`
	Avatar      string   `yaml:"avatar"`
	Template    string        `yaml:"template"`
	Social      []config.Link `yaml:"social"`
}

func Parse(data []byte) (Meta, []byte, error) {
	var m Meta
	data = bytes.ReplaceAll(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}), []byte("\r\n"), []byte("\n"))
	lines := bytes.Split(data, []byte("\n"))
	body := data
	if len(lines) > 0 && string(lines[0]) == "---" {
		end := 1
		for end < len(lines) && string(lines[end]) != "---" {
			end++
		}
		if end == len(lines) {
			return m, nil, fmt.Errorf("unclosed YAML frontmatter")
		}
		d := yaml.NewDecoder(bytes.NewReader(bytes.Join(lines[1:end], []byte("\n"))))
		d.KnownFields(true)
		if err := d.Decode(&m); err != nil && err != io.EOF {
			return m, nil, err
		}
		body = bytes.Join(lines[end+1:], []byte("\n"))
	}
	if m.Date != "" {
		if _, err := time.Parse("2006-01-02", m.Date); err != nil {
			return m, nil, fmt.Errorf("date must use YYYY-MM-DD: %w", err)
		}
	}
	switch m.Status {
	case "", "active", "archived", "wip":
	default:
		return m, nil, fmt.Errorf("status must be active, archived, or wip")
	}
	for _, u := range []string{m.LiveURL, m.SourceURL, m.Cover} {
		if err := config.ValidateURL(u); err != nil {
			return m, nil, err
		}
	}
	return m, body, nil
}

// LocalURL makes site-root URLs portable, including when mounted below a repo path.
func LocalURL(raw, prefix string) string {
	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return prefix + strings.TrimPrefix(raw, "/")
	}
	return raw
}

func Render(body []byte, file, prefix string, variants map[string]images.Variant) (template.HTML, error) {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithParserOptions(parser.WithAutoHeadingID()), goldmark.WithRendererOptions(renderer.WithNodeRenderers(util.Prioritized(&imageRenderer{file: file, prefix: prefix, variants: variants}, 100))))
	doc := md.Parser().Parse(text.NewReader(body))
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if link, ok := n.(*ast.Link); ok {
			u, err := url.Parse(string(link.Destination))
			if err != nil {
				return ast.WalkStop, err
			}
			if u.Scheme == "" && u.Host == "" {
				if strings.HasSuffix(strings.ToLower(u.Path), ".md") {
					u.Path = u.Path[:len(u.Path)-3] + ".html"
					u.RawPath = ""
				}
				link.Destination = []byte(LocalURL(u.String(), prefix))
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := md.Renderer().Render(&out, body, doc); err != nil {
		return "", err
	}
	// Goldmark omits raw HTML and unsafe link schemes by default.
	return template.HTML(out.String()), nil
}

type imageRenderer struct {
	file, prefix string
	variants     map[string]images.Variant
}

func (r *imageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.render)
}
func (r *imageRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	raw := string(n.Destination)
	u, err := url.Parse(raw)
	if err != nil {
		return ast.WalkSkipChildren, err
	}
	alt := html.EscapeString(string(n.Text(source)))
	if u.Scheme != "" && u.Scheme != "https" && u.Scheme != "http" {
		_, err = fmt.Fprintf(w, `<span>%s</span>`, alt)
		return ast.WalkSkipChildren, err
	}
	key := path.Clean(path.Join(path.Dir(r.file), u.Path))
	if strings.HasPrefix(u.Path, "/") {
		key = strings.TrimPrefix(u.Path, "/")
	}
	v, ok := r.variants[key]
	if ok && u.Scheme == "" && u.Host == "" {
		_, err = fmt.Fprintf(w, `<picture><source type="image/webp" srcset="%s"><img src="%s" alt="%s" width="%d" height="%d" loading="lazy" decoding="async"></picture>`, html.EscapeString(PathURL(r.prefix+v.WebP)), html.EscapeString(PathURL(r.prefix+v.Fallback)), alt, v.Width, v.Height)
	} else {
		_, err = fmt.Fprintf(w, `<img src="%s" alt="%s" loading="lazy" decoding="async">`, html.EscapeString(LocalURL(raw, r.prefix)), alt)
	}
	return ast.WalkSkipChildren, err
}

// PathURL encodes filesystem names without treating # or ? as URL syntax.
func PathURL(raw string) string { return (&url.URL{Path: raw}).String() }
