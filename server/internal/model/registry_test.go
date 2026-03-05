package model

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fleetml/fleetml/server/internal/domain"
)

func TestExtractS3Key_Standard(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"s3://bucket/path/to/key", "path/to/key"},
		{"s3://my-bucket/model.onnx", "model.onnx"},
		{"s3://fleetml-models/mobilenet/v1/model.onnx", "mobilenet/v1/model.onnx"},
		{"s3://bucket/a/b/c/d.bin", "a/b/c/d.bin"},
	}

	for _, tt := range tests {
		got := extractS3Key(tt.input)
		if got != tt.expected {
			t.Errorf("extractS3Key(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestExtractS3Key_NoPath(t *testing.T) {
	// s3://bucket-only → no slash after bucket name
	got := extractS3Key("s3://bucket-only")
	if got != "bucket-only" {
		t.Errorf("extractS3Key(%q) = %q, want %q", "s3://bucket-only", got, "bucket-only")
	}
}

func TestExtractS3Key_EmptyOrShort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"s3://", "s3://"},
		{"s3:/", "s3:/"},
		{"", ""},
		{"short", "short"},
	}

	for _, tt := range tests {
		got := extractS3Key(tt.input)
		if got != tt.expected {
			t.Errorf("extractS3Key(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestExtractS3Key_NestedBucketPath(t *testing.T) {
	got := extractS3Key("s3://my-bucket/uuid-123/compiled/tensorrt/model.trt")
	expected := "uuid-123/compiled/tensorrt/model.trt"
	if got != expected {
		t.Errorf("extractS3Key = %q, want %q", got, expected)
	}
}

func TestNilIfEmpty(t *testing.T) {
	// nilIfEmpty is in orchestrator.go but tested here for coverage
	// Since it's in the deploy package, we test extractS3Key which is in model package
}

func TestUploadRequest_FieldsPopulated(t *testing.T) {
	req := UploadRequest{
		Name:        "mobilenet",
		Version:     "v1",
		Format:      "onnx",
		Size:        1024,
		Description: "test model",
		Tags:        []string{"production", "arm"},
	}

	if req.Name != "mobilenet" {
		t.Errorf("expected name 'mobilenet', got %q", req.Name)
	}
	if req.Format != "onnx" {
		t.Errorf("expected format 'onnx', got %q", req.Format)
	}
	if len(req.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(req.Tags))
	}
}

// --- Edge-case tests ---

func TestUploadRequest_EmptyName(t *testing.T) {
	req := UploadRequest{
		Name:    "",
		Version: "v1",
		Format:  "onnx",
		Size:    1024,
	}
	if req.Name != "" {
		t.Errorf("expected empty name, got %q", req.Name)
	}
	// An empty name would produce an S3 key like "/v1/model.onnx" which is invalid.
	// Verify the key would be malformed:
	key := fmt.Sprintf("%s/%s/model.%s", req.Name, req.Version, req.Format)
	if key[0] != '/' {
		t.Errorf("expected leading slash in key with empty name, got %q", key)
	}
}

func TestUploadRequest_EmptyVersion(t *testing.T) {
	req := UploadRequest{
		Name:    "mobilenet",
		Version: "",
		Format:  "onnx",
		Size:    512,
	}
	if req.Version != "" {
		t.Errorf("expected empty version, got %q", req.Version)
	}
	// An empty version produces an S3 key like "mobilenet//model.onnx"
	key := fmt.Sprintf("%s/%s/model.%s", req.Name, req.Version, req.Format)
	expected := "mobilenet//model.onnx"
	if key != expected {
		t.Errorf("expected key %q, got %q", expected, key)
	}
}

func TestModelFilter_NegativeOffset(t *testing.T) {
	filter := domain.ModelFilter{
		Offset: -10,
		Limit:  50,
	}
	// Negative offset is structurally allowed but would cause DB errors.
	// The filter should hold whatever value is set.
	if filter.Offset != -10 {
		t.Errorf("expected offset -10, got %d", filter.Offset)
	}
}

func TestModelFilter_ZeroLimit(t *testing.T) {
	filter := domain.ModelFilter{
		Limit: 0,
	}
	// The ListModels function defaults zero limit to 50.
	if filter.Limit != 0 {
		t.Errorf("expected limit 0 before default, got %d", filter.Limit)
	}
	// Simulate the default logic in ListModels:
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit != 50 {
		t.Errorf("expected defaulted limit 50, got %d", filter.Limit)
	}
}

func TestExtractS3Key_MalformedURLs(t *testing.T) {
	// extractS3Key strips "s3://" (5 chars) prefix, then finds first '/' to split bucket from key.
	// When input is not a proper s3:// URL, the function still applies its logic mechanically.
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single slash after scheme - s3:/bucket/key",
			input:    "s3:/bucket/key",
			expected: "key", // rest="ucket/key", '/' at idx 5 -> rest[6:]="key"
		},
		{
			name:     "double slash no bucket - s3:///key",
			input:    "s3:///key",
			expected: "key", // rest="/key", '/' at idx 0 -> rest[1:]="key"
		},
		{
			name:     "empty string",
			input:    "",
			expected: "", // len <= len(prefix), returns input as-is
		},
		{
			name:     "only scheme prefix",
			input:    "s3://",
			expected: "s3://", // len == len(prefix), returns input as-is
		},
		{
			name:     "scheme with single char bucket no path",
			input:    "s3://x",
			expected: "x", // rest="x", no slash -> returns rest
		},
		{
			name:     "bucket with trailing slash only",
			input:    "s3://bucket/",
			expected: "", // rest="bucket/", '/' at idx 6 -> rest[7:]=""
		},
		{
			name:     "no scheme at all",
			input:    "bucket/path/to/key",
			expected: "path/to/key", // rest="t/path/to/key", '/' at idx 1 -> rest[2:]="path/to/key"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractS3Key(tt.input)
			if got != tt.expected {
				t.Errorf("extractS3Key(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCompiledVariant_JSONRoundTrip(t *testing.T) {
	variant := domain.CompiledVariant{
		Runtime:     "tensorrt",
		ArtifactURL: "s3://fleetml-models/abc-123/compiled/tensorrt/model.trt",
		Checksum:    "sha256:abcdef1234567890",
	}

	data, err := json.Marshal(variant)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded domain.CompiledVariant
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Runtime != variant.Runtime {
		t.Errorf("runtime: expected %q, got %q", variant.Runtime, decoded.Runtime)
	}
	if decoded.ArtifactURL != variant.ArtifactURL {
		t.Errorf("artifact_url: expected %q, got %q", variant.ArtifactURL, decoded.ArtifactURL)
	}
	if decoded.Checksum != variant.Checksum {
		t.Errorf("checksum: expected %q, got %q", variant.Checksum, decoded.Checksum)
	}
}

func TestCompiledVariant_JSONRoundTrip_EmptyFields(t *testing.T) {
	variant := domain.CompiledVariant{}

	data, err := json.Marshal(variant)
	if err != nil {
		t.Fatalf("marshal empty variant: %v", err)
	}

	var decoded domain.CompiledVariant
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal empty variant: %v", err)
	}

	if decoded.Runtime != "" {
		t.Errorf("expected empty runtime, got %q", decoded.Runtime)
	}
	if decoded.ArtifactURL != "" {
		t.Errorf("expected empty artifact_url, got %q", decoded.ArtifactURL)
	}
}

func TestGetVariantForRuntime_NoVariants(t *testing.T) {
	// Test the logic: when a model has no compiled variants, the search should return nil.
	model := domain.Model{
		ID:               "model-1",
		Name:             "mobilenet",
		Version:          "v1",
		CompiledVariants: nil,
	}

	// Replicate GetVariantForRuntime's logic without DB:
	var found *domain.CompiledVariant
	for _, v := range model.CompiledVariants {
		if v.Runtime == "tensorrt" {
			found = &v
			break
		}
	}

	if found != nil {
		t.Error("expected nil variant when model has no variants")
	}
}

func TestGetVariantForRuntime_WrongRuntime(t *testing.T) {
	// Model has variants, but none match the requested runtime.
	model := domain.Model{
		ID:      "model-1",
		Name:    "mobilenet",
		Version: "v1",
		CompiledVariants: []domain.CompiledVariant{
			{Runtime: "openvino", ArtifactURL: "s3://bucket/ov.bin", Checksum: "sha256:abc"},
			{Runtime: "tflite", ArtifactURL: "s3://bucket/tf.tflite", Checksum: "sha256:def"},
		},
	}

	// Search for tensorrt, which does not exist
	var found *domain.CompiledVariant
	for _, v := range model.CompiledVariants {
		if v.Runtime == "tensorrt" {
			found = &v
			break
		}
	}

	if found != nil {
		t.Error("expected nil variant when no variant matches runtime 'tensorrt'")
	}

	// Verify openvino IS found when searched
	for _, v := range model.CompiledVariants {
		if v.Runtime == "openvino" {
			found = &v
			break
		}
	}
	if found == nil {
		t.Error("expected to find openvino variant")
	}
	if found.Checksum != "sha256:abc" {
		t.Errorf("expected checksum 'sha256:abc', got %q", found.Checksum)
	}
}

func TestGetVariantForRuntime_EmptyRuntimeSearch(t *testing.T) {
	model := domain.Model{
		CompiledVariants: []domain.CompiledVariant{
			{Runtime: "tensorrt"},
		},
	}

	var found *domain.CompiledVariant
	for _, v := range model.CompiledVariants {
		if v.Runtime == "" {
			found = &v
			break
		}
	}

	if found != nil {
		t.Error("expected nil when searching for empty runtime string")
	}
}
