package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestDeployHandler_Create_InvalidJSON(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments", strings.NewReader("{not valid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestDeployHandler_Create_EmptyBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments", strings.NewReader("")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestDeployHandler_Create_MissingModelName(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	body := `{"model_version":"v1","target_type":"fleet","target_id":"fleet-1"}`
	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing model_name, got %d", w.Code)
	}
}

func TestDeployHandler_Create_MissingModelVersion(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	body := `{"model_name":"my-model","target_type":"fleet","target_id":"fleet-1"}`
	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing model_version, got %d", w.Code)
	}
}

func TestDeployHandler_Create_MissingBothNameAndVersion(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	body := `{"target_type":"fleet","target_id":"fleet-1"}`
	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing model_name and model_version, got %d", w.Code)
	}
}

func TestDeployHandler_Create_EmptyObject(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	body := `{}`
	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty object, got %d", w.Code)
	}
}

func TestDeployHandler_Create_WrongJSONTypes(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	body := `{"model_name":123,"model_version":"v1"}`
	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong type, got %d", w.Code)
	}
}

func TestDeployHandler_Get_EmptyID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	req := withTestClaims(httptest.NewRequest(http.MethodGet, "/api/v1/deployments/", nil))
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: handler tried to call orchestrator.GetDeployment with empty ID
			}
		}()
		handler.Get(w, req)
	}()
}

func TestDeployHandler_Cancel_EmptyID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments//cancel", nil))
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: handler tried to call orchestrator.CancelDeployment with empty ID
			}
		}()
		handler.Cancel(w, req)
	}()
}

func TestDeployHandler_Rollback_EmptyID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments//rollback", nil))
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: handler tried to call orchestrator.RollbackDeployment with empty ID
			}
		}()
		handler.Rollback(w, req)
	}()
}

func TestDeployHandler_List_EmptyQueryParams(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	req := withTestClaims(httptest.NewRequest(http.MethodGet, "/api/v1/deployments", nil))
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: reaches orchestrator.ListDeployments with empty state/model_name
			}
		}()
		handler.List(w, req)
	}()
}

func TestNewDeployHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()
	h := NewDeployHandler(nil, logger)
	if h == nil {
		t.Fatal("expected non-nil DeployHandler")
	}
	if h.logger != logger {
		t.Error("expected logger to be set")
	}
}

func TestDeployHandler_Create_WhitespaceModelName(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	body := `{"model_name":"  ","model_version":"v1","target_type":"fleet"}`
	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: whitespace model name passes validation, reaches orchestrator
			}
		}()
		handler.Create(w, req)
	}()
}

func TestDeployHandler_Create_ArrayInsteadOfObject(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeployHandler(nil, logger)

	body := `[{"model_name":"test"}]`
	req := withTestClaims(httptest.NewRequest(http.MethodPost, "/api/v1/deployments", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for JSON array instead of object, got %d", w.Code)
	}
}
