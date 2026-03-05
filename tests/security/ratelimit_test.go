package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mw "github.com/fleetml/fleetml/server/internal/middleware"
	"go.uber.org/zap"
)

// TestRateLimiting_BruteForceProtection validates that rapid auth requests are blocked.
func TestRateLimiting_BruteForceProtection(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := mw.StrictRateLimiter(logger.Sugar())

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Strict limiter: 5 req/sec, burst 10
	// Send 15 requests — first 10 should succeed (burst), rest should be rate limited
	successCount := 0
	limitedCount := 0

	for i := 0; i < 15; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.100:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			successCount++
		} else if rr.Code == http.StatusTooManyRequests {
			limitedCount++
		}
	}

	if successCount != 10 {
		t.Errorf("expected 10 successful requests (burst size), got %d", successCount)
	}
	if limitedCount != 5 {
		t.Errorf("expected 5 rate-limited requests, got %d", limitedCount)
	}
}

// TestRateLimiting_PerIPIsolation validates that IPs are rate-limited independently.
func TestRateLimiting_PerIPIsolation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := mw.StrictRateLimiter(logger.Sugar())

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// IP1: exhaust burst
	for i := 0; i < 12; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// IP2: should still have full burst available
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.RemoteAddr = "10.0.0.2:5678"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("IP2 should not be rate limited, got %d", rr.Code)
	}
}

// TestRateLimiting_RetryAfterHeader validates 429 response includes Retry-After.
func TestRateLimiting_RetryAfterHeader(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := mw.StrictRateLimiter(logger.Sugar())

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust burst
	for i := 0; i < 11; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "10.0.0.50:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Next request should be rate limited with Retry-After
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.RemoteAddr = "10.0.0.50:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}

	retryAfter := rr.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header on 429 response")
	}
}

// TestRateLimiting_Recovery validates tokens refill after waiting.
func TestRateLimiting_Recovery(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	// 10 req/sec, burst 5 — fast refill for testing
	rl := mw.NewRateLimiter(10, time.Second, 5, logger.Sugar())

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust all tokens
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("GET", "/api/v1/models", nil)
		req.RemoteAddr = "10.0.0.99:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Wait for refill
	time.Sleep(200 * time.Millisecond)

	// Should be able to make requests again
	req := httptest.NewRequest("GET", "/api/v1/models", nil)
	req.RemoteAddr = "10.0.0.99:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 after token refill, got %d", rr.Code)
	}
}

// TestRateLimiting_DefaultVsStrict validates different rate limits for different endpoints.
func TestRateLimiting_DefaultVsStrict(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	strict := mw.StrictRateLimiter(logger.Sugar())
	general := mw.DefaultRateLimiter(logger.Sugar())

	strictHandler := strict.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	generalHandler := general.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Strict should be exhausted after 10 requests (burst 10)
	strictLimited := false
	for i := 0; i < 15; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "10.0.0.200:1234"
		rr := httptest.NewRecorder()
		strictHandler.ServeHTTP(rr, req)
		if rr.Code == http.StatusTooManyRequests {
			strictLimited = true
			break
		}
	}
	if !strictLimited {
		t.Error("strict rate limiter should have limited after 10 requests")
	}

	// General should handle 15 requests fine (burst 200)
	generalLimited := false
	for i := 0; i < 15; i++ {
		req := httptest.NewRequest("GET", "/api/v1/models", nil)
		req.RemoteAddr = "10.0.0.201:1234"
		rr := httptest.NewRecorder()
		generalHandler.ServeHTTP(rr, req)
		if rr.Code == http.StatusTooManyRequests {
			generalLimited = true
			break
		}
	}
	if generalLimited {
		t.Error("general rate limiter should NOT have limited after 15 requests")
	}
}
