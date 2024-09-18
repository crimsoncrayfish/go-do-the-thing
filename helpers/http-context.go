package helpers

import (
	"errors"
	"go-do-the-thing/helpers/constants"
	"net/http"
	"strconv"
)

type HttpContext struct {
	Values map[string]string
}

func (u HttpContext) get(key string) string {
	return u.Values[key]
}

func GetUserFromContext(r *http.Request) (int64, string, string, error) {
	context, ok := r.Context().Value(constants.AuthContext).(HttpContext)
	if !ok {
		return 0, "", "", errors.New("could not read http context")
	}
	idString := context.get(constants.AuthUserId)
	email := context.get(constants.AuthUserEmail)
	name := context.get(constants.AuthUserName)
	if email == "" || name == "" || idString == "" {
		return 0, "", "", errors.New("context values not set")
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return 0, "", "", err
	}
	return id, email, name, err
}
