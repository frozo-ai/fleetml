package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestInit_Disabled(t *testing.T) {
	cfg := Config{Enabled: false}
	logger := zap.NewNop().Sugar()
	provider, err := Init(context.Background(), cfg, "test", logger)
	if err != nil {
		t.Fatalf("Init disabled: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil provider even when disabled")
	}
	if provider.tp != nil {
		t.Error("expected nil tracer provider when disabled")
	}
	// Shutdown should be safe on disabled provider
	if err := provider.Shutdown(context.Background()); err != nil {
		t.Errorf("shutdown disabled: %v", err)
	}
}

func TestHTTPMiddleware_SetsStatusCode(t *testing.T) {
	handler := HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))

	req := httptest.NewRequest("POST", "/api/v1/models", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
	if rr.Body.String() != "created" {
		t.Errorf("expected body 'created', got %q", rr.Body.String())
	}
}

func TestHTTPMiddleware_Default200(t *testing.T) {
	handler := HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHTTPMiddleware_Error(t *testing.T) {
	handler := HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"fail"}`))
	}))

	req := httptest.NewRequest("GET", "/api/v1/models", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()
	childCtx, span := SpanFromContext(ctx, "test-operation")
	defer span.End()

	if childCtx == nil {
		t.Error("expected non-nil context")
	}
	if span == nil {
		t.Error("expected non-nil span")
	}
}

func TestStatusWriter_WriteHeaderOnce(t *testing.T) {
	rr := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rr, statusCode: http.StatusOK}

	sw.WriteHeader(http.StatusNotFound)
	sw.WriteHeader(http.StatusOK) // should be ignored

	if sw.statusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", sw.statusCode)
	}
}

func TestStatusWriter_WriteDefaultStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rr, statusCode: http.StatusOK}

	sw.Write([]byte("hello"))

	if !sw.written {
		t.Error("expected written to be true after Write")
	}
	if sw.statusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", sw.statusCode)
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}
	if cfg.Enabled {
		t.Error("expected disabled by default")
	}
	if cfg.ServiceName != "" {
		t.Error("expected empty service name by default")
	}
	if cfg.SampleRate != 0 {
		t.Error("expected zero sample rate by default")
	}
}

func TestTracer_ReturnsNonNil(t *testing.T) {
	tracer := Tracer()
	if tracer == nil {
		t.Error("expected non-nil tracer")
	}
}
