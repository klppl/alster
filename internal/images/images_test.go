package images

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestFormatsAndSizing(t *testing.T) {
	for _, ext := range []string{".png", ".jpg"} {
		for _, width := range []int{10, 80} {
			t.Run(ext+strconv.Itoa(width), func(t *testing.T) {
				root := t.TempDir()
				file := filepath.Join(root, "input"+ext)
				f, err := os.Create(file)
				if err != nil {
					t.Fatal(err)
				}
				im := image.NewNRGBA(image.Rect(0, 0, width, width/2))
				im.Set(0, 0, color.NRGBA{R: 100, A: 120})
				if ext == ".png" {
					err = png.Encode(f, im)
				} else {
					err = jpeg.Encode(f, im, nil)
				}
				if err != nil {
					t.Fatal(err)
				}
				if err = f.Close(); err != nil {
					t.Fatal(err)
				}
				v, err := Process(file, filepath.Join(root, "out"), "project/input"+ext, 40)
				if err != nil {
					t.Fatal(err)
				}
				if v.Width != min(width, 40) || v.Height != min(width, 40)/2 {
					t.Fatalf("bad dimensions: %+v", v)
				}
				f, err = os.Open(filepath.Join(root, "out", v.Fallback))
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				fallback, _, err := image.Decode(f)
				if err != nil {
					t.Fatal(err)
				}
				if fallback.Bounds().Dx() != v.Width {
					t.Fatal("fallback dimensions differ")
				}
			})
		}
	}
}
func TestUnsupportedFormat(t *testing.T) {
	v, err := Process("diagram.svg", t.TempDir(), "diagram.svg", 100)
	if err != nil || v.WebP != "" {
		t.Fatalf("unexpected result: %+v %v", v, err)
	}
}
