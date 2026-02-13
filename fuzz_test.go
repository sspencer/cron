package cron

import (
	"testing"
	"time"
)

func FuzzParseDoesNotPanic(f *testing.F) {
	seeds := []string{
		"* * * * *",
		"0 0 * * *",
		"*/5 1-3 1,15 jan mon",
		"@weekly",
		"@hourly",
		"not a cron",
		"1-3,5-7 * * * *",
		"0 0 1 1 *",
		"0 0 * * 7",
		"0 0 * * 0",
		"59 23 31 12 6",
		"0-59 0-23 1-31 1-12 0-6",
		"*/1 */1 */1 */1 */1",
		"@yearly",
		"@annually",
		"@monthly",
		"@daily",
		"@midnight",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, spec string) {
		parsed, err := Parse(spec)
		if err != nil {
			return
		}

		// Exercise trigger paths for valid specs.
		_ = parsed.Trigger(time.Now())
		_ = parsed.Trigger(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC))

		// Exercise Next method to ensure it doesn't panic or hang.
		now := time.Now()
		next := parsed.Next(now)
		// If Next returns a non-zero time, it should be after now.
		if !next.IsZero() && !next.After(now) {
			t.Errorf("Next(%v) = %v, expected time after now", now, next)
		}

		// Exercise String method to ensure it doesn't panic.
		_ = parsed.String()
	})
}

func FuzzParseFieldDoesNotPanic(f *testing.F) {
	seeds := []string{
		"*",
		"*/5",
		"1-10",
		"1,2,3",
		"jan",
		"mon",
		"0",
		"7",
		"60",
		"-1",
		"abc",
		"1-3,5-7",
		"*/15,10-12,27",
		"",
		"*/*",
		"1-",
		"-5",
		"1,2,3,4,5",
		"JAN,FEB,MAR",
		"sun,mon,tue",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, spec string) {
		// Test with different field configurations.
		minuteBits, minuteErr := parseField("minute", spec, 0, 59, nil)
		if minuteErr == nil {
			// Verify bits are within valid range for a uint64.
			// Since max value is 59, no bit > 59 should be set.
			for i := uint64(60); i < 64; i++ {
				if isSet(minuteBits, i) {
					t.Errorf("minute field set bit %d > max (59)", i)
				}
			}
		}

		monthBits, monthErr := parseField("month", spec, 1, 12, monthValues)
		if monthErr == nil {
			// Verify bits are within valid range (1-12).
			if isSet(monthBits, 0) {
				t.Error("month field set bit 0 (invalid)")
			}
			for i := uint64(13); i < 64; i++ {
				if isSet(monthBits, i) {
					t.Errorf("month field set bit %d > max (12)", i)
				}
			}
		}

		dayBits, dayErr := parseField("dayOfWeek", spec, 0, 7, dayValues)
		if dayErr == nil {
			// Verify bits are within valid range (0-7).
			for i := uint64(8); i < 64; i++ {
				if isSet(dayBits, i) {
					t.Errorf("dayOfWeek field set bit %d > max (7)", i)
				}
			}
		}
	})
}

func FuzzNextDoesNotHang(f *testing.F) {
	seeds := []string{
		"* * * * *",
		"0 0 * * *",
		"0 0 1 1 *",
		"59 23 31 12 6",
		"@hourly",
		"@daily",
		"0 0 29 2 *",   // Feb 29 (leap year edge case)
		"0 0 31 2 *",   // Invalid: Feb 31
		"0 0 31 4 *",   // Invalid: April 31
		"0 0 31 * *",   // Only months with 31 days
		"*/15 * * * *", // Every 15 minutes
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, spec string) {
		parsed, err := Parse(spec)
		if err != nil {
			return
		}

		// Test Next() from various starting times.
		testTimes := []time.Time{
			time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.February, 29, 23, 59, 0, 0, time.UTC), // Leap year
			time.Date(2023, time.December, 31, 23, 59, 0, 0, time.UTC),
			time.Now(),
		}

		for _, from := range testTimes {
			next := parsed.Next(from)

			// If Next returns a non-zero time, verify invariants.
			if !next.IsZero() {
				// Next time must be after from.
				if !next.After(from) {
					t.Errorf("Next(%v) = %v, expected time after from", from, next)
				}

				// The returned time should trigger the spec.
				if !parsed.Trigger(next) {
					t.Errorf("Next(%v) returned %v, but spec doesn't trigger at that time", from, next)
				}

				// Calling Next again should return a later or equal time.
				next2 := parsed.Next(next)
				if !next2.IsZero() && next2.Before(next) {
					t.Errorf("Next is not monotonic: Next(%v)=%v, Next(%v)=%v", from, next, next, next2)
				}
			}
		}
	})
}
