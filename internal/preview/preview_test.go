package preview

import (
	"github.com/alex/alster/internal/config"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestPreviewHTTP(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "work"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "work/index.html"), []byte("<body>Project</body>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "_theme"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "_theme/style.css"), []byte("body { color: red; }"), 0644); err != nil {
		t.Fatal(err)
	}
	c := config.Config{Output: root}
	var mu sync.RWMutex
	version := int64(123)
	h := newHandler(&c, &mu, &version)
	for _, tc := range []struct {
		path   string
		status int
		want   string
	}{
		{"/work", http.StatusMovedPermanently, ""},
		{"/work/", http.StatusOK, "/__alster/version"},
		{"/work/index.html", http.StatusOK, "const initial=\"123\""},
		{"/_theme/style.css", http.StatusOK, "color: red"},
		{"/__alster/version", http.StatusOK, "123"},
		{"/missing", http.StatusNotFound, ""},
		{"/.alster-output", http.StatusNotFound, ""},
	} {
		t.Run(tc.path, func(t *testing.T) {
			r := httptest.NewRecorder()
			h.ServeHTTP(r, httptest.NewRequest("GET", tc.path, nil))
			if r.Code != tc.status {
				t.Fatalf("status=%d want %d", r.Code, tc.status)
			}
			if !strings.Contains(r.Body.String(), tc.want) {
				t.Fatalf("body=%s", r.Body.String())
			}
			if tc.path == "/_theme/style.css" && !strings.Contains(r.Header().Get("Content-Type"), "text/css") {
				t.Fatalf("CSS Content-Type was %q, want text/css", r.Header().Get("Content-Type"))
			}
		})
	}
}
