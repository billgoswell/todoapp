package errors

import (
	"fmt"

	"github.com/billgoswell/commandlinetodo/internal/logger"
)

// AppError represents an application error with context
type AppError struct {
	Code      string         // Machine-readable error code
	Message   string         // User-facing message
	Details   map[string]any
	Cause     error // Underlying error
	Operation string
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Operation, e.Message)
}

// Log logs the error using the logger package
func (e *AppError) Log() {
	logger.LogError(e.Operation, "message", e.Message, "code", e.Code, "cause", e.Cause)
}

// NewAppError creates a new application error
func NewAppError(operation, code, message string, cause error) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Details:   make(map[string]interface{}),
		Cause:     cause,
		Operation: operation,
	}
}

// Common error codes
const (
	// Data errors
	ErrDataNotFound     = "DATA_NOT_FOUND"
	ErrDataExists       = "DATA_EXISTS"
	ErrDataInvalid      = "DATA_INVALID"
	ErrDataCorrupted    = "DATA_CORRUPTED"

	// Database errors
	ErrDatabaseIO       = "DATABASE_IO"
	ErrDatabaseQuery    = "DATABASE_QUERY"
	ErrDatabaseTx       = "DATABASE_TX"
	ErrDatabaseSchema   = "DATABASE_SCHEMA"

	// Validation errors
	ErrValidationFailed = "VALIDATION_FAILED"
	ErrEmptyValue       = "EMPTY_VALUE"
	ErrLimitExceeded    = "LIMIT_EXCEEDED"

	// Sync errors
	ErrSyncFailed       = "SYNC_FAILED"
	ErrSyncConflict     = "SYNC_CONFLICT"
	ErrOffline          = "OFFLINE"

	// User input errors
	ErrUserInput        = "USER_INPUT"
)

// Database operation errors
func DataNotFoundError(operation, entityType string, id any) *AppError {
	return NewAppError(
		operation,
		ErrDataNotFound,
		fmt.Sprintf("%s not found: %v", entityType, id),
		nil,
	)
}

func DatabaseError(operation string, cause error) *AppError {
	return NewAppError(
		operation,
		ErrDatabaseIO,
		"Database operation failed",
		cause,
	)
}

func QueryError(operation string, cause error) *AppError {
	return NewAppError(
		operation,
		ErrDatabaseQuery,
		"Database query failed",
		cause,
	)
}

func TransactionError(operation string, cause error) *AppError {
	return NewAppError(
		operation,
		ErrDatabaseTx,
		"Database transaction failed",
		cause,
	)
}

// Validation errors
func ValidationError(operation, field, reason string) *AppError {
	err := NewAppError(
		operation,
		ErrValidationFailed,
		fmt.Sprintf("Validation failed for %s: %s", field, reason),
		nil,
	)
	err.Details["field"] = field
	err.Details["reason"] = reason
	return err
}

func EmptyValueError(operation, fieldName string) *AppError {
	return NewAppError(
		operation,
		ErrEmptyValue,
		fmt.Sprintf("%s cannot be empty", fieldName),
		nil,
	)
}

func LimitExceededError(operation, fieldName string, limit int) *AppError {
	err := NewAppError(
		operation,
		ErrLimitExceeded,
		fmt.Sprintf("%s exceeds maximum length of %d characters", fieldName, limit),
		nil,
	)
	err.Details["limit"] = limit
	return err
}

// Sync operation errors
func SyncError(operation string, cause error) *AppError {
	return NewAppError(
		operation,
		ErrSyncFailed,
		"Synchronization failed",
		cause,
	)
}

func ConflictError(operation, entityType string, conflictDetails map[string]any) *AppError {
	err := NewAppError(
		operation,
		ErrSyncConflict,
		fmt.Sprintf("Sync conflict detected for %s", entityType),
		nil,
	)
	err.Details = conflictDetails
	return err
}

func OfflineError(operation string) *AppError {
	return NewAppError(
		operation,
		ErrOffline,
		"Operation requires network connectivity",
		nil,
	)
}

// IsErrorCode checks if an error matches a specific code
func IsErrorCode(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// ErrorCode returns the error code if it's an AppError, otherwise returns empty string
func ErrorCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ""
}
