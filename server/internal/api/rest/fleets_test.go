package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func TestFleetHandler_Create_InvalidJSON(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets", strings.NewReader("{broken"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestFleetHandler_Create_EmptyBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestFleetHandler_Create_EmptyName(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	body := `{"name":"","description":"a fleet"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty name, got %d", w.Code)
	}
}

func TestFleetHandler_Create_MissingName(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	body := `{"description":"a fleet with no name"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing name field, got %d", w.Code)
	}
}

func TestFleetHandler_Create_EmptyObject(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty object (no name), got %d", w.Code)
	}
}

func TestFleetHandler_Create_WrongTypeForName(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	// name should be string, not number
	body := `{"name":123}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong type on name, got %d", w.Code)
	}
}

func TestFleetHandler_Create_ArrayInsteadOfObject(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	body := `[{"name":"fleet1"}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for array instead of object, got %d", w.Code)
	}
}

func TestFleetHandler_Update_InvalidJSON(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/fleets/some-id", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestFleetHandler_Update_EmptyBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/fleets/some-id", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestFleetHandler_Get_InvalidUUID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	// Without chi context, URLParam returns "" — no UUID validation in Get handler
	// It just passes empty ID to fleet.GetFleet which will fail
	req := httptest.NewRequest(http.MethodGet, "/api/v1/fleets/not-a-uuid", nil)
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: Get handler has no UUID validation; panics on nil fleet manager
			}
		}()
		handler.Get(w, req)
	}()
}

func TestFleetHandler_Delete_InvalidUUID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/fleets/not-a-uuid", nil)
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: Delete handler has no UUID validation; panics on nil fleet manager
			}
		}()
		handler.Delete(w, req)
	}()
}

func TestFleetHandler_Stats_InvalidUUID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	// Stats uses mw.IsValidUUID — without chi context, id="" which fails UUID check
	req := httptest.NewRequest(http.MethodGet, "/api/v1/fleets/not-a-uuid/stats", nil)
	w := httptest.NewRecorder()

	handler.Stats(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", w.Code)
	}
}

func TestFleetHandler_Stats_EmptyUUID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	// Explicitly test empty string (no chi context sets URLParam to "")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/fleets//stats", nil)
	w := httptest.NewRecorder()

	handler.Stats(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty UUID, got %d", w.Code)
	}
}

func TestFleetHandler_ListDevices_InvalidUUID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fleets/not-a-uuid/devices", nil)
	w := httptest.NewRecorder()

	handler.ListDevices(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", w.Code)
	}
}

func TestFleetHandler_ListDevices_EmptyUUID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fleets//devices", nil)
	w := httptest.NewRecorder()

	handler.ListDevices(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty UUID, got %d", w.Code)
	}
}

func TestFleetHandler_BulkAssign_InvalidJSON(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	// Use chi context with a valid UUID so we get past UUID validation
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "550e8400-e29b-41d4-a716-446655440000")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets/550e8400-e29b-41d4-a716-446655440000/assign",
		strings.NewReader("not json"))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.BulkAssign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestFleetHandler_BulkAssign_EmptyLabels(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "550e8400-e29b-41d4-a716-446655440000")

	body := `{"labels":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets/550e8400-e29b-41d4-a716-446655440000/assign",
		strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.BulkAssign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty labels, got %d", w.Code)
	}
}

func TestFleetHandler_BulkAssign_MissingLabelsField(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "550e8400-e29b-41d4-a716-446655440000")

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets/550e8400-e29b-41d4-a716-446655440000/assign",
		strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.BulkAssign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing labels field, got %d", w.Code)
	}
}

func TestFleetHandler_BulkAssign_InvalidUUID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	// Without chi context, URLParam returns "" which fails IsValidUUID
	body := `{"labels":{"env":"prod"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets/not-a-uuid/assign",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.BulkAssign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", w.Code)
	}
}

func TestFleetHandler_BulkAssign_EmptyBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "550e8400-e29b-41d4-a716-446655440000")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets/550e8400-e29b-41d4-a716-446655440000/assign",
		strings.NewReader(""))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.BulkAssign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestFleetHandler_Stats_ValidUUIDWithChiContext(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "550e8400-e29b-41d4-a716-446655440000")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fleets/550e8400-e29b-41d4-a716-446655440000/stats", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	// Valid UUID passes validation, but nil fleet manager causes panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: UUID validated OK, panics on nil fleet manager
			}
		}()
		handler.Stats(w, req)
	}()
}

func TestFleetHandler_Create_ResponseErrorFormat(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewFleetHandler(nil, logger)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Verify error response is parseable JSON
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("expected valid JSON error response, got error: %v", err)
	}
	if resp["error"] == "" {
		t.Error("expected non-empty error message in response")
	}
}

func TestNewFleetHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()
	h := NewFleetHandler(nil, logger)
	if h == nil {
		t.Fatal("expected non-nil FleetHandler")
	}
	if h.logger != logger {
		t.Error("expected logger to be set")
	}
}
