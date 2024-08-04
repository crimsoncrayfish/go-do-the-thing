package home

import (
	"fmt"
	"go-do-the-thing/helpers"
	"net/http"
)

type Handler struct {
	model     HomeObject
	templates helpers.Templates
}

func New(templates helpers.Templates) *Handler {
	return &Handler{
		model: HomeObject{
			Count: 0,
			Contacts: []Contact{
				{
					"John Doe",
					"jd@gmail.com",
				},
				{
					"Clara Doe",
					"cd@gmail.com",
				},
			},
		},
		templates: templates,
	}
}

func (h *Handler) Index(w http.ResponseWriter, _ *http.Request) {
	err := h.templates.Render(w, "index", h.model)
	if err != nil {
		fmt.Println("Failed to execute tmpl for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) IncrementCount(w http.ResponseWriter, _ *http.Request) {
	h.model.Count++

	err := h.templates.Render(w, "count", Count{h.model.Count})
	if err != nil {
		fmt.Println("Failed to execute tmpl for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h *Handler) AddContact(w http.ResponseWriter, req *http.Request) {
	h.model.Contacts = append(h.model.Contacts, Contact{req.FormValue("name"), req.FormValue("email")})
	err := h.templates.Render(w, "contacts", h.model)
	if err != nil {
		fmt.Println("Failed to execute tmpl for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
