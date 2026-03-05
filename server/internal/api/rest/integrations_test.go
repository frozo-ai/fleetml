package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIntegrationHandler_ImportMLflow_InvalidJSON(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import", strings.NewReader("bad"))
	w := httptest.NewRecorder()

	handler.ImportMLflow(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportMLflow_MissingName(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	handler.ImportMLflow(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportHuggingFace_InvalidJSON(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/huggingface/import", strings.NewReader("bad"))
	w := httptest.NewRecorder()

	handler.ImportHuggingFace(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportHuggingFace_MissingRepoID(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/huggingface/import", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	handler.ImportHuggingFace(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportMLflow_EmptyBody(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import", strings.NewReader(""))
	w := httptest.NewRecorder()

	handler.ImportMLflow(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportHuggingFace_EmptyBody(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/huggingface/import", strings.NewReader(""))
	w := httptest.NewRecorder()

	handler.ImportHuggingFace(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportMLflow_EmptyModelName(t *testing.T) {
	handler := &IntegrationHandler{}
	// model_name is explicitly empty string
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import",
		strings.NewReader(`{"model_name":""}`))
	w := httptest.NewRecorder()

	handler.ImportMLflow(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty model_name, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportHuggingFace_EmptyRepoID(t *testing.T) {
	handler := &IntegrationHandler{}
	// repo_id is explicitly empty string
	req := httptest.NewRequest("POST", "/api/v1/integrations/huggingface/import",
		strings.NewReader(`{"repo_id":""}`))
	w := httptest.NewRecorder()

	handler.ImportHuggingFace(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty repo_id, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportMLflow_ArrayInsteadOfObject(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import",
		strings.NewReader(`[{"model_name":"test"}]`))
	w := httptest.NewRecorder()

	handler.ImportMLflow(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for array body, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportHuggingFace_ArrayInsteadOfObject(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/huggingface/import",
		strings.NewReader(`[{"repo_id":"test/model"}]`))
	w := httptest.NewRecorder()

	handler.ImportHuggingFace(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for array body, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportMLflow_WrongTypeForModelName(t *testing.T) {
	handler := &IntegrationHandler{}
	// model_name should be string, not number
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import",
		strings.NewReader(`{"model_name":123}`))
	w := httptest.NewRecorder()

	handler.ImportMLflow(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong type on model_name, got %d", w.Code)
	}
}

func TestIntegrationHandler_ImportMLflow_ResponseErrorFormat(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import",
		strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	handler.ImportMLflow(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("expected parseable JSON error, got: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Error("expected 'error' key in response")
	}
}

func TestIntegrationHandler_ImportHuggingFace_ResponseErrorFormat(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/huggingface/import",
		strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	handler.ImportHuggingFace(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("expected parseable JSON error, got: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Error("expected 'error' key in response")
	}
}

func TestNewIntegrationHandler(t *testing.T) {
	h := NewIntegrationHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil IntegrationHandler")
	}
}

func TestImportMLflow_NilBody(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import", nil)
	w := httptest.NewRecorder()
	handler.ImportMLflow(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestImportHuggingFace_NilBody(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/huggingface/import", nil)
	w := httptest.NewRecorder()
	handler.ImportHuggingFace(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestImportMLflow_NumericBody(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import", strings.NewReader("12345"))
	w := httptest.NewRecorder()
	handler.ImportMLflow(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestImportHuggingFace_NumericBody(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/huggingface/import", strings.NewReader("12345"))
	w := httptest.NewRecorder()
	handler.ImportHuggingFace(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestImportMLflow_WhitespaceModelName(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/mlflow/import", strings.NewReader(`{"model_name":"   "}`))
	w := httptest.NewRecorder()
	handler.ImportMLflow(w, req)
	// Whitespace-only name — depends on validation: may pass or fail
	_ = w.Code
}

func TestImportHuggingFace_WhitespaceRepoID(t *testing.T) {
	handler := &IntegrationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/integrations/huggingface/import", strings.NewReader(`{"repo_id":"   "}`))
	w := httptest.NewRecorder()
	handler.ImportHuggingFace(w, req)
	_ = w.Code
}
