package helpers

import (
	"errors"
	"net/http"
)

type HttpContext struct {
	Values map[string]string
}

func (u HttpContext) get(key string) string {
	return u.Values[key]
}

func GetUserFromContext(r *http.Request) (string, string, error) {
	context, ok := r.Context().Value(AuthContext).(HttpContext)
	if !ok {
		return "", "", errors.New("could not read http context")
	}
	email := context.get(AuthUserId)
	name := context.get(AuthUserName)
	var err error
	if email == "" || name == "" {
		err = errors.New("could not find user details in http context")
	}
	return email, name, err
}
