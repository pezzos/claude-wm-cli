package config

import (
	"embed"
	"io"
)

//go:embed system/commands system/settings.json.template system/README.md
var EmbeddedFS embed.FS

func Open(path string) (io.ReadCloser, error) {
	return EmbeddedFS.Open(path)
}