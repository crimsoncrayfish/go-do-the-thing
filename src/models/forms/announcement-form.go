package form_models

import (
	"time"
)

type AnnouncementForm struct {
	Title       string     `json:"title"`
	ExpiresDate *time.Time `json:"expires_date"`
	errors      map[string]string
}

func NewAnnouncementForm() *AnnouncementForm {
	// Set default expiry to 1 month from today
	defaultExpiry := time.Now().AddDate(0, 1, 0)

	return &AnnouncementForm{
		ExpiresDate: &defaultExpiry,
		errors:      make(map[string]string),
	}
}

func (f *AnnouncementForm) GetErrors() map[string]string {
	return f.errors
}

func (f *AnnouncementForm) SetError(field, message string) {
	f.errors[field] = message
}

func (f *AnnouncementForm) HasErrors() bool {
	return len(f.errors) > 0
}

func (f *AnnouncementForm) GetError(field string) string {
	return f.errors[field]
}

func (f *AnnouncementForm) ClearErrors() {
	f.errors = make(map[string]string)
}
