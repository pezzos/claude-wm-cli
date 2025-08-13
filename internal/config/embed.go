package config

import (
	"embed"
	"io"
)

//go:embed system
var EmbeddedFS embed.FS

func Open(path string) (io.ReadCloser, error) {
	return EmbeddedFS.Open(path)
}