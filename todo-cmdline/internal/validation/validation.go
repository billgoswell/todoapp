package validation

import (
	"fmt"
	"strings"
	"time"

	internalApp "github.com/billgoswell/commandlinetodo/internal/app"
)

// ValidateListName checks if a list name is valid
func ValidateListName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("list name cannot be empty")
	}
	const maxListNameLength = 100
	if len(trimmed) > maxListNameLength {
		return "", fmt.Errorf("list name cannot exceed %d characters", maxListNameLength)
	}
	return trimmed, nil
}

// ValidateTaskText checks if a task text is valid
func ValidateTaskText(text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", fmt.Errorf("task cannot be empty")
	}
	return trimmed, nil
}

// ValidatePriority ensures priority is in valid range
func ValidatePriority(priority int) int {
	if priority < 1 || priority > 4 {
		return internalApp.DefaultPriority
	}
	return priority
}

// SetToEndOfDay sets time to 23:59:59 for that date
func SetToEndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, time.Local)
}

// IsDateInReasonableRange checks if date is within reasonable bounds
func IsDateInReasonableRange(t time.Time) bool {
	now := time.Now()
	// Allow dates up to 1 year in the past (for historical tracking)
	if t.Before(now.AddDate(-1, 0, 0)) {
		return false
	}
	// Don't allow dates too far in the future (100 years)
	if t.After(now.AddDate(0, 0, internalApp.MaxDaysOffset)) {
		return false
	}
	return true
}

// ParseDueDate converts user input to unix timestamp for due date
func ParseDueDate(input string) int64 {
	if input == "" {
		return 0
	}

	var days int
	if len(input) <= 3 {
		_, err := fmt.Sscanf(input, "%d", &days)
		if err == nil && days > 0 && days <= internalApp.MaxDaysOffset {
			t := time.Now().AddDate(0, 0, days)
			return SetToEndOfDay(t).Unix()
		}
	}

	t, err := time.Parse("1/2/2006", input)
	if err == nil && IsDateInReasonableRange(t) {
		return SetToEndOfDay(t).Unix()
	}
	t, err = time.Parse("1/2/06", input)
	if err == nil && IsDateInReasonableRange(t) {
		return SetToEndOfDay(t).Unix()
	}

	t, err = time.Parse("1/2", input)
	if err == nil {
		year, month, day := t.Date()
		currentyear, currentmonth, currentday := time.Now().Date()
		if currentmonth > month || (currentmonth == month && currentday > day) {
			year = currentyear + 1
		} else {
			year = currentyear
		}
		t = time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		if IsDateInReasonableRange(t) {
			return SetToEndOfDay(t).Unix()
		}
	}

	return 0
}
