package helpers

import (
	"html/template"
	"io"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewRenderer(workingDir string) *Templates {
	return &Templates{template.Must(template.ParseGlob(workingDir + "/tmpl/*.gohtml"))}
}
