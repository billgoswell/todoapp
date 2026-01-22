package main

import (
	"time"

	"github.com/billgoswell/commandlinetodo/internal/validation"
)

// Re-export validation functions for backward compatibility
func validateListName(name string) (string, error) {
	return validation.ValidateListName(name)
}

func validateTaskText(text string) (string, error) {
	return validation.ValidateTaskText(text)
}

func validatePriority(priority int) int {
	return validation.ValidatePriority(priority)
}

func setToEndOfDay(t time.Time) time.Time {
	return validation.SetToEndOfDay(t)
}

func isDateInReasonableRange(t time.Time) bool {
	return validation.IsDateInReasonableRange(t)
}

func parseDueDate(input string) int64 {
	return validation.ParseDueDate(input)
}
