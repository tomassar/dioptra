package ui

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var embeddedFS embed.FS

func FS() (fs.FS, error) {
	return fs.Sub(embeddedFS, "dist")
}
