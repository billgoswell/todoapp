package validation

import "fmt"

// ValidationError represents a validation error with field and reason information
type ValidationError struct {
	Field  string
	Reason string
}

// NewValidationError creates a new validation error
func NewValidationError(field, reason string) *ValidationError {
	return &ValidationError{
		Field:  field,
		Reason: reason,
	}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Reason)
}

// Message returns a user-friendly error message
func (e *ValidationError) Message() string {
	return e.Error()
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}
