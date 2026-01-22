package errors

import (
	"fmt"
	"testing"
)

func TestNewAppError(t *testing.T) {
	cause := fmt.Errorf("original error")
	err := NewAppError("test operation", ErrDataNotFound, "test message", cause)

	if err.Code != ErrDataNotFound {
		t.Errorf("expected code %s, got %s", ErrDataNotFound, err.Code)
	}
	if err.Message != "test message" {
		t.Errorf("expected message 'test message', got %s", err.Message)
	}
	if err.Operation != "test operation" {
		t.Errorf("expected operation 'test operation', got %s", err.Operation)
	}
	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}
}

func TestAppError_Error(t *testing.T) {
	cause := fmt.Errorf("original")
	err := NewAppError("op", ErrDataNotFound, "message", cause)
	errStr := err.Error()

	if errStr != "op: message (cause: original)" {
		t.Errorf("unexpected error string: %s", errStr)
	}
}

func TestAppError_ErrorWithoutCause(t *testing.T) {
	err := NewAppError("op", ErrDataNotFound, "message", nil)
	errStr := err.Error()

	if errStr != "op: message" {
		t.Errorf("unexpected error string: %s", errStr)
	}
}

func TestDataNotFoundError(t *testing.T) {
	err := DataNotFoundError("find task", "Task", 42)

	if err.Code != ErrDataNotFound {
		t.Errorf("expected code %s, got %s", ErrDataNotFound, err.Code)
	}
	if err.Operation != "find task" {
		t.Errorf("expected operation 'find task', got %s", err.Operation)
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError("validate task", "title", "title is too short")

	if err.Code != ErrValidationFailed {
		t.Errorf("expected code %s, got %s", ErrValidationFailed, err.Code)
	}
	if err.Details["field"] != "title" {
		t.Errorf("expected field detail 'title', got %v", err.Details["field"])
	}
	if err.Details["reason"] != "title is too short" {
		t.Errorf("expected reason detail 'title is too short', got %v", err.Details["reason"])
	}
}

func TestEmptyValueError(t *testing.T) {
	err := EmptyValueError("validate", "task name")

	if err.Code != ErrEmptyValue {
		t.Errorf("expected code %s, got %s", ErrEmptyValue, err.Code)
	}
}

func TestLimitExceededError(t *testing.T) {
	err := LimitExceededError("validate", "task name", 100)

	if err.Code != ErrLimitExceeded {
		t.Errorf("expected code %s, got %s", ErrLimitExceeded, err.Code)
	}
	if err.Details["limit"] != 100 {
		t.Errorf("expected limit detail 100, got %v", err.Details["limit"])
	}
}

func TestConflictError(t *testing.T) {
	details := map[string]any{
		"local_version":  1,
		"server_version": 2,
	}
	err := ConflictError("sync", "Task", details)

	if err.Code != ErrSyncConflict {
		t.Errorf("expected code %s, got %s", ErrSyncConflict, err.Code)
	}
	if err.Details["local_version"] != 1 {
		t.Errorf("expected local_version detail 1, got %v", err.Details["local_version"])
	}
}

func TestIsErrorCode(t *testing.T) {
	err := DataNotFoundError("op", "Item", 1)

	if !IsErrorCode(err, ErrDataNotFound) {
		t.Errorf("IsErrorCode should return true for matching code")
	}
	if IsErrorCode(err, ErrDataExists) {
		t.Errorf("IsErrorCode should return false for non-matching code")
	}
}

func TestIsErrorCode_NonAppError(t *testing.T) {
	err := fmt.Errorf("generic error")

	if IsErrorCode(err, ErrDataNotFound) {
		t.Errorf("IsErrorCode should return false for non-AppError")
	}
}

func TestErrorCode(t *testing.T) {
	err := DataNotFoundError("op", "Item", 1)

	code := ErrorCode(err)
	if code != ErrDataNotFound {
		t.Errorf("expected code %s, got %s", ErrDataNotFound, code)
	}
}

func TestErrorCode_NonAppError(t *testing.T) {
	err := fmt.Errorf("generic error")

	code := ErrorCode(err)
	if code != "" {
		t.Errorf("expected empty string for non-AppError, got %s", code)
	}
}
