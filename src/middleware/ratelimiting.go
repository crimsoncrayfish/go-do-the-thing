package middleware

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"go-do-the-thing/src/helpers/logger"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type TimeProvider interface {
	Now() time.Time
}

type RealTimeProvider struct{}

func (r RealTimeProvider) Now() time.Time {
	return time.Now()
}

type rateLimitShard struct {
	current_counts map[string]int
	current_window time.Time
	mu             sync.RWMutex
}

type RateLimitMiddleWare struct {
	shards           []*rateLimitShard
	maxRequestsLimit int
	shardCount       int
	logger           *logger.Logger
	timeProvider     TimeProvider
}

func NewRateLimiter() *RateLimitMiddleWare {
	return newRateLimiterWithTimeProvider(RealTimeProvider{})
}

func newRateLimiterWithTimeProvider(timeProvider TimeProvider) *RateLimitMiddleWare {
	log := logger.NewLogger("RateLimitMiddleware")
	maxRequests := 500
	if envLimit := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"); envLimit != "" {
		if parsed, err := strconv.Atoi(envLimit); err == nil && parsed > 0 {
			maxRequests = parsed
		}
	}

	// Configure shard count - default to 16 for good distribution
	shardCount := 16
	if envShards := os.Getenv("RATE_LIMIT_SHARD_COUNT"); envShards != "" {
		if parsed, err := strconv.Atoi(envShards); err == nil && parsed > 0 && parsed <= 256 {
			shardCount = parsed
		}
	}

	// Initialize shards
	shards := make([]*rateLimitShard, shardCount)
	for i := range shardCount {
		shards[i] = &rateLimitShard{
			current_counts: make(map[string]int),
		}
	}

	return &RateLimitMiddleWare{
		shards:           shards,
		maxRequestsLimit: maxRequests,
		shardCount:       shardCount,
		logger:           log,
		timeProvider:     timeProvider,
	}
}

func (mw *RateLimitMiddleWare) getRateAndUpdate(ip string, currentWindow time.Time) int {
	shardIndex := getShardIndex(ip, mw.shardCount)
	shard := mw.shards[shardIndex]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Check if this shard needs a window reset
	if !currentWindow.Equal(shard.current_window) {
		// Reset this shard's counts for the new window
		shard.current_counts = make(map[string]int)
		shard.current_window = currentWindow
	}

	var current_count int
	if count, ok := shard.current_counts[ip]; ok {
		current_count = count + 1
	} else {
		current_count = 1
	}
	shard.current_counts[ip] = current_count
	return current_count
}

func (mw *RateLimitMiddleWare) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ReadUserIP(r)
		now := mw.timeProvider.Now()
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

func getShardIndex(ip string, shardCount int) int {
	hashbytes := sha256.Sum256([]byte(ip))
	// Use first 4 bytes for better distribution
	n := int(binary.BigEndian.Uint32(hashbytes[:4]))
	return n % shardCount
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
