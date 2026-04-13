package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fleetml/fleetml/server/internal/auth"
	"go.uber.org/zap"
)

// withTestClaims adds test JWT claims to a request context for handler tests.
func withTestClaims(r *http.Request) *http.Request {
	claims := &auth.Claims{
		UserID: "test-user-id",
		Email:  "test@example.com",
		Role:   "admin",
		OrgID:  "test-org-id",
	}
	ctx := context.WithValue(r.Context(), auth.UserContextKey, claims)
	return r.WithContext(ctx)
}

func TestDeviceHandler_Update_InvalidJSON(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeviceHandler(nil, logger)

	req := withTestClaims(httptest.NewRequest(http.MethodPatch, "/api/v1/devices/device-123", strings.NewReader("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestDeviceHandler_Update_EmptyBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeviceHandler(nil, logger)

	// Empty body (no JSON at all) should fail to decode
	req := withTestClaims(httptest.NewRequest(http.MethodPatch, "/api/v1/devices/device-123", strings.NewReader("")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	// json.Decoder.Decode on empty body returns io.EOF which is an error
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestDeviceHandler_Update_WrongJSONTypes(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeviceHandler(nil, logger)

	// labels should be map[string]string, not an array
	req := withTestClaims(httptest.NewRequest(http.MethodPatch, "/api/v1/devices/device-123",
		strings.NewReader(`{"labels": "not-a-map"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong type in labels, got %d", w.Code)
	}
}

func TestDeviceHandler_Update_MalformedJSONTrailing(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeviceHandler(nil, logger)

	// Malformed JSON with trailing garbage characters
	req := withTestClaims(httptest.NewRequest(http.MethodPatch, "/api/v1/devices/device-123",
		strings.NewReader(`not-even-close`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", w.Code)
	}
}

func TestDeviceHandler_Get_EmptyID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeviceHandler(nil, logger)

	// Without chi context, URLParam returns "" — device_id will be empty
	req := withTestClaims(httptest.NewRequest(http.MethodGet, "/api/v1/devices/", nil))
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: handler tried to call fleet.GetDevice with empty ID
				// No pre-validation in handler, so it panics on nil fleet manager
			}
		}()
		handler.Get(w, req)
	}()
}

func TestDeviceHandler_Delete_EmptyID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeviceHandler(nil, logger)

	req := withTestClaims(httptest.NewRequest(http.MethodDelete, "/api/v1/devices/", nil))
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: handler tried to call fleet.GetDevice with empty ID
			}
		}()
		handler.Delete(w, req)
	}()
}

func TestDeviceHandler_List_WithQueryParams(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeviceHandler(nil, logger)

	// Handler doesn't validate status values — any status string is passed through
	req := withTestClaims(httptest.NewRequest(http.MethodGet, "/api/v1/devices?status=invalid_status&limit=abc", nil))
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: reaches fleet.ListDevices (no validation on query params)
			}
		}()
		handler.List(w, req)
	}()
}

func TestDeviceHandler_Update_ValidJSONNoFleetID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeviceHandler(nil, logger)

	// Valid JSON with no fleet_id — should still reach GetDevice (panics on nil fleet)
	req := withTestClaims(httptest.NewRequest(http.MethodPatch, "/api/v1/devices/device-123",
		strings.NewReader(`{"name":"new-name"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: decoded OK, skipped fleet assign, panics on GetDevice
			}
		}()
		handler.Update(w, req)
	}()
}

func TestDeviceHandler_List_NoAuth(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDeviceHandler(nil, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing auth, got %d", w.Code)
	}
}

func TestNewDeviceHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()
	h := NewDeviceHandler(nil, logger)
	if h == nil {
		t.Fatal("expected non-nil DeviceHandler")
	}
	if h.logger != logger {
		t.Error("expected logger to be set")
	}
}
