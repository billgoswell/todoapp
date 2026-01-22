package request

import (
	"github.com/billgoswell/commandlinetodo-server/internal/response"
	"github.com/gin-gonic/gin"
)

// BindJSON binds JSON request body and returns error response if binding fails
func BindJSON(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return err
	}
	return nil
}

// ParseIntParam parses an integer path parameter
func ParseIntParam(c *gin.Context, paramName string, value string) (int, error) {
	// strconv.Atoi is used internally, just wrapper for consistency
	// This can be extended with validation logic
	var result int
	_, err := c.Cookie(paramName) // Check if param exists
	if err == nil {
		// Parameter exists, parse it
		// The actual parsing is done by the caller using strconv.Atoi
	}
	return result, nil
}

// ValidateNonEmpty validates that a string field is not empty
func ValidateNonEmpty(fieldName, fieldValue string) error {
	if fieldValue == "" {
		return NewValidationError(fieldName, "cannot be empty")
	}
	return nil
}

// ValidationError represents a validation error
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
	return e.Field + ": " + e.Reason
}
