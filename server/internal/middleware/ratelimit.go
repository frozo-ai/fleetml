package middleware

import (
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RateLimiter implements per-IP token bucket rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // tokens per interval
	interval time.Duration // refill interval
	burst    int           // max burst size
	logger   *zap.SugaredLogger
}

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter. rate is requests per interval.
// burst is the max tokens that can accumulate (allows short bursts).
func NewRateLimiter(rate int, interval time.Duration, burst int, logger *zap.SugaredLogger) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		interval: interval,
		burst:    burst,
		logger:   logger,
	}

	// Clean up stale entries every 5 minutes
	go rl.cleanup()

	return rl
}

// Middleware returns an HTTP middleware that enforces rate limits per IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		// Use X-Real-IP if set (via chi middleware.RealIP)
		if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			ip = realIP
		}

		if !rl.allow(ip) {
			if rl.logger != nil {
				rl.logger.Warnw("rate limit exceeded", "ip", ip, "path", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded","code":"RATE_LIMITED"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]
	if !exists {
		rl.visitors[key] = &visitor{
			tokens:   float64(rl.burst) - 1,
			lastSeen: now,
		}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(v.lastSeen)
	refill := elapsed.Seconds() / rl.interval.Seconds() * float64(rl.rate)
	v.tokens += refill
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}
	v.lastSeen = now

	if v.tokens < 1 {
		return false
	}

	v.tokens--
	return true
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		for key, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// StrictRateLimiter creates a rate limiter suitable for auth endpoints (5 req/sec, burst 10).
func StrictRateLimiter(logger *zap.SugaredLogger) *RateLimiter {
	return NewRateLimiter(5, time.Second, 10, logger)
}

// DefaultRateLimiter creates a rate limiter suitable for general API (100 req/sec, burst 200).
func DefaultRateLimiter(logger *zap.SugaredLogger) *RateLimiter {
	return NewRateLimiter(100, time.Second, 200, logger)
}
