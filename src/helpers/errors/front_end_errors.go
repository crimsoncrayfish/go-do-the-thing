package errors

import (
	"go-do-the-thing/src/helpers/slog"
	templ_shared "go-do-the-thing/src/shared/templ"
	"net/http"
)

func BadRequest(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, msg string, a ...any) {
	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusBadRequest)
	logger.Error(err, msg, a...)

	userMessage := getUserFriendlyMessage(err, msg)
	_ = templ_shared.ToastMessage(userMessage, "error").Render(r.Context(), w)
}

func NotFound(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, msg string, a ...any) {
	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusNotFound)
	logger.Error(err, msg, a...)

	userMessage := getUserFriendlyMessage(err, msg)
	_ = templ_shared.ToastMessage(userMessage, "error").Render(r.Context(), w)
}

func Unauthorized(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, msg string, a ...any) {
	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusUnauthorized)
	logger.Error(err, msg, a...)

	userMessage := getUserFriendlyMessage(err, msg)
	_ = templ_shared.ToastMessage(userMessage, "error").Render(r.Context(), w)
}

func InternalServerError(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, msg string, a ...any) {
	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusInternalServerError)
	logger.Error(err, msg, a...)

	userMessage := getUserFriendlyMessage(err, msg)
	_ = templ_shared.ToastMessage(userMessage, "error").Render(r.Context(), w)
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
