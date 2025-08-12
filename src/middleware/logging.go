package middleware

import (
	"log/slog"
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
	logger *slog.Logger
}

func NewLoggingMiddleWare() *LoggingMiddleWare {
	logger := slog.With("source", "LoggingMiddleware")
	return &LoggingMiddleWare{
		logger: logger,
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
		mw.logger.Info("HTTP request", "method", r.Method, "path", r.URL.Path, "status", wr.statusCode, "execution_time", time.Since(start))
	})
}
