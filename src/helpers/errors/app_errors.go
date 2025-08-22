package errors

import (
	"errors"
	"fmt"
	"strings"
)

type AppErrorCode string

const (
	AuthFailed            AppErrorCode = "auth checks failed"
	ErrNotImplemented     AppErrorCode = "not implemented"
	ErrAccessDenied       AppErrorCode = "no access"
	ErrPermissionDenied   AppErrorCode = "no permission"
	ErrNotFound           AppErrorCode = "not found"
	ErrDBReadFailed       AppErrorCode = "db read failed"
	ErrDBInsertFailed     AppErrorCode = "db insert failed"
	ErrDBUpdateFailed     AppErrorCode = "db update failed"
	ErrDBDeleteFailed     AppErrorCode = "db delete failed"
	ErrDBGenericError     AppErrorCode = "db error"
	ErrKeysNotLoadedError AppErrorCode = "keys not loaded"
	ErrEnumMismatch       AppErrorCode = "enum value mismatch"
	ErrValidationFailed   AppErrorCode = "validation failed"
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

func NewNotImplemented(reason string, args ...any) *AppError {
	return New(ErrNotImplemented, reason, args...)
}

func NewDbReadFailed(entityType string, err error, additionalMsg string, args ...any) *AppError {
	if additionalMsg == "" {
		return New(ErrDBReadFailed, "failed to read %s: %w", entityType, err)
	}
	fullMsg := fmt.Sprintf("failed to read %s - %s: %%w", entityType, additionalMsg)
	allArgs := append(args, err)
	return New(ErrDBReadFailed, fullMsg, allArgs...)
}

func NewDbInsertFailed(entityType string, err error, additionalMsg string, args ...any) *AppError {
	if additionalMsg == "" {
		return New(ErrDBInsertFailed, "failed to create %s: %w", entityType, err)
	}
	fullMsg := fmt.Sprintf("failed to create %s - %s: %%w", entityType, additionalMsg)
	allArgs := append(args, err)
	return New(ErrDBInsertFailed, fullMsg, allArgs...)
}

func NewDbUpdateFailed(entityType string, err error, additionalMsg string, args ...any) *AppError {
	if additionalMsg == "" {
		return New(ErrDBUpdateFailed, "failed to update %s: %w", entityType, err)
	}
	fullMsg := fmt.Sprintf("failed to update %s - %s: %%w", entityType, additionalMsg)
	allArgs := append(args, err)
	return New(ErrDBUpdateFailed, fullMsg, allArgs...)
}

func NewDbDeleteFailed(entityType string, err error, additionalMsg string, args ...any) *AppError {
	if additionalMsg == "" {
		return New(ErrDBDeleteFailed, "failed to delete %s: %w", entityType, err)
	}
	fullMsg := fmt.Sprintf("failed to delete %s - %s: %%w", entityType, additionalMsg)
	allArgs := append(args, err)
	return New(ErrDBDeleteFailed, fullMsg, allArgs...)
}

func NewDbGenericError(entityType string, err error, additionalMsg string, args ...any) *AppError {
	if additionalMsg == "" {
		return New(ErrDBGenericError, "database error for %s: %w", entityType, err)
	}
	fullMsg := fmt.Sprintf("database error for %s - %s: %%w", entityType, additionalMsg)
	allArgs := append(args, err)
	return New(ErrDBGenericError, fullMsg, allArgs...)
}

func NewAccessDenied(resourceType string, reason string, args ...any) *AppError {
	if reason == "" {
		return New(ErrAccessDenied, "access denied to %s", resourceType)
	}
	fullMsg := fmt.Sprintf("access denied to %s - %s", resourceType, reason)
	return New(ErrAccessDenied, fullMsg, args...)
}

func NewAuthFailed(operation string, reason string, args ...any) *AppError {
	if reason == "" {
		return New(AuthFailed, "%s failed", operation)
	}
	fullMsg := fmt.Sprintf("%s failed - %s", operation, reason)
	return New(AuthFailed, fullMsg, args...)
}

func NewValidationFailed(field string, reason string, args ...any) *AppError {
	if reason == "" {
		return New(ErrValidationFailed, "validation failed for %s", field)
	}
	fullMsg := fmt.Sprintf("validation failed for %s - %s", field, reason)
	return New(ErrValidationFailed, fullMsg, args...)
}
