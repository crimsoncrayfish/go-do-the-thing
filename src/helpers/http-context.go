package helpers

import (
	"context"
	"errors"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/constants"
	"net/http"
	"strconv"
)

type HttpContext struct {
	Values map[constants.ContextKey]string
}

func (u HttpContext) Get(key constants.ContextKey) string {
	return u.Values[key]
}

func GetUserFromContext(r *http.Request) (id int64, email string, name string, is_admin bool, err error) {
	context, ok := r.Context().Value(constants.AuthContext).(HttpContext)
	if !ok {
		return 0, "", "", false, errors.New("could not read http context")
	}
	idString := context.Get(constants.AuthUserId)
	email = context.Get(constants.AuthUserEmail)
	name = context.Get(constants.AuthUserName)
	is_admin = context.Get(constants.AuthIsAdmin) == "true"
	if email == "" || name == "" || idString == "" {
		return 0, "", "", false, errors.New("auth context values not set")
	}
	id, err = strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return 0, "", "", false, err
	}
	return id, email, name, is_admin, err
}

const source = "HttpContext"

func GetNameFromContext(ctx context.Context) string {
	context, ok := ctx.Value(constants.AuthContext).(HttpContext)
	assert.IsTrue(ok, source, "Failed to get the user name from the context. Context not set")
	return context.Get(constants.AuthUserName)
}

func GetEmailFromContext(ctx context.Context) string {
	context, ok := ctx.Value(constants.AuthContext).(HttpContext)
	assert.IsTrue(ok, source, "Failed to get the user name from the context. Context not set")
	return context.Get(constants.AuthUserEmail)
}

func GetIsAdminFromContext(ctx context.Context) bool {
	context, ok := ctx.Value(constants.AuthContext).(HttpContext)
	if !ok {
		return false
	}
	return context.Get(constants.AuthIsAdmin) == "true"
}
