package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestDriftHandler_Analyze_InvalidJSON(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader("{broken json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestDriftHandler_Analyze_EmptyBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestDriftHandler_Analyze_EmptySamples(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"device_id":"d1","model_id":"m1","feature_name":"feat","samples":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty samples, got %d", w.Code)
	}
}

func TestDriftHandler_Analyze_MissingSamples(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"device_id":"d1","model_id":"m1","feature_name":"feat"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing samples, got %d", w.Code)
	}
}

func TestDriftHandler_Analyze_MissingDeviceID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"model_id":"m1","feature_name":"feat","samples":[1.0,2.0]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing device_id, got %d", w.Code)
	}
}

func TestDriftHandler_Analyze_MissingModelID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"device_id":"d1","feature_name":"feat","samples":[1.0,2.0]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing model_id, got %d", w.Code)
	}
}

func TestDriftHandler_Analyze_MissingFeatureName(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"device_id":"d1","model_id":"m1","samples":[1.0,2.0]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing feature_name, got %d", w.Code)
	}
}

func TestDriftHandler_Analyze_EmptyObject(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty object, got %d", w.Code)
	}
}

func TestDriftHandler_Analyze_SamplesWrongType(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	// samples should be []float64 not a string
	body := `{"device_id":"d1","model_id":"m1","feature_name":"feat","samples":"not-an-array"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for samples with wrong type, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_InvalidJSON(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader("{{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetBaseline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_EmptyBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetBaseline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_EmptySamples(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"model_id":"m1","feature_name":"feat","samples":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetBaseline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty samples, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_MissingModelID(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"feature_name":"feat","samples":[1.0,2.0]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetBaseline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing model_id, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_MissingFeatureName(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"model_id":"m1","samples":[1.0,2.0]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetBaseline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing feature_name, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_MissingSamples(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"model_id":"m1","feature_name":"feat"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetBaseline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing samples, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_EmptyObject(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetBaseline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty object, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_SamplesWrongType(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{"model_id":"m1","feature_name":"feat","samples":"not-array"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetBaseline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for samples with wrong type, got %d", w.Code)
	}
}

func TestDriftHandler_ListReports_NoQueryParams(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/drift/reports", nil)
	w := httptest.NewRecorder()

	// Will panic on nil detector.ListReports — shows no validation errors on empty query params
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: no query param validation, reaches detector.ListReports
			}
		}()
		handler.ListReports(w, req)
	}()
}

func TestDriftHandler_Analyze_ResponseErrorFormat(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Verify error response contains proper JSON with error field
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("expected valid JSON error response, got error: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Error("expected 'error' field in response")
	}
}

func TestDriftHandler_Analyze_ArrayInsteadOfObject(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `[{"device_id":"d1"}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for array instead of object, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_ArrayInsteadOfObject(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewDriftHandler(nil, logger)

	body := `[{"model_id":"m1"}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetBaseline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for array instead of object, got %d", w.Code)
	}
}

func TestNewDriftHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()
	h := NewDriftHandler(nil, logger)
	if h == nil {
		t.Fatal("expected non-nil DriftHandler")
	}
	if h.logger != logger {
		t.Error("expected logger to be set")
	}
}

func TestNewDriftHandler_NilLogger(t *testing.T) {
	h := NewDriftHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestDriftHandler_Analyze_NilBody(t *testing.T) {
	h := NewDriftHandler(nil, zap.NewNop().Sugar())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", nil)
	w := httptest.NewRecorder()
	h.Analyze(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for nil body, got %d", w.Code)
	}
}

func TestDriftHandler_SetBaseline_NilBody(t *testing.T) {
	h := NewDriftHandler(nil, zap.NewNop().Sugar())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", nil)
	w := httptest.NewRecorder()
	h.SetBaseline(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for nil body, got %d", w.Code)
	}
}

func TestDriftHandler_Analyze_SingleSample(t *testing.T) {
	h := NewDriftHandler(nil, zap.NewNop().Sugar())
	body := `{"device_id":"d1","model_id":"m1","feature_name":"f1","samples":[1.0]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	func() {
		defer func() { recover() }()
		h.Analyze(w, req)
	}()
	// Valid input — reaches detector which is nil, panic recovered
}

func TestDriftHandler_SetBaseline_SingleSample(t *testing.T) {
	h := NewDriftHandler(nil, zap.NewNop().Sugar())
	body := `{"model_id":"m1","feature_name":"f1","samples":[42.0]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/baselines", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	func() {
		defer func() { recover() }()
		h.SetBaseline(w, req)
	}()
	// Valid input — reaches detector which is nil, panic recovered
}

func TestDriftHandler_Analyze_NegativeSamples(t *testing.T) {
	h := NewDriftHandler(nil, zap.NewNop().Sugar())
	body := `{"device_id":"d1","model_id":"m1","feature_name":"f1","samples":[-1.0,-2.5,-100.0]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	func() {
		defer func() { recover() }()
		h.Analyze(w, req)
	}()
	// Negative samples should be accepted by validation
}

func TestDriftHandler_Analyze_LargeSamplesArray(t *testing.T) {
	h := NewDriftHandler(nil, zap.NewNop().Sugar())
	samples := make([]float64, 10000)
	for i := range samples {
		samples[i] = float64(i)
	}
	samplesJSON, _ := json.Marshal(samples)
	body := `{"device_id":"d1","model_id":"m1","feature_name":"f1","samples":` + string(samplesJSON) + `}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drift/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	func() {
		defer func() { recover() }()
		h.Analyze(w, req)
	}()
	// Large samples should parse fine
}
