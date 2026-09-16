package config

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Link struct {
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
}
type Config struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	BaseURL     string `yaml:"base_url"`
	Accent      string `yaml:"accent"`
	Content     string `yaml:"content"`
	Output      string `yaml:"output"`
	Templates   string `yaml:"templates"`
	ImageWidth  int    `yaml:"image_width"`
	Social      []Link `yaml:"social"`
	Nav         []Link `yaml:"nav"`
}

func Load(filename string) (Config, error) {
	c := Config{Title: "Alster", Accent: "#315d49", Content: "content", Output: "dist", Templates: "templates", ImageWidth: 1600}
	b, err := os.ReadFile(filename)
	if err != nil {
		return c, err
	}
	d := yaml.NewDecoder(bytes.NewReader(b))
	d.KnownFields(true)
	if err := d.Decode(&c); err != nil {
		return c, fmt.Errorf("config: %w", err)
	}
	if c.Title == "" {
		return c, fmt.Errorf("config: title must not be empty")
	}
	if !regexp.MustCompile(`^#[0-9a-fA-F]{6}$`).MatchString(c.Accent) {
		return c, fmt.Errorf("config: accent must be a six-digit hex color")
	}
	if c.ImageWidth < 1 || c.ImageWidth > 4096 {
		return c, fmt.Errorf("config: image_width must be between 1 and 4096")
	}
	if err := ValidateBaseURL(c.BaseURL); err != nil {
		return c, err
	}

	for _, l := range append(append([]Link{}, c.Nav...), c.Social...) {
		if l.Label == "" || l.URL == "" {
			return c, fmt.Errorf("config: links require label and url")
		}
		if err := ValidateURL(l.URL); err != nil {
			return c, err
		}
	}
	root := filepath.Dir(filename)
	for _, p := range []*string{&c.Content, &c.Output, &c.Templates} {
		if *p == "" {
			return c, fmt.Errorf("config: directory paths must not be empty")
		}
		if !filepath.IsAbs(*p) {
			*p = filepath.Join(root, *p)
		}
		*p, err = filepath.Abs(*p)
		if err != nil {
			return c, err
		}
	}
	return c, nil
}

func ValidateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", raw, err)
	}
	if u.Scheme != "" && u.Scheme != "https" && u.Scheme != "http" && u.Scheme != "mailto" {
		return fmt.Errorf("unsupported URL scheme in %q", raw)
	}
	return nil
}

func ValidateBaseURL(raw string) error {
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return fmt.Errorf("base_url must be an absolute HTTP(S) URL without credentials, query, or fragment")
	}
	return nil
}
