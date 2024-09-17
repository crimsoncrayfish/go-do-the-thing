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
	email := context.get(constants.AuthUserEmail)
	name := context.get(constants.AuthUserName)
	idString := context.get(constants.AuthUserId)
	var err error
	if email == "" || name == "" {
		err = errors.New("could not find user details in http context")
		return 0, "", "", err
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		err = errors.New("could not parse id from context")
		return 0, "", "", err
	}
	return id, email, name, err
}
