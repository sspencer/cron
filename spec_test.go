package cron

import (
	"errors"
	"testing"
	"time"
)

func TestMatchesDayFieldsAndOr(t *testing.T) {
	spec := Spec{
		dayOfMonth: setBit(0, 15),
		dayOfWeek:  setBit(0, 1), // Monday
	}

	t.Run("and-mode", func(t *testing.T) {
		spec.daysMatchingModeOR = false
		if !spec.matchesDayFields(time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)) {
			t.Fatal("expected match when day-of-month and weekday both match")
		}
		if spec.matchesDayFields(time.Date(2024, time.February, 15, 0, 0, 0, 0, time.UTC)) {
			t.Fatal("unexpected match when only day-of-month matches")
		}
		if spec.matchesDayFields(time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)) {
			t.Fatal("unexpected match when only weekday matches")
		}
	})

	t.Run("or-mode", func(t *testing.T) {
		spec.daysMatchingModeOR = true
		if !spec.matchesDayFields(time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)) {
			t.Fatal("expected match when both fields match")
		}
		if !spec.matchesDayFields(time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)) {
			t.Fatal("expected match when only weekday matches")
		}
		if !spec.matchesDayFields(time.Date(2024, time.February, 15, 0, 0, 0, 0, time.UTC)) {
			t.Fatal("expected match when only day-of-month matches")
		}
	})
}

func TestMatchesDayFieldsSundayAlias(t *testing.T) {
	sunday := time.Date(2024, time.January, 7, 0, 0, 0, 0, time.UTC) // Sunday

	dayAll := allBits(1, 31)
	specZero := Spec{dayOfMonth: dayAll, dayOfWeek: setBit(0, 0)}
	if !specZero.matchesDayFields(sunday) {
		t.Fatal("expected Sunday to match day-of-week bit 0")
	}

	specSeven := Spec{dayOfMonth: dayAll, dayOfWeek: setBit(0, 7)}
	if !specSeven.matchesDayFields(sunday) {
		t.Fatal("expected Sunday to match day-of-week bit 7")
	}
}

func TestTrigger(t *testing.T) {
	spec, err := Parse("0 12 * jan mon")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	match := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC) // Monday
	if !spec.Trigger(match) {
		t.Fatal("expected trigger to match")
	}

	if spec.Trigger(match.Add(time.Minute)) {
		t.Fatal("unexpected match when minute does not match")
	}

	if spec.Trigger(time.Date(2024, time.February, 5, 12, 0, 0, 0, time.UTC)) {
		t.Fatal("unexpected match when month does not match")
	}
}

func TestParseFieldMultiError(t *testing.T) {
	_, err := parseField("minute", "1-3,abc,5-", 0, 59, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	var multi *MultiError
	if !errors.As(err, &multi) {
		t.Fatalf("expected MultiError, got %T", err)
	}
	if len(multi.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(multi.Errors))
	}
}

func TestParseDayOfMonthStepRespectsMin(t *testing.T) {
	spec, err := Parse("0 0 */1 * *")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if isSet(spec.dayOfMonth, 0) {
		t.Fatal("did not expect day-of-month bit 0 to be set")
	}
	if !isSet(spec.dayOfMonth, 1) {
		t.Fatal("expected day-of-month bit 1 to be set")
	}
}

func TestParseRejectsExtraWhitespaceFields(t *testing.T) {
	if _, err := Parse("0  0 * * *"); err != nil {
		t.Fatalf("unexpected error for spec with extra whitespace: %v", err)
	}
	if _, err := Parse("0\t0\t*\t*\t*"); err != nil {
		t.Fatalf("unexpected error for tab-separated spec: %v", err)
	}
}

func allBits(min, max uint64) uint64 {
	var bits uint64
	for i := min; i <= max; i++ {
		bits = setBit(bits, i)
	}
	return bits
}

func TestNext(t *testing.T) {
	spec, err := Parse("*/15 * * * *")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	start := time.Date(2024, time.January, 1, 12, 7, 0, 0, time.UTC)
	next := spec.Next(start)
	expected := time.Date(2024, time.January, 1, 12, 15, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}

	midnightSpec, err := Parse("0 0 * * *")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	start = time.Date(2024, time.January, 1, 23, 59, 30, 0, time.UTC)
	next = midnightSpec.Next(start)
	expected = time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}
