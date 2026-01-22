package handler

import (
	"github.com/billgoswell/commandlinetodo/internal/errors"
)

// SafeDataStoreOperation wraps a data store operation with error handling
// It converts AppError to HandlerMessage and returns both the result and the message
func SafeDataStoreOperation[T any](
	operation string,
	fn func() (T, error),
) (T, *HandlerMessage) {
	result, err := fn()
	if err == nil {
		return result, nil
	}

	// If it's an AppError, convert it to a message
	if appErr, ok := err.(*errors.AppError); ok {
		appErr.Log()
		msg := FromAppError(appErr)
		return result, &msg
	}

	// For non-AppErrors, create a generic error message
	msg := NewErrorMessage(err.Error())
	return result, &msg
}

// SafeOperation wraps any operation with error handling
func SafeOperation(
	operation string,
	fn func() error,
) *HandlerMessage {
	err := fn()
	if err == nil {
		return nil
	}

	// If it's an AppError, convert it to a message
	if appErr, ok := err.(*errors.AppError); ok {
		appErr.Log()
		msg := FromAppError(appErr)
		return &msg
	}

	// For non-AppErrors, create a generic error message
	msg := NewErrorMessage(err.Error())
	return &msg
}

// HandleTaskUpdate performs a task update with consistent error handling
func HandleTaskUpdate(
	operation string,
	updateFn func() error,
) *HandlerMessage {
	if errMsg := SafeOperation(operation, updateFn); errMsg != nil {
		return errMsg
	}
	msg := NewSuccessMessage("Task updated successfully")
	return &msg
}

// HandleTaskDelete performs a task deletion with consistent error handling
func HandleTaskDelete(
	operation string,
	deleteFn func() error,
) *HandlerMessage {
	if errMsg := SafeOperation(operation, deleteFn); errMsg != nil {
		return errMsg
	}
	msg := NewSuccessMessage("Task deleted successfully")
	return &msg
}

// HandleListCreate performs a list creation with consistent error handling
func HandleListCreate(
	operation string,
	createFn func() (int, error),
) (int, *HandlerMessage) {
	id, err := createFn()
	if err == nil {
		msg := NewSuccessMessage("List created successfully")
		return id, &msg
	}

	if appErr, ok := err.(*errors.AppError); ok {
		appErr.Log()
		msg := FromAppError(appErr)
		return 0, &msg
	}

	msg := NewErrorMessage(err.Error())
	return 0, &msg
}

// HandleListUpdate performs a list update with consistent error handling
func HandleListUpdate(
	operation string,
	updateFn func() error,
) *HandlerMessage {
	if errMsg := SafeOperation(operation, updateFn); errMsg != nil {
		return errMsg
	}
	msg := NewSuccessMessage("List updated successfully")
	return &msg
}

// HandleListDelete performs a list deletion with consistent error handling
func HandleListDelete(
	operation string,
	deleteFn func() error,
) *HandlerMessage {
	if errMsg := SafeOperation(operation, deleteFn); errMsg != nil {
		return errMsg
	}
	msg := NewSuccessMessage("List deleted successfully")
	return &msg
}

// IsRecoverableError checks if an error can be recovered from in the UI
func IsRecoverableError(err *errors.AppError) bool {
	switch err.Code {
	case errors.ErrDataNotFound:
		return true
	case errors.ErrValidationFailed:
		return true
	case errors.ErrEmptyValue:
		return true
	case errors.ErrLimitExceeded:
		return true
	case errors.ErrOffline:
		return true
	default:
		return false
	}
}

// IsFatalError checks if an error should terminate the application
func IsFatalError(err *errors.AppError) bool {
	switch err.Code {
	case errors.ErrDatabaseSchema:
		return true
	case errors.ErrDatabaseTx:
		return true
	default:
		return false
	}
}
