package theme

import "embed"

// Files embeds the default theme so the binary works outside its source checkout.
//
//go:embed templates/*.html static/*
var Files embed.FS
