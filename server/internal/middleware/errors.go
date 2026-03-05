package middleware

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard error response format for all REST endpoints.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// WriteError writes a structured JSON error response.
func WriteError(w http.ResponseWriter, statusCode int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// WriteBadRequest writes a 400 error response.
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, message, "BAD_REQUEST")
}

// WriteNotFound writes a 404 error response.
func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, message, "NOT_FOUND")
}

// WriteUnauthorized writes a 401 error response.
func WriteUnauthorized(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusUnauthorized, message, "UNAUTHORIZED")
}

// WriteForbidden writes a 403 error response.
func WriteForbidden(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusForbidden, message, "FORBIDDEN")
}

// WriteInternalError writes a 500 error response.
func WriteInternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, message, "INTERNAL_ERROR")
}

// WriteConflict writes a 409 error response.
func WriteConflict(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, message, "CONFLICT")
}
