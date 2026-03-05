package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_AllowsInitialBurst(t *testing.T) {
	rl := NewRateLimiter(10, time.Second, 5, nil)

	for i := 0; i < 5; i++ {
		if !rl.allow("192.168.1.1") {
			t.Errorf("request %d should be allowed within burst", i)
		}
	}
}

func TestRateLimiter_BlocksAfterBurst(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 3, nil)

	// Use up the burst
	for i := 0; i < 3; i++ {
		rl.allow("192.168.1.1")
	}

	// Next request should be blocked
	if rl.allow("192.168.1.1") {
		t.Error("expected request to be blocked after burst exhausted")
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 2, nil)

	// Exhaust IP1
	rl.allow("192.168.1.1")
	rl.allow("192.168.1.1")

	// IP2 should still work
	if !rl.allow("192.168.1.2") {
		t.Error("different IP should not be affected by rate limit")
	}
}

func TestRateLimiter_RefillsOverTime(t *testing.T) {
	rl := NewRateLimiter(10, time.Second, 2, nil)

	// Exhaust tokens
	rl.allow("10.0.0.1")
	rl.allow("10.0.0.1")

	if rl.allow("10.0.0.1") {
		t.Error("should be blocked immediately after burst")
	}

	// Simulate time passing by manipulating lastSeen
	rl.mu.Lock()
	rl.visitors["10.0.0.1"].lastSeen = time.Now().Add(-200 * time.Millisecond)
	rl.mu.Unlock()

	// Should be allowed after refill
	if !rl.allow("10.0.0.1") {
		t.Error("should be allowed after token refill")
	}
}

func TestRateLimiter_Middleware_Returns429(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 1, nil)

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request should pass
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", w.Code)
	}

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", w2.Code)
	}
}

func TestRateLimiter_Middleware_UsesXRealIP(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 1, nil)

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request with X-Real-IP
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Same X-Real-IP should be rate limited
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w2.Code)
	}
}

func TestRateLimiter_Middleware_RetryAfterHeader(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 1, nil)

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "1.2.3.4:1234"

	// Exhaust
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Rate limited — check Retry-After header
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Header().Get("Retry-After") != "1" {
		t.Errorf("expected Retry-After: 1, got %s", w2.Header().Get("Retry-After"))
	}
}

func TestStrictRateLimiter(t *testing.T) {
	rl := StrictRateLimiter(nil)
	if rl == nil {
		t.Fatal("expected non-nil rate limiter")
	}
	if rl.rate != 5 {
		t.Errorf("expected rate 5, got %d", rl.rate)
	}
	if rl.burst != 10 {
		t.Errorf("expected burst 10, got %d", rl.burst)
	}
}

func TestDefaultRateLimiter(t *testing.T) {
	rl := DefaultRateLimiter(nil)
	if rl == nil {
		t.Fatal("expected non-nil rate limiter")
	}
	if rl.rate != 100 {
		t.Errorf("expected rate 100, got %d", rl.rate)
	}
	if rl.burst != 200 {
		t.Errorf("expected burst 200, got %d", rl.burst)
	}
}

func TestRateLimiter_TokensCapAtBurst(t *testing.T) {
	rl := NewRateLimiter(100, time.Second, 5, nil)

	// Allow many refill cycles
	rl.mu.Lock()
	rl.visitors["test"] = &visitor{
		tokens:   0,
		lastSeen: time.Now().Add(-10 * time.Second),
	}
	rl.mu.Unlock()

	rl.allow("test")

	// Check that tokens didn't exceed burst
	rl.mu.Lock()
	tokens := rl.visitors["test"].tokens
	rl.mu.Unlock()

	if tokens > float64(rl.burst) {
		t.Errorf("tokens %f exceeded burst %d", tokens, rl.burst)
	}
}
