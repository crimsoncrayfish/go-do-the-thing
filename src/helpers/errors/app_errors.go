package errors

import (
	"errors"
	"fmt"
	"strings"
)

type AppErrorCode string

const (
	ErrAccessDenied       AppErrorCode = "no access"
	ErrPermissionDenied   AppErrorCode = "no permission"
	ErrNotFound           AppErrorCode = "not found"
	ErrDBReadFailed       AppErrorCode = "db read failed"
	ErrDBInsertFailed     AppErrorCode = "db insert failed"
	ErrDBUpdateFailed     AppErrorCode = "db update failed"
	ErrDBDeleteFailed     AppErrorCode = "db delete failed"
	ErrDBGenericError     AppErrorCode = "db error"
	ErrKeysNotLoadedError AppErrorCode = "keys not loaded"
)

type AppError struct {
	code    AppErrorCode
	message string
	error   error
}

func (e *AppError) Error() string {
	if e.error != nil {
		return fmt.Sprintf("%s: %s\n %v", e.code, e.message, e.error)
	}
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

func (e *AppError) Code() AppErrorCode {
	return e.code
}

func New(code AppErrorCode, msg_fmt string, args ...any) *AppError {
	err := fmt.Errorf(msg_fmt, args...)
	unwrapped := errors.Unwrap(err)

	if index := strings.Index(msg_fmt, "%w"); index >= 0 {
		msg_fmt = msg_fmt[:index] + "%v" + msg_fmt[index+2:]
	}
	message := fmt.Sprintf(msg_fmt, args...)

	return &AppError{
		code:    code,
		message: message,
		error:   unwrapped,
	}
}
