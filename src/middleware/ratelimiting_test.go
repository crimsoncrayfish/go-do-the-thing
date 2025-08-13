package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockTimeProvider struct {
	currentTime time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.currentTime
}

func (m *MockTimeProvider) SetTime(t time.Time) {
	m.currentTime = t
}

func (m *MockTimeProvider) AdvanceTime(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

func TestRateLimitBasicFunctionality(t *testing.T) {
	mockTime := &MockTimeProvider{}
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	mockTime.SetTime(baseTime)

	rateLimiter := newRateLimiterWithTimeProvider(mockTime)
	rateLimiter.maxRequestsLimit = 3

	handler := rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed, got status %d", i, w.Code)
		}

		remaining := w.Header().Get("X-RateLimit-Remaining")
		if i == 1 && remaining != "2" {
			t.Errorf("Request %d: expected remaining 2, got %s", i, remaining)
		}
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request over limit should return 429, got %d", w.Code)
	}

	if remaining := w.Header().Get("X-RateLimit-Remaining"); remaining != "0" {
		t.Errorf("Over-limit request should show 0 remaining, got %s", remaining)
	}
}

func TestRateLimitWindowReset(t *testing.T) {
	mockTime := &MockTimeProvider{}
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	mockTime.SetTime(baseTime)

	rateLimiter := newRateLimiterWithTimeProvider(mockTime)
	rateLimiter.maxRequestsLimit = 2

	handler := rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := range 2 {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed, got status %d", i+1, w.Code)
		}
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request over limit should return 429, got %d", w.Code)
	}

	mockTime.AdvanceTime(time.Minute)

	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Request after window reset should succeed, got status %d", w.Code)
	}
}

func TestRateLimitDifferentIPs(t *testing.T) {
	mockTime := &MockTimeProvider{}
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	mockTime.SetTime(baseTime)

	rateLimiter := newRateLimiterWithTimeProvider(mockTime)
	rateLimiter.maxRequestsLimit = 1

	handler := rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First IP first request should succeed, got status %d", w1.Code)
	}

	req1_2 := httptest.NewRequest("GET", "/test", nil)
	req1_2.RemoteAddr = "192.168.1.1:12345"
	w1_2 := httptest.NewRecorder()
	handler.ServeHTTP(w1_2, req1_2)

	if w1_2.Code != http.StatusTooManyRequests {
		t.Errorf("First IP second request should be rate limited, got status %d", w1_2.Code)
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second IP request should succeed, got status %d", w2.Code)
	}
}

func TestReadUserIP(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		xRealIP       string
		xForwardedFor string
		expectedIP    string
	}{
		{
			name:       "X-Real-Ip header",
			remoteAddr: "192.168.1.1:12345",
			xRealIP:    "10.0.0.1",
			expectedIP: "10.0.0.1",
		},
		{
			name:          "X-Forwarded-For header",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "10.0.0.2",
			expectedIP:    "10.0.0.2",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.xRealIP != "" {
				req.Header.Set("X-Real-Ip", tt.xRealIP)
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			ip := ReadUserIP(req)
			if ip != tt.expectedIP {
				t.Errorf("ReadUserIP() = %v, want %v", ip, tt.expectedIP)
			}
		})
	}
}

func TestRateLimitHeaders(t *testing.T) {
	mockTime := &MockTimeProvider{}
	baseTime := time.Date(2024, 1, 1, 12, 0, 30, 0, time.UTC)
	mockTime.SetTime(baseTime)

	rateLimiter := newRateLimiterWithTimeProvider(mockTime)
	rateLimiter.maxRequestsLimit = 5

	handler := rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if limit := w.Header().Get("X-RateLimit-Limit"); limit != "5" {
		t.Errorf("X-RateLimit-Limit should be 5, got %s", limit)
	}

	if remaining := w.Header().Get("X-RateLimit-Remaining"); remaining != "4" {
		t.Errorf("X-RateLimit-Remaining should be 4, got %s", remaining)
	}

	for range 5 {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if retryAfter := w.Header().Get("Retry-After"); retryAfter != "30" {
		t.Errorf("Retry-After should be 30 (seconds until reset), got %s", retryAfter)
	}

	if remaining := w.Header().Get("X-RateLimit-Remaining"); remaining != "0" {
		t.Errorf("X-RateLimit-Remaining should be 0 when rate limited, got %s", remaining)
	}
}

func BenchmarkRateLimitConcurrency(b *testing.B) {
	mockTime := &MockTimeProvider{}
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	mockTime.SetTime(baseTime)

	rateLimiter := newRateLimiterWithTimeProvider(mockTime)
	rateLimiter.maxRequestsLimit = 10000

	handler := rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", (i%254)+1)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			i++
		}
	})
}

func BenchmarkRateLimitSingleIP(b *testing.B) {
	mockTime := &MockTimeProvider{}
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	mockTime.SetTime(baseTime)

	rateLimiter := newRateLimiterWithTimeProvider(mockTime)
	rateLimiter.maxRequestsLimit = 10000

	handler := rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
		}
	})
}
