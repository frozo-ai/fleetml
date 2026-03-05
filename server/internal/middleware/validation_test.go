package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c8", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"not-a-uuid", false},
		{"550e8400-e29b-41d4-a716", false},
		{"", false},
		{"550e8400e29b41d4a716446655440000", false}, // no dashes
		{"550e8400-e29b-41d4-a716-44665544000g", false}, // invalid hex
		{"XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX", false},
	}

	for _, tt := range tests {
		got := IsValidUUID(tt.input)
		if got != tt.valid {
			t.Errorf("IsValidUUID(%q) = %v, want %v", tt.input, got, tt.valid)
		}
	}
}

func TestMaxBodySize_AllowsSmallBody(t *testing.T) {
	handler := MaxBodySize(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader(`{"key": "value"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.ContentLength = int64(body.Len())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMaxBodySize_RejectsLargeContentLength(t *testing.T) {
	handler := MaxBodySize(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("small"))
	req.ContentLength = 200 // declared larger than limit
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", w.Code)
	}
}

func TestRequestSizeLimit(t *testing.T) {
	mw := RequestSizeLimit()
	if mw == nil {
		t.Fatal("expected non-nil middleware")
	}
}

func TestModelUploadSizeLimit(t *testing.T) {
	mw := ModelUploadSizeLimit()
	if mw == nil {
		t.Fatal("expected non-nil middleware")
	}
}

func TestValidateStringLength(t *testing.T) {
	tests := []struct {
		s      string
		min    int
		max    int
		valid  bool
	}{
		{"hello", 1, 10, true},
		{"", 1, 10, false},          // too short
		{"hello world foo", 1, 10, false}, // too long
		{"a", 1, 1, true},           // exact min=max
		{"abc", 0, 100, true},       // within range
	}

	for _, tt := range tests {
		got := ValidateStringLength(tt.s, tt.min, tt.max)
		if got != tt.valid {
			t.Errorf("ValidateStringLength(%q, %d, %d) = %v, want %v",
				tt.s, tt.min, tt.max, got, tt.valid)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":          "DENY",
		"X-XSS-Protection":         "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	}

	for header, expected := range expectedHeaders {
		got := w.Header().Get(header)
		if got != expected {
			t.Errorf("expected %s: %s, got %s", header, expected, got)
		}
	}
}

func TestSecurityHeaders_Passthrough(t *testing.T) {
	called := false
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
