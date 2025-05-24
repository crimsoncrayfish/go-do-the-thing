package errors

import (
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	templ_shared "go-do-the-thing/src/shared/templ"
	"net/http"
)

func FrontendError(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, msg string, a ...any) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, "FrontendError", "user auth failed unsuccessfully")

	w.Header().Set("HX-Retarget", "#toast-message")
	w.WriteHeader(http.StatusInternalServerError)
	logger.Error(err, msg, a...)
	err = templ_shared.ToastMessage(err.Error(), "error").Render(r.Context(), w)
	if err != nil {
		logger.Error(err, "failed to render toast for user %d", current_user_id)
		return
	}
}

func FrontendErrorBadRequest(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, msg string, a ...any) {
	FrontendError(w, r, logger, err, msg, a...)
	w.WriteHeader(http.StatusBadRequest)
}

func FrontendErrorNotFound(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, msg string, a ...any) {
	FrontendError(w, r, logger, err, msg, a...)
	w.WriteHeader(http.StatusNotFound)
}

func FrontendErrorUnauthorized(w http.ResponseWriter, r *http.Request, logger slog.Logger, err error, msg string, a ...any) {
	FrontendError(w, r, logger, err, msg, a...)
	w.WriteHeader(http.StatusUnauthorized)
}
