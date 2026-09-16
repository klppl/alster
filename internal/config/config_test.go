package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		bad        bool
	}{
		{"defaults", "title: Test", false}, {"typo", "titel: Test", true}, {"color", "accent: 'red;display:none'", true}, {"width", "image_width: 0", true}, {"base", "base_url: /repo", true}, {"link", "nav: [{label: X, url: 'javascript:alert(1)'}]", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			file := filepath.Join(root, "alster.yaml")
			if err := os.WriteFile(file, []byte(tc.body), 0644); err != nil {
				t.Fatal(err)
			}
			c, err := Load(file)
			if (err != nil) != tc.bad {
				t.Fatalf("error=%v", err)
			}
			if !tc.bad && c.Content != filepath.Join(root, "content") {
				t.Fatal("paths should be config-relative")
			}
		})
	}
}
