package helpers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	return &Templates{template.Must(parseGlobRecurse(workingDir))}
}

func parseGlobRecurse(directory string) (*template.Template, error) {
	templates := template.New("")
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".gohtml") {
			_, err = templates.ParseFiles(path)
			if err != nil {
				log.Println(err)
			}
		}
		return err
	})

	if err != nil {
		return nil, err
	}
	return templates, nil
}
