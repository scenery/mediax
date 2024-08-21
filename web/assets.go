package web

import (
	"embed"
	"io/fs"
)

//go:embed tmpl/* static/*
var webFiles embed.FS

func GetTemplateFileSystem() (fs.FS, error) {
	return fs.Sub(webFiles, "tmpl")
}

func GetStaticFileSystem() (fs.FS, error) {
	return fs.Sub(webFiles, "static")
}
