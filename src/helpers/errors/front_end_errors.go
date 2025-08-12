package errors

import (
	"go-do-the-thing/src/helpers/logger"
	"net/http"
)

func BadRequest(logger *logger.Logger, w http.ResponseWriter, r *http.Request, err error, msg string, a ...any) {
	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusBadRequest)
	logger.Error(msg, "error", err)

	userMessage := getUserFriendlyMessage(err, msg)
	_ = ToastMessage(userMessage, "error").Render(r.Context(), w)
}

func NotFound(logger *logger.Logger, w http.ResponseWriter, r *http.Request, err error, msg string, a ...any) {
	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusNotFound)
	logger.Error(msg, "error", err)

	userMessage := getUserFriendlyMessage(err, msg)
	_ = ToastMessage(userMessage, "error").Render(r.Context(), w)
}

func Unauthorized(logger *logger.Logger, w http.ResponseWriter, r *http.Request, err error, msg string, a ...any) {
	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusUnauthorized)
	logger.Error(msg, "error", err)

	userMessage := getUserFriendlyMessage(err, msg)
	_ = ToastMessage(userMessage, "error").Render(r.Context(), w)
}

func InternalServerError(logger *logger.Logger, w http.ResponseWriter, r *http.Request, err error, msg string, a ...any) {
	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusInternalServerError)
	logger.Error(msg, "error", err)

	userMessage := getUserFriendlyMessage(err, msg)
	_ = ToastMessage(userMessage, "error").Render(r.Context(), w)
}

func getUserFriendlyMessage(err error, defaultMsg string) string {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Code() {
		case ErrAccessDenied:
			return "You don't have permission to perform this action"
		case ErrPermissionDenied:
			return "You don't have permission to access this resource"
		case ErrNotFound:
			return "The requested resource was not found"
		case ErrDBReadFailed:
			return "Unable to retrieve data. Please try again"
		case ErrDBInsertFailed:
			return "Unable to save data. Please try again"
		case ErrDBUpdateFailed:
			return "Unable to update data. Please try again"
		case ErrDBDeleteFailed:
			return "Unable to delete data. Please try again"
		case ErrDBGenericError:
			return "An error occurred. Please try again"
		}
	}

	return defaultMsg
}
