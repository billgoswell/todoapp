package handler

import (
	"fmt"
	"testing"

	"github.com/billgoswell/commandlinetodo/internal/errors"
)

func TestSafeDataStoreOperation_Success(t *testing.T) {
	result, msg := SafeDataStoreOperation("test", func() (int, error) {
		return 42, nil
	})

	if result != 42 {
		t.Errorf("expected result 42, got %d", result)
	}
	if msg != nil {
		t.Errorf("expected no message, got %v", msg)
	}
}

func TestSafeDataStoreOperation_AppError(t *testing.T) {
	testErr := errors.DataNotFoundError("test", "Task", 1)
	result, msg := SafeDataStoreOperation("test", func() (int, error) {
		return 0, testErr
	})

	if result != 0 {
		t.Errorf("expected zero result")
	}
	if msg == nil {
		t.Error("expected error message")
	}
	if msg.Level != MessageError {
		t.Errorf("expected error level")
	}
}

func TestSafeDataStoreOperation_GenericError(t *testing.T) {
	testErr := fmt.Errorf("generic error")
	_, msg := SafeDataStoreOperation("test", func() (int, error) {
		return 0, testErr
	})

	if msg == nil {
		t.Error("expected error message")
	}
	if msg.Content != "generic error" {
		t.Errorf("expected generic error content")
	}
}

func TestSafeOperation_Success(t *testing.T) {
	msg := SafeOperation("test", func() error {
		return nil
	})

	if msg != nil {
		t.Errorf("expected no message, got %v", msg)
	}
}

func TestSafeOperation_AppError(t *testing.T) {
	testErr := errors.ValidationError("test", "name", "too short")
	msg := SafeOperation("test", func() error {
		return testErr
	})

	if msg == nil {
		t.Error("expected error message")
	}
	if msg.Level != MessageError {
		t.Errorf("expected error level")
	}
}

func TestHandleTaskUpdate_Success(t *testing.T) {
	msg := HandleTaskUpdate("update", func() error {
		return nil
	})

	if msg == nil {
		t.Error("expected success message")
	}
	if msg.Level != MessageSuccess {
		t.Errorf("expected success level, got %v", msg.Level)
	}
}

func TestHandleTaskUpdate_Error(t *testing.T) {
	testErr := errors.DatabaseError("update", fmt.Errorf("db error"))
	msg := HandleTaskUpdate("update", func() error {
		return testErr
	})

	if msg == nil {
		t.Error("expected error message")
	}
	if msg.Level != MessageError {
		t.Errorf("expected error level, got %v", msg.Level)
	}
}

func TestHandleTaskDelete_Success(t *testing.T) {
	msg := HandleTaskDelete("delete", func() error {
		return nil
	})

	if msg == nil {
		t.Error("expected success message")
	}
	if msg.Level != MessageSuccess {
		t.Errorf("expected success level, got %v", msg.Level)
	}
}

func TestHandleListCreate_Success(t *testing.T) {
	id, msg := HandleListCreate("create", func() (int, error) {
		return 42, nil
	})

	if id != 42 {
		t.Errorf("expected id 42, got %d", id)
	}
	if msg == nil {
		t.Error("expected success message")
	}
	if msg.Level != MessageSuccess {
		t.Errorf("expected success level, got %v", msg.Level)
	}
}

func TestHandleListCreate_Error(t *testing.T) {
	testErr := errors.DatabaseError("create", fmt.Errorf("db error"))
	id, msg := HandleListCreate("create", func() (int, error) {
		return 0, testErr
	})

	if id != 0 {
		t.Errorf("expected zero id on error")
	}
	if msg == nil {
		t.Error("expected error message")
	}
	if msg.Level != MessageError {
		t.Errorf("expected error level, got %v", msg.Level)
	}
}

func TestHandleListUpdate_Success(t *testing.T) {
	msg := HandleListUpdate("update", func() error {
		return nil
	})

	if msg == nil {
		t.Error("expected success message")
	}
	if msg.Level != MessageSuccess {
		t.Errorf("expected success level, got %v", msg.Level)
	}
}

func TestHandleListDelete_Success(t *testing.T) {
	msg := HandleListDelete("delete", func() error {
		return nil
	})

	if msg == nil {
		t.Error("expected success message")
	}
	if msg.Level != MessageSuccess {
		t.Errorf("expected success level, got %v", msg.Level)
	}
}

func TestIsRecoverableError(t *testing.T) {
	testCases := []struct {
		err        *errors.AppError
		recoverable bool
	}{
		{
			err:         errors.DataNotFoundError("op", "Item", 1),
			recoverable: true,
		},
		{
			err:         errors.ValidationError("op", "field", "reason"),
			recoverable: true,
		},
		{
			err:         errors.EmptyValueError("op", "field"),
			recoverable: true,
		},
		{
			err:         errors.LimitExceededError("op", "field", 100),
			recoverable: true,
		},
		{
			err:         errors.OfflineError("op"),
			recoverable: true,
		},
		{
			err:         errors.DatabaseError("op", fmt.Errorf("error")),
			recoverable: false,
		},
	}

	for i, tc := range testCases {
		if IsRecoverableError(tc.err) != tc.recoverable {
			t.Errorf("test case %d: expected recoverable=%v", i, tc.recoverable)
		}
	}
}

func TestIsFatalError(t *testing.T) {
	testCases := []struct {
		err   *errors.AppError
		fatal bool
	}{
		{
			err:   errors.NewAppError("op", errors.ErrDatabaseSchema, "schema error", nil),
			fatal: true,
		},
		{
			err:   errors.TransactionError("op", fmt.Errorf("tx error")),
			fatal: true,
		},
		{
			err:   errors.DataNotFoundError("op", "Item", 1),
			fatal: false,
		},
		{
			err:   errors.ValidationError("op", "field", "reason"),
			fatal: false,
		},
	}

	for i, tc := range testCases {
		if IsFatalError(tc.err) != tc.fatal {
			t.Errorf("test case %d: expected fatal=%v", i, tc.fatal)
		}
	}
}
