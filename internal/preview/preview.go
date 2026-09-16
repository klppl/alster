package preview

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alex/alster/internal/build"
	"github.com/alex/alster/internal/config"
)

// Run polls input files. Failed builds leave the last successful preview intact.
func Run(ctx context.Context, filename, addr string) error {
	c, err := config.Load(filename)
	if err != nil {
		return err
	}
	if err = build.Run(c); err != nil {
		return err
	}
	var mu sync.RWMutex
	version := time.Now().UnixNano()
	fingerprint, err := digest(filename, c)
	if err != nil {
		return err
	}
	mux := newHandler(&c, &mu, &version)
	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				if err := server.Shutdown(shutdownCtx); err != nil {
					log.Printf("shutdown: %v", err)
				}
				cancel()
				return
			case <-ticker.C:
				next, e := digest(filename, c)
				if e != nil {
					log.Printf("watch: %v", e)
					continue
				}
				if next == fingerprint {
					continue
				}
				fingerprint = next
				fresh, e := config.Load(filename)
				if e != nil {
					log.Printf("rebuild: %v", e)
					continue
				}
				mu.Lock()
				e = build.Run(fresh)
				if e == nil {
					c = fresh
					version = time.Now().UnixNano()
				}
				mu.Unlock()
				if e != nil {
					log.Printf("rebuild: %v", e)
				} else {
					log.Print("rebuilt")
				}
			}
		}
	}()
	log.Printf("Preview: http://%s", addr)
	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func digest(filename string, c config.Config) (string, error) {
	h := sha256.New()
	for _, root := range []string{filename, c.Content, c.Templates, "frontpage-about.md"} {
		err := filepath.WalkDir(root, func(file string, d fs.DirEntry, err error) error {
			if os.IsNotExist(err) {
				fmt.Fprint(h, file, "missing")
				return nil
			}
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink: %s", file)
			}
			b, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			fmt.Fprint(h, file, "\x00")
			h.Write(b)
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func newHandler(c *config.Config, mu *sync.RWMutex, version *int64) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /__alster/version", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprint(w, *version)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()
		w.Header().Set("Cache-Control", "no-store")
		name := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))
		if name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			http.NotFound(w, r)
			return
		}
		for _, part := range strings.Split(filepath.ToSlash(name), "/") {
			if strings.HasPrefix(part, ".") && part != "." {
				http.NotFound(w, r)
				return
			}
		}
		file := filepath.Join(c.Output, name)
		info, err := os.Stat(file)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if info.IsDir() {
			if !strings.HasSuffix(r.URL.Path, "/") {
				target := *r.URL
				target.Path += "/"
				target.RawPath = ""
				http.Redirect(w, r, target.String(), http.StatusMovedPermanently)
				return
			}
			file = filepath.Join(file, "index.html")
			info, err = os.Stat(file)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		b, err := os.ReadFile(file)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if strings.EqualFold(filepath.Ext(file), ".html") {
			script := fmt.Sprintf(`<script>(()=>{const initial=%q;setInterval(async()=>{try{const r=await fetch('/__alster/version',{cache:'no-store'});if(r.ok&&(await r.text())!==initial)location.reload()}catch{}},1000)})();</script>`, fmt.Sprint(*version))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, strings.Replace(string(b), "</body>", script+"</body>", 1))
			return
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeContent(w, r, filepath.Base(file), time.Time{}, bytes.NewReader(b))
	})
	return mux
}
