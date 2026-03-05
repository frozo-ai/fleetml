package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestABTestHandler_Create_MissingFields(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", `{}`},
		{"missing name", `{"model_a_id":"a","model_b_id":"b"}`},
		{"missing model_a", `{"name":"test","model_b_id":"b"}`},
		{"missing model_b", `{"name":"test","model_a_id":"a"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/ab-tests", bytes.NewReader([]byte(tt.body)))
			rr := httptest.NewRecorder()
			handler.Create(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}

			var resp map[string]string
			json.NewDecoder(rr.Body).Decode(&resp)
			if resp["error"] == "" {
				t.Error("expected error message in response")
			}
		})
	}
}

func TestABTestHandler_Create_InvalidJSON(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	req := httptest.NewRequest("POST", "/api/v1/ab-tests", bytes.NewReader([]byte("not json")))
	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestABTestHandler_Get_InvalidUUID(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	// Simulate chi URL param by setting it directly isn't possible without chi context,
	// but we can test the UUID validation by calling the handler with a path that
	// would produce an invalid UUID param
	req := httptest.NewRequest("GET", "/api/v1/ab-tests/not-a-uuid", nil)
	rr := httptest.NewRecorder()

	// Without chi context, URLParam returns "" which fails IsValidUUID
	handler.Get(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", rr.Code)
	}
}

func TestABTestHandler_Stop_InvalidUUID(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	req := httptest.NewRequest("POST", "/api/v1/ab-tests/invalid/stop", bytes.NewReader([]byte(`{}`)))
	rr := httptest.NewRecorder()

	handler.Stop(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", rr.Code)
	}
}

func TestABTestHandler_Create_EmptyBody(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	req := httptest.NewRequest("POST", "/api/v1/ab-tests", strings.NewReader(""))
	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", rr.Code)
	}
}

func TestABTestHandler_Create_EmptyNameString(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	// Name is explicitly empty string
	body := `{"name":"","model_a_id":"a","model_b_id":"b"}`
	req := httptest.NewRequest("POST", "/api/v1/ab-tests", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty name string, got %d", rr.Code)
	}
}

func TestABTestHandler_Create_EmptyModelIDs(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	// model_a_id and model_b_id are empty strings
	body := `{"name":"test","model_a_id":"","model_b_id":""}`
	req := httptest.NewRequest("POST", "/api/v1/ab-tests", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty model IDs, got %d", rr.Code)
	}
}

func TestABTestHandler_Create_ArrayInsteadOfObject(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	body := `[{"name":"test","model_a_id":"a","model_b_id":"b"}]`
	req := httptest.NewRequest("POST", "/api/v1/ab-tests", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for array instead of object, got %d", rr.Code)
	}
}

func TestABTestHandler_Create_WrongTypeForName(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	// name should be string, not number
	body := `{"name":123,"model_a_id":"a","model_b_id":"b"}`
	req := httptest.NewRequest("POST", "/api/v1/ab-tests", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong type on name, got %d", rr.Code)
	}
}

func TestABTestHandler_Create_ResponseErrorFormat(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	req := httptest.NewRequest("POST", "/api/v1/ab-tests", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Errorf("expected parseable JSON error, got: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Error("expected 'error' key in response")
	}
}

func TestABTestHandler_Get_EmptyID(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	// Without chi context, URLParam returns "" which fails IsValidUUID
	req := httptest.NewRequest("GET", "/api/v1/ab-tests/", nil)
	rr := httptest.NewRecorder()
	handler.Get(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty UUID, got %d", rr.Code)
	}
}

func TestABTestHandler_Stop_EmptyBody(t *testing.T) {
	handler := NewABTestHandler(nil, nil)

	// Stop accepts optional body — with no chi context, UUID validation fails first
	req := httptest.NewRequest("POST", "/api/v1/ab-tests//stop", strings.NewReader(""))
	rr := httptest.NewRecorder()
	handler.Stop(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID (empty), got %d", rr.Code)
	}
}

func TestNewABTestHandler(t *testing.T) {
	h := NewABTestHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil ABTestHandler")
	}
}
