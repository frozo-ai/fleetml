package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout redirects os.Stdout to a pipe, calls fn, and returns whatever
// was written to stdout as a string.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	return buf.String()
}

// captureStderr redirects os.Stderr to a pipe, calls fn, and returns whatever
// was written to stderr as a string.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	return buf.String()
}

// ---------------------------------------------------------------------------
// PrintTable
// ---------------------------------------------------------------------------

func TestPrintTable_BasicTable(t *testing.T) {
	headers := []string{"NAME", "STATUS", "VERSION"}
	rows := [][]string{
		{"model-a", "active", "1.0.0"},
		{"model-b", "pending", "2.1.0"},
	}

	output := captureStdout(t, func() {
		PrintTable(headers, rows)
	})

	// Verify headers appear.
	if !strings.Contains(output, "NAME") {
		t.Error("expected output to contain 'NAME'")
	}
	if !strings.Contains(output, "STATUS") {
		t.Error("expected output to contain 'STATUS'")
	}
	if !strings.Contains(output, "VERSION") {
		t.Error("expected output to contain 'VERSION'")
	}

	// Verify row data appears.
	if !strings.Contains(output, "model-a") {
		t.Error("expected output to contain 'model-a'")
	}
	if !strings.Contains(output, "model-b") {
		t.Error("expected output to contain 'model-b'")
	}
	if !strings.Contains(output, "active") {
		t.Error("expected output to contain 'active'")
	}
	if !strings.Contains(output, "2.1.0") {
		t.Error("expected output to contain '2.1.0'")
	}
}

func TestPrintTable_EmptyRows(t *testing.T) {
	headers := []string{"ID", "NAME"}
	rows := [][]string{}

	output := captureStdout(t, func() {
		PrintTable(headers, rows)
	})

	// Headers should still appear.
	if !strings.Contains(output, "ID") {
		t.Error("expected headers even with no rows")
	}
	// A separator line should exist.
	if !strings.Contains(output, "-") {
		t.Error("expected separator line")
	}
}

func TestPrintTable_SingleColumn(t *testing.T) {
	headers := []string{"DEVICE"}
	rows := [][]string{
		{"jetson-001"},
		{"rpi-002"},
	}

	output := captureStdout(t, func() {
		PrintTable(headers, rows)
	})

	if !strings.Contains(output, "DEVICE") {
		t.Error("expected 'DEVICE' header")
	}
	if !strings.Contains(output, "jetson-001") {
		t.Error("expected 'jetson-001' in output")
	}
	if !strings.Contains(output, "rpi-002") {
		t.Error("expected 'rpi-002' in output")
	}
}

func TestPrintTable_SpecialCharacters(t *testing.T) {
	headers := []string{"NAME", "DESCRIPTION"}
	rows := [][]string{
		{"model/v1", "test & validate <models>"},
		{"model@2", "100% accuracy (hopefully)"},
	}

	output := captureStdout(t, func() {
		PrintTable(headers, rows)
	})

	if !strings.Contains(output, "model/v1") {
		t.Error("expected 'model/v1' in output")
	}
	if !strings.Contains(output, "test & validate <models>") {
		t.Error("expected special characters preserved in output")
	}
	if !strings.Contains(output, "100% accuracy (hopefully)") {
		t.Error("expected special characters preserved in output")
	}
}

func TestPrintTable_MultipleLines(t *testing.T) {
	headers := []string{"A", "B"}
	rows := [][]string{
		{"1", "2"},
	}

	output := captureStdout(t, func() {
		PrintTable(headers, rows)
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	// Expect at least 3 lines: header, separator, data row.
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d: %q", len(lines), output)
	}
}

// ---------------------------------------------------------------------------
// PrintJSON
// ---------------------------------------------------------------------------

func TestPrintJSON_ValidStruct(t *testing.T) {
	data := struct {
		Name    string `json:"name"`
		Version int    `json:"version"`
	}{
		Name:    "test-model",
		Version: 3,
	}

	output := captureStdout(t, func() {
		PrintJSON(data)
	})

	// Verify it is valid JSON.
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("expected valid JSON output, got parse error: %v\noutput: %s", err, output)
	}

	if parsed["name"] != "test-model" {
		t.Errorf("expected name 'test-model', got %v", parsed["name"])
	}
	// JSON numbers are float64.
	if parsed["version"] != float64(3) {
		t.Errorf("expected version 3, got %v", parsed["version"])
	}
}

func TestPrintJSON_NilValue(t *testing.T) {
	output := captureStdout(t, func() {
		PrintJSON(nil)
	})

	trimmed := strings.TrimSpace(output)
	if trimmed != "null" {
		t.Fatalf("expected 'null' for nil value, got %q", trimmed)
	}
}

func TestPrintJSON_NestedObject(t *testing.T) {
	data := map[string]interface{}{
		"model": map[string]interface{}{
			"name":    "resnet50",
			"version": 2,
			"tags":    []string{"production", "optimized"},
		},
		"status": "active",
	}

	output := captureStdout(t, func() {
		PrintJSON(data)
	})

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	model, ok := parsed["model"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'model' to be a nested object")
	}
	if model["name"] != "resnet50" {
		t.Errorf("expected model.name 'resnet50', got %v", model["name"])
	}
}

func TestPrintJSON_Indented(t *testing.T) {
	data := map[string]string{"key": "value"}

	output := captureStdout(t, func() {
		PrintJSON(data)
	})

	// PrintJSON uses SetIndent("", "  "), so output should be multi-line.
	if !strings.Contains(output, "\n") {
		t.Error("expected indented (multi-line) JSON output")
	}
	if !strings.Contains(output, "  ") {
		t.Error("expected 2-space indentation in JSON output")
	}
}

func TestPrintJSON_SliceOutput(t *testing.T) {
	data := []string{"alpha", "beta", "gamma"}

	output := captureStdout(t, func() {
		PrintJSON(data)
	})

	var parsed []string
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("expected valid JSON array, got error: %v", err)
	}
	if len(parsed) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(parsed))
	}
	if parsed[0] != "alpha" {
		t.Errorf("expected first element 'alpha', got %q", parsed[0])
	}
}

// ---------------------------------------------------------------------------
// Success
// ---------------------------------------------------------------------------

func TestSuccess_MessageContainsInput(t *testing.T) {
	output := captureStdout(t, func() {
		Success("Model deployed successfully")
	})

	if !strings.Contains(output, "Model deployed successfully") {
		t.Fatalf("expected output to contain message, got %q", output)
	}
}

func TestSuccess_EmptyMessage(t *testing.T) {
	output := captureStdout(t, func() {
		Success("")
	})

	// Should still produce output (the formatting prefix/suffix).
	if len(output) == 0 {
		t.Fatal("expected non-empty output even for empty message")
	}
}

func TestSuccess_SpecialCharacters(t *testing.T) {
	output := captureStdout(t, func() {
		Success("Deployed model <v2> with 100% success & rollback ready")
	})

	if !strings.Contains(output, "100%") {
		t.Error("expected special characters to be preserved")
	}
	if !strings.Contains(output, "<v2>") {
		t.Error("expected angle brackets to be preserved")
	}
}

// ---------------------------------------------------------------------------
// Error
// ---------------------------------------------------------------------------

func TestError_MessageContainsInput(t *testing.T) {
	output := captureStderr(t, func() {
		Error("deployment failed: timeout")
	})

	if !strings.Contains(output, "deployment failed: timeout") {
		t.Fatalf("expected stderr output to contain message, got %q", output)
	}
}

func TestError_EmptyMessage(t *testing.T) {
	output := captureStderr(t, func() {
		Error("")
	})

	if len(output) == 0 {
		t.Fatal("expected non-empty output even for empty error message")
	}
}

func TestError_WritesToStderr(t *testing.T) {
	// Verify Error writes to stderr, not stdout.
	stdoutOutput := captureStdout(t, func() {
		// We cannot capture stderr here simultaneously without more complex
		// plumbing, so we just verify stdout is empty.
		// Error writes to os.Stderr, which is NOT redirected by captureStdout.

		// We need to also redirect stderr to /dev/null to avoid test noise.
		origStderr := os.Stderr
		devNull, _ := os.Open(os.DevNull)
		os.Stderr = devNull
		Error("this goes to stderr")
		os.Stderr = origStderr
		devNull.Close()
	})

	if strings.Contains(stdoutOutput, "this goes to stderr") {
		t.Error("Error() should not write to stdout")
	}
}

func TestError_SpecialCharacters(t *testing.T) {
	output := captureStderr(t, func() {
		Error("failed: connection refused (port 50051)")
	})

	if !strings.Contains(output, "connection refused (port 50051)") {
		t.Error("expected special characters to be preserved in error output")
	}
}
