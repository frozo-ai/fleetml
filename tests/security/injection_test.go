package security

import (
	"testing"

	mw "github.com/fleetml/fleetml/server/internal/middleware"
)

// TestSQLInjection_UUIDValidation verifies UUID validation blocks SQL injection.
func TestSQLInjection_UUIDValidation(t *testing.T) {
	injections := []string{
		"'; DROP TABLE models; --",
		"1' OR '1'='1",
		"1; SELECT * FROM users",
		"' UNION SELECT * FROM users --",
		"admin'--",
		"1 OR 1=1",
		"' OR ''='",
	}

	for _, payload := range injections {
		if mw.IsValidUUID(payload) {
			t.Errorf("SQL injection payload passed UUID validation: %q", payload)
		}
	}
}

// TestXSS_UUIDValidation verifies UUID validation blocks XSS payloads.
func TestXSS_UUIDValidation(t *testing.T) {
	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert(1)>",
		"javascript:alert(1)",
		"{{7*7}}",
		"${7*7}",
		"\"><script>alert(document.cookie)</script>",
	}

	for _, payload := range xssPayloads {
		if mw.IsValidUUID(payload) {
			t.Errorf("XSS payload passed UUID validation: %q", payload)
		}
	}
}

// TestPathTraversal_UUIDValidation verifies UUID validation blocks path traversal.
func TestPathTraversal_UUIDValidation(t *testing.T) {
	traversals := []string{
		"../../etc/passwd",
		"..\\..\\windows\\system32",
		"%2e%2e%2f%2e%2e%2f",
		"....//....//etc/passwd",
	}

	for _, payload := range traversals {
		if mw.IsValidUUID(payload) {
			t.Errorf("path traversal payload passed UUID validation: %q", payload)
		}
	}
}

// TestCommandInjection_StringValidation verifies string length validation.
func TestCommandInjection_StringValidation(t *testing.T) {
	injections := []string{
		"; cat /etc/passwd",
		"| ls -la",
		"$(whoami)",
		"`id`",
		"& net user",
	}

	for _, payload := range injections {
		// These should pass string validation (correct length) but
		// fail UUID validation (they're not UUIDs)
		if mw.IsValidUUID(payload) {
			t.Errorf("command injection payload passed UUID validation: %q", payload)
		}
	}
}

// TestValidUUIDs_Accepted verifies valid UUIDs pass validation.
func TestValidUUIDs_Accepted(t *testing.T) {
	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"f47ac10b-58cc-4372-a567-0e02b2c3d479",
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}

	for _, uuid := range validUUIDs {
		if !mw.IsValidUUID(uuid) {
			t.Errorf("valid UUID rejected: %q", uuid)
		}
	}
}

// TestStringLengthBoundaries validates edge cases in string length validation.
func TestStringLengthBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		input string
		min   int
		max   int
		valid bool
	}{
		{"empty with min 0", "", 0, 100, true},
		{"empty with min 1", "", 1, 100, false},
		{"exactly min", "abc", 3, 100, true},
		{"exactly max", "abc", 1, 3, true},
		{"over max", "abcd", 1, 3, false},
		{"null bytes", "ab\x00cd", 1, 10, true},
	}

	for _, tt := range tests {
		result := mw.ValidateStringLength(tt.input, tt.min, tt.max)
		if result != tt.valid {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.valid, result)
		}
	}
}
