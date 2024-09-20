package middleware

import (
	"go-do-the-thing/src/helpers/slog"
	"net/http"
	"time"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

type LoggingMiddleWare struct {
	logger slog.Logger
}

func NewLoggingMiddleWare() *LoggingMiddleWare {
	return &LoggingMiddleWare{
		logger: slog.NewLogger("HTTP"),
	}
}

func (mw *LoggingMiddleWare) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wr := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(wr, r)
		mw.logger.HttpInfo("'METHOD: %s, PATH: %s, EXECUTION_TIME: %s'", wr.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}
