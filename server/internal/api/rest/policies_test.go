package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestPolicyHandler_Create_InvalidJSON(t *testing.T) {
	handler := &PolicyHandler{}
	req := httptest.NewRequest("POST", "/api/v1/policies", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPolicyHandler_Get_InvalidUUID(t *testing.T) {
	handler := &PolicyHandler{}
	req := httptest.NewRequest("GET", "/api/v1/policies/not-a-uuid", nil)
	w := httptest.NewRecorder()

	// chi URL param won't be set without chi context, but IsValidUUID will catch it
	handler.Get(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPolicyHandler_Update_InvalidJSON(t *testing.T) {
	handler := &PolicyHandler{}
	req := httptest.NewRequest("PATCH", "/api/v1/policies/not-a-uuid", strings.NewReader("bad"))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	// Will fail on UUID validation first
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPolicyHandler_Delete_InvalidUUID(t *testing.T) {
	handler := &PolicyHandler{}
	req := httptest.NewRequest("DELETE", "/api/v1/policies/not-a-uuid", nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPolicyHandler_Create_EmptyBody(t *testing.T) {
	handler := &PolicyHandler{}
	req := httptest.NewRequest("POST", "/api/v1/policies", strings.NewReader(""))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestPolicyHandler_Create_ArrayInsteadOfObject(t *testing.T) {
	handler := &PolicyHandler{}
	req := httptest.NewRequest("POST", "/api/v1/policies", strings.NewReader(`[{"name":"test"}]`))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for array body, got %d", w.Code)
	}
}

func TestPolicyHandler_Create_WrongTypeForName(t *testing.T) {
	handler := &PolicyHandler{}
	// name should be string, not a number
	req := httptest.NewRequest("POST", "/api/v1/policies", strings.NewReader(`{"name":123}`))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong type on name, got %d", w.Code)
	}
}

func TestPolicyHandler_Update_EmptyBody(t *testing.T) {
	handler := &PolicyHandler{}
	// UUID is invalid (no chi context), so fails on UUID validation before body parse
	req := httptest.NewRequest("PATCH", "/api/v1/policies/not-a-uuid", strings.NewReader(""))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", w.Code)
	}
}

func TestPolicyHandler_Update_ValidUUID_InvalidJSON(t *testing.T) {
	handler := &PolicyHandler{}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "550e8400-e29b-41d4-a716-446655440000")

	req := httptest.NewRequest("PATCH", "/api/v1/policies/550e8400-e29b-41d4-a716-446655440000",
		strings.NewReader("not json"))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON with valid UUID, got %d", w.Code)
	}
}

func TestPolicyHandler_Update_ValidUUID_EmptyBody(t *testing.T) {
	handler := &PolicyHandler{}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "550e8400-e29b-41d4-a716-446655440000")

	req := httptest.NewRequest("PATCH", "/api/v1/policies/550e8400-e29b-41d4-a716-446655440000",
		strings.NewReader(""))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body with valid UUID, got %d", w.Code)
	}
}

func TestPolicyHandler_Get_EmptyID(t *testing.T) {
	handler := &PolicyHandler{}
	req := httptest.NewRequest("GET", "/api/v1/policies/", nil)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty ID, got %d", w.Code)
	}
}

func TestPolicyHandler_Delete_EmptyID(t *testing.T) {
	handler := &PolicyHandler{}
	req := httptest.NewRequest("DELETE", "/api/v1/policies/", nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty ID, got %d", w.Code)
	}
}

func TestPolicyHandler_Create_ResponseErrorFormat(t *testing.T) {
	handler := &PolicyHandler{}
	req := httptest.NewRequest("POST", "/api/v1/policies", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("expected parseable JSON error, got: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Error("expected 'error' key in response")
	}
}

func TestNewPolicyHandler(t *testing.T) {
	h := NewPolicyHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil PolicyHandler")
	}
}
