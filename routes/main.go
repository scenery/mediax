package routes

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"

	"github.com/scenery/mediax/version"
	"github.com/scenery/mediax/web"
)

var (
	baseTemplates *template.Template
	funcMap       template.FuncMap
	tmplFS        fs.FS
	staticFS      fs.FS
)

func Init() {
	funcMap = createFuncMap()
	initBaseTemplates()
	setupRoutes()
}

func createFuncMap() template.FuncMap {
	return template.FuncMap{
		"add":     func(a, b int) int { return a + b },
		"sub":     func(a, b int) int { return a - b },
		"mul":     func(a, b int) int { return a * b },
		"div":     func(a, b int) int { return a / b },
		"version": func() string { return fmt.Sprintf("%s-%s", version.Version, version.CommitSHA) },
	}
}

func initBaseTemplates() {
	var err error

	tmplFS, err = web.GetTemplateFileSystem()
	if err != nil {
		log.Fatalf("failed to create template filesystem: %v", err)
	}

	baseTemplates, err = template.New("").Funcs(funcMap).ParseFS(tmplFS,
		"baseof.html",
		"header.html",
		"footer.html",
	)
	if err != nil {
		log.Fatalf("failed to parse base templates: %v", err)
	}
}
