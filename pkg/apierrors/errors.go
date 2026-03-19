package apierrors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// FieldError represents a single field validation error.
type FieldError struct {
	Field      string `json:"field"`
	Constraint string `json:"constraint"`
	Message    string `json:"message"`
}

// ValidationError represents a 400 Bad Request with field-level details.
type ValidationError struct {
	Details []FieldError
}

func (e *ValidationError) Error() string {
	if len(e.Details) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", e.Details[0].Message)
}

// NewValidation creates a ValidationError with a single field error.
func NewValidation(field, constraint, message string) *ValidationError {
	return &ValidationError{
		Details: []FieldError{{Field: field, Constraint: constraint, Message: message}},
	}
}

// ValidationBuilder helps construct multi-field validation errors.
type ValidationBuilder struct {
	errors []FieldError
}

// NewValidationBuilder creates a new builder.
func NewValidationBuilder() *ValidationBuilder {
	return &ValidationBuilder{}
}

// Add appends a field error.
func (b *ValidationBuilder) Add(field, constraint, message string) *ValidationBuilder {
	b.errors = append(b.errors, FieldError{Field: field, Constraint: constraint, Message: message})
	return b
}

// HasErrors returns true if any errors were added.
func (b *ValidationBuilder) HasErrors() bool {
	return len(b.errors) > 0
}

// Build returns the ValidationError or nil if no errors.
func (b *ValidationBuilder) Build() *ValidationError {
	if len(b.errors) == 0 {
		return nil
	}
	return &ValidationError{Details: b.errors}
}

// BusinessError represents a 422 Unprocessable Entity — business logic violation.
type BusinessError struct {
	Code    string
	Message string
}

func (e *BusinessError) Error() string {
	return e.Message
}

// NewBusiness creates a BusinessError.
func NewBusiness(code, message string) *BusinessError {
	return &BusinessError{Code: code, Message: message}
}

// GoneError represents a 410 Gone — resource no longer available.
type GoneError struct {
	Code    string
	Message string
}

func (e *GoneError) Error() string {
	return e.Message
}

// NewGone creates a GoneError.
func NewGone(code, message string) *GoneError {
	return &GoneError{Code: code, Message: message}
}

// ConflictError represents a 409 Conflict.
type ConflictError struct {
	Code    string
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

// NewConflict creates a ConflictError.
func NewConflict(code, message string) *ConflictError {
	return &ConflictError{Code: code, Message: message}
}

// NotFoundError represents a 404 Not Found.
type NotFoundError struct {
	Code    string
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// NewNotFound creates a NotFoundError.
func NewNotFound(code, message string) *NotFoundError {
	return &NotFoundError{Code: code, Message: message}
}

// WriteError inspects the error type and writes the appropriate HTTP response.
func WriteError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *ValidationError:
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"statusCode": http.StatusBadRequest,
			"message":    "Validation failed",
			"error":      "Bad Request",
			"details":    e.Details,
		})
	case *BusinessError:
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"statusCode": http.StatusUnprocessableEntity,
			"message":    e.Message,
			"error":      "Unprocessable Entity",
			"code":       e.Code,
		})
	case *GoneError:
		writeJSON(w, http.StatusGone, map[string]any{
			"statusCode": http.StatusGone,
			"message":    e.Message,
			"error":      "Gone",
			"code":       e.Code,
		})
	case *ConflictError:
		writeJSON(w, http.StatusConflict, map[string]any{
			"statusCode": http.StatusConflict,
			"message":    e.Message,
			"error":      "Conflict",
			"code":       e.Code,
		})
	case *NotFoundError:
		writeJSON(w, http.StatusNotFound, map[string]any{
			"statusCode": http.StatusNotFound,
			"message":    e.Message,
			"error":      "Not Found",
			"code":       e.Code,
		})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"statusCode": http.StatusInternalServerError,
			"message":    "An unexpected error occurred",
			"error":      "Internal Server Error",
		})
	}
}

// WriteSimpleError writes a simple error response with the given status code.
func WriteSimpleError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"statusCode": status,
		"message":    message,
		"error":      http.StatusText(status),
		"code":       code,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
