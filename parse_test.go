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
