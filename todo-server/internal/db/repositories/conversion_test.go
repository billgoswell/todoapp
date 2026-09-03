package repositories

import (
	"testing"
	"time"
)

// TestConvertTimestamp covers all value shapes that reach the upsert paths.
// Regression test: *int64 values (from the shared Task model's optional
// timestamp fields) previously fell through to the default case, silently
// erasing date_completed, due_date, and deleted_at on every push.
func TestConvertTimestamp(t *testing.T) {
	ts := int64(1720000000)
	var nilPtr *int64
	zero := int64(0)
	goTime := time.Unix(ts, 0)
	var nilTimePtr *time.Time

	cases := []struct {
		name string
		in   any
		want time.Time
	}{
		{"int64", ts, time.Unix(ts, 0)},
		{"int64 zero means unset", int64(0), time.Time{}},
		{"float64", float64(ts), time.Unix(ts, 0)},
		{"float64 zero means unset", float64(0), time.Time{}},
		{"pointer to int64", &ts, time.Unix(ts, 0)},
		{"nil pointer means unset", nilPtr, time.Time{}},
		{"pointer to zero means unset", &zero, time.Time{}},
		{"time.Time passthrough", goTime, goTime},
		{"pointer to time.Time", &goTime, goTime},
		{"nil time pointer means unset", nilTimePtr, time.Time{}},
		{"untyped nil", nil, time.Time{}},
		{"unsupported type means unset", "not a timestamp", time.Time{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := convertTimestamp(tc.in)
			if !got.Equal(tc.want) {
				t.Errorf("convertTimestamp(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
