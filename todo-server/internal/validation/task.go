package validation

import (
	"fmt"

	"github.com/billgoswell/commandlinetodo-server/internal/models"
)

// ValidateTask validates a task before creation or update
func ValidateTask(task models.Task) error {
	if err := ValidateTaskText(task.Todo); err != nil {
		return err
	}

	if err := ValidateTaskPriority(task.Priority); err != nil {
		return err
	}

	if task.TodoListID <= 0 {
		return NewValidationError("todo_list_id", "must be greater than 0")
	}

	return nil
}

// ValidateTaskText validates that task text is not empty and within length limit
func ValidateTaskText(text string) error {
	if text == "" {
		return NewValidationError("todo", "cannot be empty")
	}

	if len(text) > 500 {
		return NewValidationError("todo", fmt.Sprintf("cannot exceed 500 characters (got %d)", len(text)))
	}

	return nil
}

// ValidateTaskPriority validates that priority is within valid range (1-4)
func ValidateTaskPriority(priority int) error {
	if priority < 1 || priority > 4 {
		return NewValidationError("priority", "must be between 1 and 4")
	}

	return nil
}

// ValidateTaskUpdate validates a task update request
func ValidateTaskUpdate(original models.Task, updated models.Task) error {
	// Can't change the client_id
	if original.ClientID != updated.ClientID {
		return NewValidationError("client_id", "cannot be changed")
	}

	// Validate the updated fields
	return ValidateTask(updated)
}
