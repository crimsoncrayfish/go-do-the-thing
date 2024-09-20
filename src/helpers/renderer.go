package helpers

import (
	"go-do-the-thing/src/helpers/slog"
	"html/template"
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
	logger := slog.NewLogger("Renderer")
	return &Templates{template.Must(parseGlobRecurse(workingDir, logger))}
}

func parseGlobRecurse(directory string, logger slog.Logger) (*template.Template, error) {
	templates := template.New("")
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".gohtml") {
			_, err = templates.ParseFiles(path)
			if err != nil {
				logger.Error(err, "failed to collect template files")
			}
		}
		return err
	})

	if err != nil {
		return nil, err
	}
	return templates, nil
}
