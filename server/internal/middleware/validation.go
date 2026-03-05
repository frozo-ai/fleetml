package middleware

import (
	"net/http"
	"regexp"

	"go.uber.org/zap"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// IsValidUUID checks if a string is a valid UUID v4 format.
func IsValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// MaxBodySize returns middleware that limits request body size.
// Default: 100MB for model uploads, 1MB for API requests.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				WriteError(w, http.StatusRequestEntityTooLarge,
					"request body too large", "BODY_TOO_LARGE")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// RequestSizeLimit returns middleware that limits general API request body to 1MB.
func RequestSizeLimit() func(http.Handler) http.Handler {
	return MaxBodySize(1 * 1024 * 1024) // 1MB
}

// ModelUploadSizeLimit returns middleware that limits model upload body to 500MB.
func ModelUploadSizeLimit() func(http.Handler) http.Handler {
	return MaxBodySize(500 * 1024 * 1024) // 500MB
}

// ValidateStringLength checks if a string is within acceptable bounds.
func ValidateStringLength(s string, minLen, maxLen int) bool {
	return len(s) >= minLen && len(s) <= maxLen
}

// SecurityHeaders adds security headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// RequestLogger logs incoming requests with timing info.
func RequestLogger(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debugw("request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			next.ServeHTTP(w, r)
		})
	}
}
