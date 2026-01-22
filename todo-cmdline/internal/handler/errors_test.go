package handler

import (
	"fmt"
	"testing"

	"github.com/billgoswell/commandlinetodo/internal/errors"
)

func TestTaskOperationError(t *testing.T) {
	cause := fmt.Errorf("database error")
	err := TaskOperationError("update task", cause)

	if err.Code != errors.ErrDatabaseIO {
		t.Errorf("expected ErrDatabaseIO, got %s", err.Code)
	}
	if err.Details["entity_type"] != "Task" {
		t.Errorf("expected entity_type Task, got %v", err.Details["entity_type"])
	}
	if err.Cause != cause {
		t.Errorf("expected cause to be set")
	}
}

func TestListOperationError(t *testing.T) {
	cause := fmt.Errorf("database error")
	err := ListOperationError("create list", cause)

	if err.Code != errors.ErrDatabaseIO {
		t.Errorf("expected ErrDatabaseIO, got %s", err.Code)
	}
	if err.Details["entity_type"] != "TodoList" {
		t.Errorf("expected entity_type TodoList, got %v", err.Details["entity_type"])
	}
}

func TestSyncOperationError(t *testing.T) {
	cause := fmt.Errorf("network error")
	err := SyncOperationError("full sync", cause)

	if err.Code != errors.ErrSyncFailed {
		t.Errorf("expected ErrSyncFailed, got %s", err.Code)
	}
	if err.Cause != cause {
		t.Errorf("expected cause to be set")
	}
}

func TestUserInputError(t *testing.T) {
	err := UserInputError("validate input", "task name", "cannot be empty")

	if err.Code != errors.ErrValidationFailed {
		t.Errorf("expected ErrValidationFailed, got %s", err.Code)
	}
	if err.Details["field"] != "task name" {
		t.Errorf("expected field detail")
	}
	if err.Details["reason"] != "cannot be empty" {
		t.Errorf("expected reason detail")
	}
}

func TestUserConfirmationError(t *testing.T) {
	err := UserConfirmationError("delete task")

	if err.Code != errors.ErrUserInput {
		t.Errorf("expected ErrUserInput, got %s", err.Code)
	}
	if err.Message != "Operation cancelled by user" {
		t.Errorf("expected correct message")
	}
}

func TestFromAppError_DefaultLevel(t *testing.T) {
	appErr := errors.NewAppError("op", errors.ErrDatabaseIO, "Database error", nil)
	msg := FromAppError(appErr)

	if msg.Level != MessageError {
		t.Errorf("expected MessageError level, got %s", msg.Level)
	}
	if msg.Content != "Database error" {
		t.Errorf("expected message content")
	}
	if msg.ErrorCode != errors.ErrDatabaseIO {
		t.Errorf("expected error code to be set")
	}
}

func TestFromAppError_OfflineWarning(t *testing.T) {
	appErr := errors.OfflineError("sync")
	msg := FromAppError(appErr)

	if msg.Level != MessageWarn {
		t.Errorf("expected MessageWarn level for offline error, got %s", msg.Level)
	}
}

func TestNewSuccessMessage(t *testing.T) {
	msg := NewSuccessMessage("Task created successfully")

	if msg.Level != MessageSuccess {
		t.Errorf("expected MessageSuccess level")
	}
	if msg.Content != "Task created successfully" {
		t.Errorf("expected correct content")
	}
}

func TestNewErrorMessage(t *testing.T) {
	msg := NewErrorMessage("Task update failed")

	if msg.Level != MessageError {
		t.Errorf("expected MessageError level")
	}
	if msg.Content != "Task update failed" {
		t.Errorf("expected correct content")
	}
}

func TestNewWarnMessage(t *testing.T) {
	msg := NewWarnMessage("No network connection")

	if msg.Level != MessageWarn {
		t.Errorf("expected MessageWarn level")
	}
	if msg.Content != "No network connection" {
		t.Errorf("expected correct content")
	}
}

func TestHandlerMessage_String(t *testing.T) {
	msg := NewErrorMessage("test error")

	if msg.Level == "" {
		t.Error("expected message level to be set")
	}
}
