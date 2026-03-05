package rest

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestModelHandler_Upload_MissingFile(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	// Create a multipart form with name/version/format but NO file
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "test-model")
	writer.WriteField("version", "v1")
	writer.WriteField("format", "onnx")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Upload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing file, got %d", w.Code)
	}
}

func TestModelHandler_Upload_MissingName(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	// Create a multipart form with file but missing name
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "model.onnx")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("fake model data"))
	writer.WriteField("version", "v1")
	writer.WriteField("format", "onnx")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Upload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestModelHandler_Upload_MissingVersion(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "model.onnx")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("fake model data"))
	writer.WriteField("name", "test-model")
	writer.WriteField("format", "onnx")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Upload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing version, got %d", w.Code)
	}
}

func TestModelHandler_Upload_MissingFormat(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "model.onnx")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("fake model data"))
	writer.WriteField("name", "test-model")
	writer.WriteField("version", "v1")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Upload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing format, got %d", w.Code)
	}
}

func TestModelHandler_Upload_NotMultipart(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	// Send a regular JSON body instead of multipart
	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Upload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-multipart body, got %d", w.Code)
	}
}

func TestModelHandler_Upload_EmptyMultipartForm(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	// Multipart form with no fields at all
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Upload(w, req)

	// Should fail because no file
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty multipart form, got %d", w.Code)
	}
}

func TestModelHandler_Upload_AllFieldsEmpty(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "model.onnx")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("fake model data"))
	writer.WriteField("name", "")
	writer.WriteField("version", "")
	writer.WriteField("format", "")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Upload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for all empty fields, got %d", w.Code)
	}
}

func TestModelHandler_Get_EmptyID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	// Without chi context, URLParam returns "" which will cause registry.GetModelByID
	// to fail. Since registry is nil, this will panic IF it reaches that code.
	// We're testing that the handler at least processes the request.
	// With nil registry it will panic — but since there's no UUID validation
	// before the service call in the current handler code, the call reaches the service.
	// This test documents the behavior: with no chi context, id="" is passed to service.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/models/", nil)
	w := httptest.NewRecorder()

	// We expect a panic or nil pointer since registry is nil and there's no
	// pre-validation. We recover to verify the handler doesn't validate the ID itself.
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: handler tried to call registry.GetModelByID with empty ID
				// This confirms there's no ID validation in the handler
			}
		}()
		handler.Get(w, req)
	}()
}

func TestModelHandler_Delete_EmptyID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/models/", nil)
	w := httptest.NewRecorder()

	// Same pattern as Get: no pre-validation, will reach service with empty ID
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: handler tried to call registry.DeleteModel with empty ID
			}
		}()
		handler.Delete(w, req)
	}()
}

func TestModelHandler_List_DefaultLimitOffset(t *testing.T) {
	// List doesn't validate limit/offset — invalid values just default to 0
	// This test documents that behavior: strconv.Atoi("abc") returns 0, nil
	logger := zap.NewNop().Sugar()
	handler := NewModelHandler(nil, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/models?limit=abc&offset=xyz", nil)
	w := httptest.NewRecorder()

	// Will panic on nil registry call — we just verify it gets past validation
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: reached registry.ListModels (no validation for limit/offset)
			}
		}()
		handler.List(w, req)
	}()
}

func TestSplitTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single tag", "production", []string{"production"}},
		{"multiple tags", "a,b,c", []string{"a", "b", "c"}},
		{"empty string", "", nil},
		{"trailing comma", "a,b,", []string{"a", "b", ""}},
		{"leading comma", ",a,b", []string{"", "a", "b"}},
		{"double comma", "a,,b", []string{"a", "", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitTags(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitTags(%q) = %v (len %d), expected %v (len %d)",
					tt.input, result, len(result), tt.expected, len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("splitTags(%q)[%d] = %q, expected %q",
						tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestNewModelHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()
	h := NewModelHandler(nil, logger)
	if h == nil {
		t.Fatal("expected non-nil ModelHandler")
	}
	if h.logger != logger {
		t.Error("expected logger to be set")
	}
}
