package images

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/draw"
)

// Variant paths are relative to the output root. Originals remain downloadable.
type Variant struct {
	WebP, Fallback string
	Width, Height  int
}

func Process(src, output, rel string, maxWidth int) (Variant, error) {
	var v Variant
	ext := strings.ToLower(filepath.Ext(src))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return v, nil
	}
	f, err := os.Open(src)
	if err != nil {
		return v, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return v, fmt.Errorf("decode %s: %w", src, err)
	}
	const maxPixels = 40_000_000
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > maxPixels {
		return v, fmt.Errorf("image %s exceeds 40 megapixels", src)
	}
	if _, err = f.Seek(0, 0); err != nil {
		return v, err
	}
	im, _, err := image.Decode(f)
	if err != nil {
		return v, err
	}
	w, h := cfg.Width, cfg.Height
	if w > maxWidth {
		h = max(1, h*maxWidth/w)
		w = maxWidth
	}
	resized := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(resized, resized.Bounds(), im, im.Bounds(), draw.Over, nil)
	// Preserve the original extension in the generated name to avoid jpg/png collisions.
	v = Variant{WebP: filepath.ToSlash(filepath.Join("_images", rel+".webp")), Fallback: filepath.ToSlash(filepath.Join("_images", rel)), Width: w, Height: h}
	var fallback, webp bytes.Buffer
	if ext == ".png" {
		err = png.Encode(&fallback, resized)
	} else {
		err = jpeg.Encode(&fallback, resized, &jpeg.Options{Quality: 85})
	}
	if err != nil {
		return Variant{}, err
	}
	if err = nativewebp.Encode(&webp, resized, nil); err != nil {
		return Variant{}, err
	}
	for name, data := range map[string][]byte{v.Fallback: fallback.Bytes(), v.WebP: webp.Bytes()} {
		dest := filepath.Join(output, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return Variant{}, err
		}
		if err = os.WriteFile(dest, data, 0644); err != nil {
			return Variant{}, err
		}
	}
	return v, nil
}
