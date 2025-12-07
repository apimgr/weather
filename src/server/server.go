package server

import (
	"embed"
	"html/template"
	"io/fs"
)

// Embed all templates and static files into the binary
// Following TEMPLATE.md specification lines 802-816

//go:embed templates/**/*.tmpl
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// GetTemplatesFS returns the embedded templates filesystem
func GetTemplatesFS() embed.FS {
	return templatesFS
}

// GetStaticFS returns the embedded static files filesystem
func GetStaticFS() embed.FS {
	return staticFS
}

// GetStaticSubFS returns the static files as a sub-filesystem for http.FileServer
func GetStaticSubFS() (fs.FS, error) {
	return fs.Sub(staticFS, "static")
}

// LoadTemplates loads all templates from the embedded filesystem
func LoadTemplates() (*template.Template, error) {
	return template.ParseFS(templatesFS, "templates/**/*.tmpl")
}
