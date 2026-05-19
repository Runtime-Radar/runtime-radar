package template

import (
	"html/template"
	"path/filepath"

	history "github.com/runtime-radar/runtime-radar/history-api/pkg/model"
)

var (
	HTMLFilePaths map[string]string
	DefaultHTMLs  map[string]*template.Template
)

func initHTMLs(templatesFolder string) {
	templatesHTMLFolder := filepath.Join(templatesFolder, "html")
	HTMLFilePaths = map[string]string{
		history.EventTypeRuntimeEvent: filepath.Join(templatesHTMLFolder, "runtime_event.html"),
	}

	DefaultHTMLs = map[string]*template.Template{
		history.EventTypeRuntimeEvent: mustParseHTML(HTMLFilePaths[history.EventTypeRuntimeEvent]),
	}
}

func NewHTML(name, tpl string) (*template.Template, error) {
	return template.New(name).Funcs(funcs).Parse(tpl)
}

func mustParseHTML(filePath string) *template.Template {
	base := filepath.Base(filePath)
	return template.Must(template.New(base).Funcs(funcs).ParseFiles(filePath))
}
