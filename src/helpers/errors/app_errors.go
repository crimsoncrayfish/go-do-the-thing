package errors

import (
	"errors"
	"fmt"
	"strings"
)

type AppErrorCode string

const (
	AuthFailed            AppErrorCode = "auth checks failed"
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

func NewDbReadFailed(entityType string, err error, additionalMsg string, args ...any) *AppError {
	baseMsg := fmt.Sprintf("failed to read %s", entityType)
	if additionalMsg != "" {
		fullMsg := baseMsg + " - " + additionalMsg + ": %w"
		allArgs := append(args, err)
		return New(ErrDBReadFailed, fullMsg, allArgs...)
	}
	return New(ErrDBReadFailed, baseMsg+": %w", err)
}

func NewDbInsertFailed(entityType string, err error, additionalMsg string, args ...any) *AppError {
	baseMsg := fmt.Sprintf("failed to create %s", entityType)
	if additionalMsg != "" {
		fullMsg := baseMsg + " - " + additionalMsg + ": %w"
		allArgs := append(args, err)
		return New(ErrDBInsertFailed, fullMsg, allArgs...)
	}
	return New(ErrDBInsertFailed, baseMsg+": %w", err)
}

func NewDbUpdateFailed(entityType string, err error, additionalMsg string, args ...any) *AppError {
	baseMsg := fmt.Sprintf("failed to update %s", entityType)
	if additionalMsg != "" {
		fullMsg := baseMsg + " - " + additionalMsg + ": %w"
		allArgs := append(args, err)
		return New(ErrDBUpdateFailed, fullMsg, allArgs...)
	}
	return New(ErrDBUpdateFailed, baseMsg+": %w", err)
}

func NewDbDeleteFailed(entityType string, err error, additionalMsg string, args ...any) *AppError {
	baseMsg := fmt.Sprintf("failed to delete %s", entityType)
	if additionalMsg != "" {
		fullMsg := baseMsg + " - " + additionalMsg + ": %w"
		allArgs := append(args, err)
		return New(ErrDBDeleteFailed, fullMsg, allArgs...)
	}
	return New(ErrDBDeleteFailed, baseMsg+": %w", err)
}

func NewDbGenericError(entityType string, err error, additionalMsg string, args ...any) *AppError {
	baseMsg := fmt.Sprintf("database error for %s", entityType)
	if additionalMsg != "" {
		fullMsg := baseMsg + " - " + additionalMsg + ": %w"
		allArgs := append(args, err)
		return New(ErrDBGenericError, fullMsg, allArgs...)
	}
	return New(ErrDBGenericError, baseMsg+": %w", err)
}

func NewAccessDenied(resourceType string, reason string, args ...any) *AppError {
	baseMsg := fmt.Sprintf("access denied to %s", resourceType)
	if reason != "" {
		fullMsg := baseMsg + " - " + reason
		if len(args) > 0 {
			return New(ErrAccessDenied, fullMsg, args...)
		}
		return New(ErrAccessDenied, fullMsg)
	}
	return New(ErrAccessDenied, baseMsg)
}

func NewAuthFailed(operation string, reason string, args ...any) *AppError {
	baseMsg := fmt.Sprintf("%s failed", operation)
	if reason != "" {
		fullMsg := baseMsg + " - " + reason
		if len(args) > 0 {
			return New(AuthFailed, fullMsg, args...)
		}
		return New(AuthFailed, fullMsg)
	}
	return New(AuthFailed, baseMsg)
}

func NewValidationFailed(field string, reason string, args ...any) *AppError {
	baseMsg := fmt.Sprintf("validation failed for %s", field)
	if reason != "" {
		fullMsg := baseMsg + " - " + reason
		if len(args) > 0 {
			return New(ErrValidationFailed, fullMsg, args...)
		}
		return New(ErrValidationFailed, fullMsg)
	}
	return New(ErrValidationFailed, baseMsg)
}
