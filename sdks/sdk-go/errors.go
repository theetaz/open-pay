package openpay

import "fmt"

// APIError represents an error returned by the Open Pay API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("openpay: %s (%s, HTTP %d)", e.Message, e.Code, e.Status)
}

// IsNotFound returns true if the error is a 404 not found.
func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Status == 404
	}
	return false
}

// IsAuthError returns true if the error is an authentication error.
func IsAuthError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Status == 401
	}
	return false
}
