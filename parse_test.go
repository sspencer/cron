package cron

import "testing"

func TestParse(t *testing.T) {
	testCases := []struct {
		entry string
		valid bool
	}{
		{"* * * * *", true},
		{"*/5 * * * *", true},
		{"5,10 * * * *", true},
		{"*/15,10-12,27 * * * *", true},
		{"1-8,19,21 * * * *", true},
		{"* * * * cafe", false},
		{"1-8,19,67 * * * *", false},
		{"*/0 * * * *", false},
		{"*/-5 * * * *", false},
		{"*/65 * * * *", false},
		{"*/5-10 * * * *", false},
		{"*/ * * * *", false},
		{"5- * * * *", false},
		{"-5 * * * *", false},
		{"- * * * *", false},
		{"* * * *", false},
		{"@yearly", true},
		{"@annually", true},
		{"@monthly", true},
		{"@weekly", true},
		{"@daily", true},
		{"@midnight", true},
		{"@hourly", true},
		{"@secondly", false},
		{"some random words", false},
		{"here are five random words", false},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			_, err := Parse(tc.entry)
			if tc.valid && err != nil {
				t.Errorf("valid entry %q returned error", tc.entry)
			}
			if !tc.valid && err == nil {
				t.Errorf("invalid entry %q did not return error", tc.entry)
			}
		})
	}
}

func TestBits(t *testing.T) {
	testCases := []struct {
		entry string
		bits  []uint64
	}{
		{"*/5 * * * *", []uint64{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55}},
		{"1-3,5-7 * * * *", []uint64{1, 2, 3, 5, 6, 7}},
		{"1-3,5-7,9,11 * * * *", []uint64{1, 2, 3, 5, 6, 7, 9, 11}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := Parse(tc.entry)
			if err != nil {
				t.Errorf("spec %s failed to parse", err)
			}
			for _, b := range tc.bits {
				if !isSet(spec.minute, b) {
					t.Errorf("bit %d not set %b", b, spec.minute)
				}
			}

			expected := len(tc.bits)
			actual := numBits(spec.minute)
			if numBits(spec.minute) != len(tc.bits) {
				t.Errorf("expecting %d bits set, not %d", expected, actual)
			}
		})
	}
}

func TestMinuteSpec(t *testing.T) {
	testCases := []struct {
		entry    string
		expected Spec
	}{
		{"0 * * * *", Spec{minute: 1}},
		{"1 * * * *", Spec{minute: 2}},
		{"2 * * * *", Spec{minute: 4}},
		{"1,2 * * * *", Spec{minute: 6}},
		{"1-3 * * * *", Spec{minute: 14}},
		{"@hourly", Spec{minute: 1}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := Parse(tc.entry)
			if err != nil {
				t.Errorf("spec %s failed to parse", err)
			}

			if tc.expected.minute != spec.minute {
				t.Errorf("expecting minute %d not %d", tc.expected.minute, spec.minute)
			}

		})
	}
}

func TestHourSpec(t *testing.T) {
	testCases := []struct {
		entry    string
		expected Spec
	}{
		{"* 0 * * *", Spec{hour: 1}},
		{"* 1 * * *", Spec{hour: 2}},
		{"* 2 * * *", Spec{hour: 4}},
		{"* 1,2 * * *", Spec{hour: 6}},
		{"* 1-3 * * *", Spec{hour: 14}},
		{"@monthly", Spec{hour: 1}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := Parse(tc.entry)
			if err != nil {
				t.Errorf("spec %s failed to parse", err)
			}

			if tc.expected.hour != spec.hour {
				t.Errorf("expecting hour %d not %d", tc.expected.hour, spec.hour)
			}

		})
	}
}

func TestDayOfMonthSpec(t *testing.T) {
	testCases := []struct {
		entry    string
		expected Spec
	}{
		{"* * 1 * *", Spec{dayOfMonth: 2}},
		{"* * 2 * *", Spec{dayOfMonth: 4}},
		{"* * 1,2 * *", Spec{dayOfMonth: 6}},
		{"* * 1-3 * *", Spec{dayOfMonth: 14}},
		{"@monthly", Spec{dayOfMonth: 2}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := Parse(tc.entry)
			if err != nil {
				t.Errorf("spec %s failed to parse", err)
			}

			if tc.expected.dayOfMonth != spec.dayOfMonth {
				t.Errorf("expecting dayOfMonth %d not %d", tc.expected.dayOfMonth, spec.dayOfMonth)
			}

		})
	}
}

func TestMonthSpec(t *testing.T) {
	testCases := []struct {
		entry    string
		expected Spec
	}{
		{"* * * 1 *", Spec{month: 2}},
		{"* * * 2 *", Spec{month: 4}},
		{"* * * 1,2 *", Spec{month: 6}},
		{"* * * 1-3 *", Spec{month: 14}},
		{"* * * jan *", Spec{month: 2}},
		{"* * * feb *", Spec{month: 4}},
		{"* * * mar *", Spec{month: 8}},
		{"* * * apr *", Spec{month: 16}},
		{"* * * may *", Spec{month: 32}},
		{"* * * jun *", Spec{month: 64}},
		{"* * * jul *", Spec{month: 128}},
		{"* * * aug *", Spec{month: 256}},
		{"* * * sep *", Spec{month: 512}},
		{"* * * oct *", Spec{month: 1024}},
		{"* * * nov *", Spec{month: 2048}},
		{"* * * dec *", Spec{month: 4096}},
		{"* * * Feb *", Spec{month: 4}},
		{"* * * fEB *", Spec{month: 4}},
		{"* * * FEB *", Spec{month: 4}},
		{"@yearly", Spec{month: 2}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := Parse(tc.entry)
			if err != nil {
				t.Errorf("spec %s failed to parse", err)
			}

			if tc.expected.month != spec.month {
				t.Errorf("expecting month %d not %d", tc.expected.month, spec.month)
			}

		})
	}
}

func TestDayOfWeekSpec(t *testing.T) {
	testCases := []struct {
		entry    string
		expected Spec
	}{
		{"* * * * *", Spec{dayOfWeek: 255}},
		{"* * * * 1", Spec{dayOfWeek: 2}},
		{"* * * * 2", Spec{dayOfWeek: 4}},
		{"* * * * 1,2", Spec{dayOfWeek: 6}},
		{"* * * * 1-3", Spec{dayOfWeek: 14}},
		{"* * * * sun", Spec{dayOfWeek: 128}},
		{"* * * * tue", Spec{dayOfWeek: 4}},
		{"* * * * wed", Spec{dayOfWeek: 8}},
		{"* * * * thu", Spec{dayOfWeek: 16}},
		{"* * * * fri", Spec{dayOfWeek: 32}},
		{"* * * * sat", Spec{dayOfWeek: 64}},
		{"* * * * mon,tue", Spec{dayOfWeek: 6}},
		{"@weekly", Spec{dayOfWeek: 1}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := Parse(tc.entry)
			if err != nil {
				t.Errorf("spec %s failed to parse", err)
			}

			if tc.expected.dayOfWeek != spec.dayOfWeek {
				t.Errorf("expecting dayOfWeek %d not %d", tc.expected.dayOfWeek, spec.dayOfWeek)
			}

		})
	}
}

func TestMultiErrorError(t *testing.T) {
	err1 := parseError("minute")
	err2 := parseError("hour")
	err3 := ErrParseStep

	multiErr := &MultiError{
		Errors: []error{err1, err2, err3},
	}

	errStr := multiErr.Error()
	if errStr == "" {
		t.Error("MultiError.Error() returned empty string")
	}

	// Should contain information about all errors
	if !contains(errStr, "minute") {
		t.Errorf("error string should contain 'minute': %s", errStr)
	}
	if !contains(errStr, "hour") {
		t.Errorf("error string should contain 'hour': %s", errStr)
	}
}

func TestMultiErrorUnwrap(t *testing.T) {
	err1 := parseError("minute")
	err2 := parseError("hour")

	multiErr := &MultiError{
		Errors: []error{err1, err2},
	}

	unwrapped := multiErr.Unwrap()
	if len(unwrapped) != 2 {
		t.Errorf("expected 2 unwrapped errors, got %d", len(unwrapped))
	}

	if unwrapped[0] != err1 {
		t.Errorf("first unwrapped error doesn't match")
	}
	if unwrapped[1] != err2 {
		t.Errorf("second unwrapped error doesn't match")
	}
}

func TestMultiErrorEmpty(t *testing.T) {
	multiErr := &MultiError{
		Errors: []error{},
	}

	errStr := multiErr.Error()
	if errStr != "" {
		t.Errorf("expected empty string for empty MultiError, got: %s", errStr)
	}

	unwrapped := multiErr.Unwrap()
	if len(unwrapped) != 0 {
		t.Errorf("expected 0 unwrapped errors, got %d", len(unwrapped))
	}
}

func TestIsNumberSpec(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"0", true},
		{"1", true},
		{"42", true},
		{"999", true},
		{"-1", true},
		{"-999", true},
		{"", false},
		{"abc", false},
		{"12abc", false},
		{"abc12", false},
		{"1.5", false},
		{"*", false},
		{"1-3", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isNumberSpec(tc.input)
			if result != tc.expected {
				t.Errorf("isNumberSpec(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestParseEmptyFieldInCommaList(t *testing.T) {
	// Test that empty strings in comma-separated lists are rejected
	testCases := []string{
		",1,2 * * * *", // Leading comma
		"1,,2 * * * *", // Double comma
		"1,2, * * * *", // Trailing comma
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			_, err := Parse(tc)
			if err == nil {
				t.Errorf("Parse(%q) should return error for empty field in comma list", tc)
			}
		})
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
