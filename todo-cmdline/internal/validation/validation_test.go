package validation

import (
	"fmt"
	"testing"
	"time"
)

func TestParseDueDate_EmptyString(t *testing.T) {
	result := ParseDueDate("")
	if result != 0 {
		t.Errorf("expected 0 for empty string, got %d", result)
	}
}

func TestParseDueDate_InvalidInput(t *testing.T) {
	tests := []string{
		"invalid",
		"abc",
		"13/32/2025",
		"-1",
		"0",
	}
	for _, input := range tests {
		result := ParseDueDate(input)
		if result != 0 {
			t.Errorf("expected 0 for invalid input %q, got %d", input, result)
		}
	}
}

func TestParseDueDate_DaysFromNow(t *testing.T) {
	tests := []struct {
		input string
		days  int
	}{
		{"1", 1},
		{"3", 3},
		{"7", 7},
		{"30", 30},
		{"999", 999},
	}

	for _, tt := range tests {
		result := ParseDueDate(tt.input)
		if result == 0 {
			t.Errorf("expected non-zero result for %q days", tt.input)
			continue
		}

		resultTime := time.Unix(result, 0)
		expected := time.Now().AddDate(0, 0, tt.days)

		if resultTime.Year() != expected.Year() ||
			resultTime.Month() != expected.Month() ||
			resultTime.Day() != expected.Day() {
			t.Errorf("for input %q: expected date %v, got %v",
				tt.input, expected.Format("2006-01-02"), resultTime.Format("2006-01-02"))
		}

		if resultTime.Hour() != 23 || resultTime.Minute() != 59 || resultTime.Second() != 59 {
			t.Errorf("for input %q: expected time 23:59:59, got %v",
				tt.input, resultTime.Format("15:04:05"))
		}
	}
}

// testDates returns dates relative to now so the tests never expire:
// ParseDueDate rejects dates more than 1 year in the past.
func testDates() []time.Time {
	now := time.Now()
	return []time.Time{
		now.AddDate(0, 0, 30),
		now.AddDate(0, 0, 180),
		now.AddDate(0, 0, -30), // recent past is allowed (up to 1 year back)
	}
}

func TestParseDueDate_FullDateFormat(t *testing.T) {
	for _, d := range testDates() {
		input := fmt.Sprintf("%d/%d/%d", int(d.Month()), d.Day(), d.Year())
		expected := d.Format("2006-01-02")

		result := ParseDueDate(input)
		if result == 0 {
			t.Errorf("expected non-zero result for %q", input)
			continue
		}

		resultTime := time.Unix(result, 0)
		resultDate := resultTime.Format("2006-01-02")
		if resultDate != expected {
			t.Errorf("for input %q: expected %s, got %s", input, expected, resultDate)
		}
	}
}

func TestParseDueDate_ShortYearFormat(t *testing.T) {
	for _, d := range testDates() {
		input := fmt.Sprintf("%d/%d/%02d", int(d.Month()), d.Day(), d.Year()%100)
		expected := d.Format("2006-01-02")

		result := ParseDueDate(input)
		if result == 0 {
			t.Errorf("expected non-zero result for %q", input)
			continue
		}

		resultTime := time.Unix(result, 0)
		resultDate := resultTime.Format("2006-01-02")
		if resultDate != expected {
			t.Errorf("for input %q: expected %s, got %s", input, expected, resultDate)
		}
	}
}

func TestParseDueDate_MonthDayOnly(t *testing.T) {
	currentYear := time.Now().Year()

	result := ParseDueDate("6/15")
	if result == 0 {
		t.Error("expected non-zero result for month/day format")
		return
	}

	resultTime := time.Unix(result, 0)
	resultYear := resultTime.Year()

	if resultYear != currentYear && resultYear != currentYear+1 {
		t.Errorf("expected year to be %d or %d, got %d", currentYear, currentYear+1, resultYear)
	}
}

func TestParseDueDate_TimeSetToEndOfDay(t *testing.T) {
	future := time.Now().AddDate(0, 0, 60)
	inputs := []string{
		"1",
		fmt.Sprintf("%d/%d/%d", int(future.Month()), future.Day(), future.Year()),
		fmt.Sprintf("%d/%d/%02d", int(future.Month()), future.Day(), future.Year()%100),
		"6/15",
	}

	for _, input := range inputs {
		result := ParseDueDate(input)
		if result == 0 {
			continue
		}

		resultTime := time.Unix(result, 0)
		if resultTime.Hour() != 23 || resultTime.Minute() != 59 || resultTime.Second() != 59 {
			t.Errorf("for input %q: expected time 23:59:59, got %s",
				input, resultTime.Format("15:04:05"))
		}
	}
}
