package helpers

import (
	"html/template"
	"net/http"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) RenderOk(w http.ResponseWriter, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (t *Templates) RenderWithCode(w http.ResponseWriter, code int, name string, data interface{}) error {
	w.WriteHeader(code)
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewRenderer(workingDir string) *Templates {
	return &Templates{template.Must(template.ParseGlob(workingDir + "/tmpl/*.gohtml"))}
}
