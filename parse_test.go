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
			_, err := parse(tc.entry)
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
			spec, err := parse(tc.entry)
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
		expected cronSpec
	}{
		{"0 * * * *", cronSpec{minute: 1}},
		{"1 * * * *", cronSpec{minute: 2}},
		{"2 * * * *", cronSpec{minute: 4}},
		{"1,2 * * * *", cronSpec{minute: 6}},
		{"1-3 * * * *", cronSpec{minute: 14}},
		{"@hourly", cronSpec{minute: 1}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := parse(tc.entry)
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
		expected cronSpec
	}{
		{"* 0 * * *", cronSpec{hour: 1}},
		{"* 1 * * *", cronSpec{hour: 2}},
		{"* 2 * * *", cronSpec{hour: 4}},
		{"* 1,2 * * *", cronSpec{hour: 6}},
		{"* 1-3 * * *", cronSpec{hour: 14}},
		{"@monthly", cronSpec{hour: 1}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := parse(tc.entry)
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
		expected cronSpec
	}{
		{"* * 1 * *", cronSpec{dayOfMonth: 2}},
		{"* * 2 * *", cronSpec{dayOfMonth: 4}},
		{"* * 1,2 * *", cronSpec{dayOfMonth: 6}},
		{"* * 1-3 * *", cronSpec{dayOfMonth: 14}},
		{"@monthly", cronSpec{dayOfMonth: 2}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := parse(tc.entry)
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
		expected cronSpec
	}{
		{"* * * 1 *", cronSpec{month: 2}},
		{"* * * 2 *", cronSpec{month: 4}},
		{"* * * 1,2 *", cronSpec{month: 6}},
		{"* * * 1-3 *", cronSpec{month: 14}},
		{"* * * jan *", cronSpec{month: 2}},
		{"* * * feb *", cronSpec{month: 4}},
		{"* * * mar *", cronSpec{month: 8}},
		{"* * * apr *", cronSpec{month: 16}},
		{"* * * may *", cronSpec{month: 32}},
		{"* * * jun *", cronSpec{month: 64}},
		{"* * * jul *", cronSpec{month: 128}},
		{"* * * aug *", cronSpec{month: 256}},
		{"* * * sep *", cronSpec{month: 512}},
		{"* * * oct *", cronSpec{month: 1024}},
		{"* * * nov *", cronSpec{month: 2048}},
		{"* * * dec *", cronSpec{month: 4096}},
		{"* * * Feb *", cronSpec{month: 4}},
		{"* * * fEB *", cronSpec{month: 4}},
		{"* * * FEB *", cronSpec{month: 4}},
		{"@yearly", cronSpec{month: 2}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := parse(tc.entry)
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
		expected cronSpec
	}{
		{"* * * * *", cronSpec{dayOfWeek: 255}},
		{"* * * * 1", cronSpec{dayOfWeek: 2}},
		{"* * * * 2", cronSpec{dayOfWeek: 4}},
		{"* * * * 1,2", cronSpec{dayOfWeek: 6}},
		{"* * * * 1-3", cronSpec{dayOfWeek: 14}},
		{"* * * * sun", cronSpec{dayOfWeek: 128}},
		{"* * * * tue", cronSpec{dayOfWeek: 4}},
		{"* * * * wed", cronSpec{dayOfWeek: 8}},
		{"* * * * thu", cronSpec{dayOfWeek: 16}},
		{"* * * * fri", cronSpec{dayOfWeek: 32}},
		{"* * * * sat", cronSpec{dayOfWeek: 64}},
		{"* * * * mon,tue", cronSpec{dayOfWeek: 6}},
		{"@weekly", cronSpec{dayOfWeek: 1}},
	}

	for _, tc := range testCases {
		t.Run(tc.entry, func(t *testing.T) {
			spec, err := parse(tc.entry)
			if err != nil {
				t.Errorf("spec %s failed to parse", err)
			}

			if tc.expected.dayOfWeek != spec.dayOfWeek {
				t.Errorf("expecting dayOfWeek %d not %d", tc.expected.dayOfWeek, spec.dayOfWeek)
			}

		})
	}
}
