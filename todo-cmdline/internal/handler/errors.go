package handler

import (
	"github.com/billgoswell/commandlinetodo/internal/errors"
)

// Handler-specific error creation helpers

// TaskOperationError creates an error for task operations
func TaskOperationError(operation string, err error) *errors.AppError {
	appErr := errors.DatabaseError(operation, err)
	appErr.Details["entity_type"] = "Task"
	return appErr
}

// ListOperationError creates an error for list operations
func ListOperationError(operation string, err error) *errors.AppError {
	appErr := errors.DatabaseError(operation, err)
	appErr.Details["entity_type"] = "TodoList"
	return appErr
}

// SyncOperationError creates an error for sync operations
func SyncOperationError(operation string, err error) *errors.AppError {
	return errors.SyncError(operation, err)
}

// UserInputError creates an error for invalid user input
func UserInputError(operation, field, reason string) *errors.AppError {
	return errors.ValidationError(operation, field, reason)
}

// UserConfirmationError creates an error when user cancels an operation
func UserConfirmationError(operation string) *errors.AppError {
	return errors.NewAppError(
		operation,
		errors.ErrUserInput,
		"Operation cancelled by user",
		nil,
	)
}

// HandlerMessage represents a user-facing message from a handler
type HandlerMessage struct {
	Level   MessageLevel
	Content string
	// Optional error code for structured handling
	ErrorCode string
}

// MessageLevel defines the severity of a handler message
type MessageLevel string

const (
	MessageSuccess MessageLevel = "success"
	MessageInfo    MessageLevel = "info"
	MessageWarn    MessageLevel = "warn"
	MessageError   MessageLevel = "error"
)

// FromAppError converts an AppError to a user-facing message
func FromAppError(err *errors.AppError) HandlerMessage {
	level := MessageError

	// Map error codes to message levels
	switch err.Code {
	case errors.ErrOffline:
		level = MessageWarn
	}

	return HandlerMessage{
		Level:     level,
		Content:   err.Message,
		ErrorCode: err.Code,
	}
}

// NewSuccessMessage creates a success message
func NewSuccessMessage(content string) HandlerMessage {
	return HandlerMessage{
		Level:   MessageSuccess,
		Content: content,
	}
}

// NewErrorMessage creates an error message
func NewErrorMessage(content string) HandlerMessage {
	return HandlerMessage{
		Level:   MessageError,
		Content: content,
	}
}

// NewWarnMessage creates a warning message
func NewWarnMessage(content string) HandlerMessage {
	return HandlerMessage{
		Level:   MessageWarn,
		Content: content,
	}
}
