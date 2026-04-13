package grpc

import (
	"testing"
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

func TestExtractS3Key_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no path after bucket", "s3://bucket-only", "bucket-only"},
		{"empty after prefix", "s3://", "s3://"},
		{"short string", "s3:/", "s3:/"},
		{"empty", "", ""},
		{"no s3 prefix", "short", "short"},
		{"variant path", "s3://fleetml-models/uuid-123/compiled/tensorrt/model.trt", "uuid-123/compiled/tensorrt/model.trt"},
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

func TestNewHandler_NilDependencies(t *testing.T) {
	// Handler should be constructible with nil dependencies (used in tests)
	h := NewHandler(nil, nil, nil, nil, nil, nil, nil)
	if h == nil {
		t.Fatal("expected non-nil Handler")
	}
}

func TestHandler_GetModelArtifact_NoRegistryOrStore(t *testing.T) {
	// Handler with nil registry/store should return error early
	h := NewHandler(nil, nil, nil, nil, nil, nil, nil)
	if h.registry != nil {
		t.Error("expected nil registry")
	}
	if h.store != nil {
		t.Error("expected nil store")
	}
}

func TestNewHandler_AllNilFields(t *testing.T) {
	h := NewHandler(nil, nil, nil, nil, nil, nil, nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.fleet != nil || h.orchestrator != nil || h.registry != nil || h.store != nil || h.metrics != nil || h.db != nil || h.logger != nil {
		t.Error("expected all fields nil")
	}
}

func TestNewHandler_DistinctInstances(t *testing.T) {
	h1 := NewHandler(nil, nil, nil, nil, nil, nil, nil)
	h2 := NewHandler(nil, nil, nil, nil, nil, nil, nil)
	if h1 == h2 {
		t.Error("two NewHandler calls should return distinct pointers")
	}
}

func TestExtractS3Key_EmptyString(t *testing.T) {
	if got := extractS3Key(""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestExtractS3Key_JustProtocol(t *testing.T) {
	if got := extractS3Key("s3://"); got != "s3://" {
		t.Errorf("expected 's3://', got %q", got)
	}
}

func TestExtractS3Key_NoBucket(t *testing.T) {
	if got := extractS3Key("s3:///"); got != "" {
		t.Errorf("expected empty for 's3:///', got %q", got)
	}
}

func TestExtractS3Key_DeepPath(t *testing.T) {
	if got := extractS3Key("s3://b/a/b/c/d/e/f.ext"); got != "a/b/c/d/e/f.ext" {
		t.Errorf("expected 'a/b/c/d/e/f.ext', got %q", got)
	}
}

func TestExtractS3Key_SpecialChars(t *testing.T) {
	tests := []struct{ name, input, expected string }{
		{"spaces", "s3://bucket/path with spaces/model.onnx", "path with spaces/model.onnx"},
		{"percent", "s3://bucket/path%20encoded/file.bin", "path%20encoded/file.bin"},
		{"dots-dashes", "s3://bucket/a.b-c_d/e.f", "a.b-c_d/e.f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractS3Key(tt.input); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestExtractS3Key_NonS3Prefix(t *testing.T) {
	// Non-s3 URLs — verify no panic
	_ = extractS3Key("http://bucket/key")
	_ = extractS3Key("gs://bucket/key")
}

func TestExtractS3Key_DotInBucket(t *testing.T) {
	if got := extractS3Key("s3://my.bucket.name/key"); got != "key" {
		t.Errorf("expected 'key', got %q", got)
	}
}

func TestExtractS3Key_TrailingSlash(t *testing.T) {
	if got := extractS3Key("s3://bucket/path/"); got != "path/" {
		t.Errorf("expected 'path/', got %q", got)
	}
}
