package middleware

import (
	"go-do-the-thing/helpers/slog"
	"net/http"
)

type RateLimitMiddleWare struct {
	logger slog.Logger
}

func NewRateLimiter() *RateLimitMiddleWare {
	return &RateLimitMiddleWare{
		logger: slog.NewLogger("Ratelimiter"),
	}
}

func (mw *RateLimitMiddleWare) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement ratelimiting
		ip := ReadUserIP(r)
		mw.logger.Info("Request from IP: %s", ip)
		next.ServeHTTP(w, r)
	})
}

// TODO: check if this is correct
// From stackoverflow: https://stackoverflow.com/questions/27234861/correct-way-of-getting-clients-ip-addresses-from-http-request
func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
