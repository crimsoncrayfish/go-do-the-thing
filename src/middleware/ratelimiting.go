package middleware

import (
	"fmt"
	"go-do-the-thing/src/helpers/logger"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type RateLimitMiddleWare struct {
	current_counts   map[string]int
	current_window   time.Time
	mu               sync.Mutex
	maxRequestsLimit int
	logger           *logger.Logger
}

func NewRateLimiter() *RateLimitMiddleWare {
	log := logger.NewLogger("RateLimitMiddleware")
	maxRequests := 500
	if envLimit := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"); envLimit != "" {
		if parsed, err := strconv.Atoi(envLimit); err == nil && parsed > 0 {
			maxRequests = parsed
		}
	}

	logger.TODO("Consider implementing sliding window algorithm instead of fixed windows")
	logger.TODO("Add circuit breaker logic for graceful degradation under high load")
	logger.TODO("Implement Redis/external store for multi-instance deployments")
	logger.TODO("Add configurable rate limits per endpoint or user type")
	logger.TODO("Add metrics collection (requests/sec, rate limit hits)")
	logger.TODO("Consider rate limiting by user ID for authenticated requests")

	return &RateLimitMiddleWare{
		current_counts:   make(map[string]int),
		maxRequestsLimit: maxRequests,
		logger:           log,
	}
}

func (mw *RateLimitMiddleWare) getRateAndUpdate(ip string, currentWindow time.Time) int {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	if !currentWindow.Equal(mw.current_window) {
		mw.current_counts = make(map[string]int)
		mw.current_window = currentWindow
	}
	var current_count int
	if count, ok := mw.current_counts[ip]; ok {
		current_count = count + 1
	} else {
		current_count = 1
	}
	mw.current_counts[ip] = current_count
	return current_count
}

func (mw *RateLimitMiddleWare) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ReadUserIP(r)
		now := time.Now()
		currentWindow := now.Truncate(time.Minute)
		secondsUntilReset := 60 - now.Second()

		current_count := mw.getRateAndUpdate(ip, currentWindow)

		if current_count > mw.maxRequestsLimit {
			mw.logger.Warn("Ratelimited IP", "ip", ip)
			w.Header().Set("Retry-After", fmt.Sprintf("%d", secondsUntilReset))
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", mw.maxRequestsLimit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		remaining := mw.maxRequestsLimit - current_count
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", mw.maxRequestsLimit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		mw.logger.Debug("Request from IP", "ip", ip, "current_count", current_count, "window", currentWindow.Format("15:04"))
		next.ServeHTTP(w, r)
	})
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {

		IPAddress, _, _ = net.SplitHostPort(r.RemoteAddr)
		if IPAddress == "" {
			IPAddress = r.RemoteAddr
		}
	}
	return IPAddress
}
