package validation

import (
	"fmt"

	"github.com/billgoswell/commandlinetodo-server/internal/models"
)

// ValidateList validates a list before creation or update
func ValidateList(list models.TodoList) error {
	if err := ValidateListName(list.Name); err != nil {
		return err
	}

	if err := ValidateDisplayOrder(list.DisplayOrder); err != nil {
		return err
	}

	return nil
}

// ValidateListName validates that list name is not empty and within length limit
func ValidateListName(name string) error {
	if name == "" {
		return NewValidationError("name", "cannot be empty")
	}

	if len(name) > 100 {
		return NewValidationError("name", fmt.Sprintf("cannot exceed 100 characters (got %d)", len(name)))
	}

	return nil
}

// ValidateDisplayOrder validates that display order is non-negative
func ValidateDisplayOrder(order int) error {
	if order < 0 {
		return NewValidationError("display_order", "cannot be negative")
	}

	return nil
}

// ValidateListUpdate validates a list update request
func ValidateListUpdate(original models.TodoList, updated models.TodoList) error {
	// Can't change the client_id
	if original.ClientID != updated.ClientID {
		return NewValidationError("client_id", "cannot be changed")
	}

	// Validate the updated fields
	return ValidateList(updated)
}
