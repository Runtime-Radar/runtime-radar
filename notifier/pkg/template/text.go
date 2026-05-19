package template

import (
	"path/filepath"
	"text/template"

	history "github.com/runtime-radar/runtime-radar/history-api/pkg/model"
)

var (
	TextFilePaths map[string]string
	DefaultTexts  map[string]*template.Template
)

func initTexts(templatesFolder string) {
	templatesTextFolder := filepath.Join(templatesFolder, "text")
	TextFilePaths = map[string]string{
		history.EventTypeRuntimeEvent: filepath.Join(templatesTextFolder, "runtime_event.tmpl"),
	}

	DefaultTexts = map[string]*template.Template{
		history.EventTypeRuntimeEvent: mustParseText(TextFilePaths[history.EventTypeRuntimeEvent]),
	}
}

func NewText(name, tpl string) (*template.Template, error) {
	return template.New(name).Funcs(funcs).Parse(tpl)
}

func mustParseText(filePath string) *template.Template {
	base := filepath.Base(filePath)
	return template.Must(template.New(base).Funcs(funcs).ParseFiles(filePath))
}
