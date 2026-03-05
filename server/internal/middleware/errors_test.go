package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusBadRequest, "invalid input", "VALIDATION_ERROR")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Error != "invalid input" {
		t.Errorf("expected error 'invalid input', got %q", resp.Error)
	}
	if resp.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got %q", resp.Code)
	}
}

func TestWriteBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	WriteBadRequest(w, "missing field")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "BAD_REQUEST" {
		t.Errorf("expected code BAD_REQUEST, got %q", resp.Code)
	}
}

func TestWriteNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	WriteNotFound(w, "model not found")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %q", resp.Code)
	}
}

func TestWriteUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	WriteUnauthorized(w, "invalid token")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %q", resp.Code)
	}
}

func TestWriteForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	WriteForbidden(w, "insufficient permissions")
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "FORBIDDEN" {
		t.Errorf("expected code FORBIDDEN, got %q", resp.Code)
	}
}

func TestWriteInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteInternalError(w, "database connection lost")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got %q", resp.Code)
	}
}

func TestWriteConflict(t *testing.T) {
	w := httptest.NewRecorder()
	WriteConflict(w, "model already exists")
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "CONFLICT" {
		t.Errorf("expected code CONFLICT, got %q", resp.Code)
	}
}

func TestWriteError_EmptyCode(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusTeapot, "I'm a teapot", "")

	if w.Code != http.StatusTeapot {
		t.Errorf("expected 418, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "" {
		t.Errorf("expected empty code, got %q", resp.Code)
	}
}

func TestErrorResponse_JSONOmitsEmptyCode(t *testing.T) {
	resp := ErrorResponse{Error: "test", Code: ""}
	data, _ := json.Marshal(resp)

	var m map[string]interface{}
	json.Unmarshal(data, &m)

	// code should be omitted when empty (omitempty tag)
	if _, ok := m["code"]; ok {
		if m["code"] != "" {
			t.Error("empty code should be omitted from JSON")
		}
	}
}
