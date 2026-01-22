package validation

import (
	"testing"

	"github.com/billgoswell/commandlinetodo-server/internal/models"
)

// TestValidateTaskText_Valid tests valid task text
func TestValidateTaskText_Valid(t *testing.T) {
	err := ValidateTaskText("Buy groceries")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestValidateTaskText_Empty tests empty task text
func TestValidateTaskText_Empty(t *testing.T) {
	err := ValidateTaskText("")
	if err == nil {
		t.Error("expected error for empty text")
	}
	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %T", err)
	}
}

// TestValidateTaskText_TooLong tests task text exceeding max length
func TestValidateTaskText_TooLong(t *testing.T) {
	longText := ""
	for i := 0; i < 501; i++ {
		longText += "a"
	}
	err := ValidateTaskText(longText)
	if err == nil {
		t.Error("expected error for text exceeding 500 chars")
	}
}

// TestValidateTaskPriority_Valid tests valid priorities
func TestValidateTaskPriority_Valid(t *testing.T) {
	for priority := 1; priority <= 4; priority++ {
		err := ValidateTaskPriority(priority)
		if err != nil {
			t.Errorf("expected no error for priority %d, got %v", priority, err)
		}
	}
}

// TestValidateTaskPriority_TooLow tests priority below valid range
func TestValidateTaskPriority_TooLow(t *testing.T) {
	err := ValidateTaskPriority(0)
	if err == nil {
		t.Error("expected error for priority < 1")
	}
}

// TestValidateTaskPriority_TooHigh tests priority above valid range
func TestValidateTaskPriority_TooHigh(t *testing.T) {
	err := ValidateTaskPriority(5)
	if err == nil {
		t.Error("expected error for priority > 4")
	}
}

// TestValidateTask_Complete tests complete task validation
func TestValidateTask_Complete(t *testing.T) {
	task := models.Task{
		ClientID:   "client-1",
		TodoListID: 1,
		Todo:       "Buy groceries",
		Priority:   2,
		Done:       false,
	}
	err := ValidateTask(task)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestValidateTask_InvalidTodoListID tests task with invalid todo list id
func TestValidateTask_InvalidTodoListID(t *testing.T) {
	task := models.Task{
		ClientID:   "client-1",
		TodoListID: 0,
		Todo:       "Buy groceries",
		Priority:   2,
	}
	err := ValidateTask(task)
	if err == nil {
		t.Error("expected error for invalid todo_list_id")
	}
}

// TestValidateListName_Valid tests valid list name
func TestValidateListName_Valid(t *testing.T) {
	err := ValidateListName("Work")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestValidateListName_Empty tests empty list name
func TestValidateListName_Empty(t *testing.T) {
	err := ValidateListName("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestValidateListName_TooLong tests list name exceeding max length
func TestValidateListName_TooLong(t *testing.T) {
	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}
	err := ValidateListName(longName)
	if err == nil {
		t.Error("expected error for name exceeding 100 chars")
	}
}

// TestValidateDisplayOrder_Valid tests valid display order
func TestValidateDisplayOrder_Valid(t *testing.T) {
	for order := 0; order <= 10; order++ {
		err := ValidateDisplayOrder(order)
		if err != nil {
			t.Errorf("expected no error for order %d, got %v", order, err)
		}
	}
}

// TestValidateDisplayOrder_Negative tests negative display order
func TestValidateDisplayOrder_Negative(t *testing.T) {
	err := ValidateDisplayOrder(-1)
	if err == nil {
		t.Error("expected error for negative display_order")
	}
}

// TestValidateList_Complete tests complete list validation
func TestValidateList_Complete(t *testing.T) {
	list := models.TodoList{
		ClientID:     "list-1",
		Name:         "Work",
		DisplayOrder: 0,
		Archived:     false,
	}
	err := ValidateList(list)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestValidationError_Message tests validation error message formatting
func TestValidationError_Message(t *testing.T) {
	err := NewValidationError("field_name", "invalid value")
	expectedMsg := "field_name: invalid value"
	if err.Error() != expectedMsg {
		t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestIsValidationError tests validation error type checking
func TestIsValidationError(t *testing.T) {
	testCases := []struct {
		err      error
		expected bool
	}{
		{
			err:      NewValidationError("field", "reason"),
			expected: true,
		},
		{
			err:      nil,
			expected: false,
		},
	}

	for _, tc := range testCases {
		if tc.err == nil {
			continue
		}
		result := IsValidationError(tc.err)
		if result != tc.expected {
			t.Errorf("expected %v, got %v", tc.expected, result)
		}
	}
}
