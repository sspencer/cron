package cron

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// cronField represents a single field in the Spec.
type cronField struct {
	name   string
	min    uint64
	max    uint64
	values []string
}

// Parse parses a cron specification string and returns a Spec.
// It supports the standard 5-field Unix cron format and common shortcuts (e.g. "@weekly").
func Parse(spec string) (c Spec, err error) {
	cronFields := []cronField{
		{name: "minute", min: 0, max: 59},
		{name: "hour", min: 0, max: 23},
		{name: "dayOfMonth", min: 1, max: 31},
		{name: "month", min: 1, max: 12, values: monthValues},
		{name: "dayOfWeek", min: 0, max: 7, values: dayValues},
	}

	spec = strings.TrimSpace(spec)

	// substitute strings like "@monthly" for spec like "0 0 1 * *"
	if str, ok := shortcuts[spec]; ok {
		spec = str
	}

	s := strings.Fields(spec)
	if len(s) != len(cronFields) {
		return c, ErrCronParse
	}
	mode := s[2] != "*" && s[4] != "*"

	f := make([]uint64, len(cronFields))

	for i, cf := range cronFields {
		f[i], err = parseField(cf.name, s[i], cf.min, cf.max, cf.values)
		if err != nil {
			return c, err
		}
	}

	return Spec{
		minute:             f[0],
		hour:               f[1],
		dayOfMonth:         f[2],
		month:              f[3],
		dayOfWeek:          f[4],
		daysMatchingModeOR: mode,
	}, nil
}

// parseField parses separate cron specifications from a field and aggregates all errors if any.
func parseField(fieldName, fieldSpec string, minAllowed, maxAllowed uint64, keywords []string) (uint64, error) {
	var fieldBits uint64
	var errs MultiError

	specs := strings.SplitSeq(fieldSpec, ",")
	for spec := range specs {
		bit, err := parseSpec(spec, minAllowed, maxAllowed, keywords)
		if err != nil {
			// If a specification parsing error occurred, accumulate it in the
			// multi-error instead of returning immediately
			errs.Errors = append(errs.Errors, errors.Join(parseError(fieldName), err))
			continue
		}
		fieldBits |= bit
	}

	if len(errs.Errors) > 0 {
		return 0, &errs
	}
	return fieldBits, nil
}

// parseSpec parses the cron specification string and returns the corresponding bit representation.
// It takes the spec string to parse, the minimum and maximum allowed values for the field, and a slice of keywords.
// It returns the bit representation of the field and any parsing error encountered.
func parseSpec(spec string, min uint64, max uint64, keywords []string) (uint64, error) {
	var values uint64

	// Reject empty strings immediately.
	if spec == "" {
		return values, ErrCronParse
	}

	switch {
	case spec == "*":
		return handleAsterisk(values, min, max)
	case len(spec) >= 2 && strings.HasPrefix(spec, "*/"):
		return handleStep(spec[2:], min, max)
	case strings.Contains(spec, "-"):
		return handleRange(spec, min, max)
	case isNumberSpec(spec):
		return handleNumber(spec, min, max)
	case keywords != nil:
		return handleKeyword(spec, keywords)
	default:
		return values, ErrCronParse
	}
}

func handleAsterisk(values uint64, min uint64, max uint64) (uint64, error) {
	for i := min; i <= max; i++ {
		values = setBit(values, i)
	}
	return values, nil
}

func handleStep(step string, min uint64, max uint64) (uint64, error) {
	stepValue, err := strconv.Atoi(step)
	stepUint := uint64(stepValue)
	if stepUint == 0 || stepUint < min || stepUint > max || err != nil {
		return 0, ErrParseStep
	}
	var values uint64
	for i := min; i <= max; i += stepUint {
		values = setBit(values, i)
	}
	return values, nil
}

func handleRange(spec string, min uint64, max uint64) (uint64, error) {
	parts := strings.Split(spec, "-")
	initialRange, err1 := strconv.Atoi(parts[0])
	rangeStart := uint64(initialRange)
	finalRange, err2 := strconv.Atoi(parts[1])
	rangeEnd := uint64(finalRange)

	if rangeStart < min || err1 != nil || rangeEnd > max || err2 != nil || rangeStart >= rangeEnd {
		return 0, ErrParseRange
	}

	var values uint64
	for i := rangeStart; i <= rangeEnd; i++ {
		values = setBit(values, i)
	}
	return values, nil
}

func handleNumber(spec string, min uint64, max uint64) (uint64, error) {
	num, err := strconv.Atoi(spec)
	value := uint64(num)
	if value < min || value > max || err != nil {
		return 0, ErrParseNumber
	}
	return setBit(0, value), nil
}

func handleKeyword(spec string, keywords []string) (uint64, error) {
	index := sliceIndex(keywords, strings.ToLower(spec))
	if index == -1 {
		return 0, ErrParseKeyword
	}

	return setBit(0, uint64(index)), nil
}

func parseError(field string) error {
	return fmt.Errorf("error parsing %s field", field)
}

// MultiError stores multiple errors that occurred during parsing.
type MultiError struct {
	// Errors contains the list of parsing errors that occurred.
	Errors []error
}

// Error implements the error interface for MultiError.
func (m *MultiError) Error() string {
	var errs []string

	for _, err := range m.Errors {
		errs = append(errs, err.Error())
	}

	return strings.Join(errs, "; ")
}

// Unwrap allows errors.Is/As to inspect individual errors.
func (m *MultiError) Unwrap() []error {
	return m.Errors
}

// sliceIndex returns the index of the first occurrence of v in s,
// or -1 if not present.
func sliceIndex[S ~[]E, E comparable](s S, v E) int {
	for i := range s {
		if v == s[i] {
			return i
		}
	}
	return -1
}

func isNumberSpec(spec string) bool {
	if spec == "" {
		return false
	}
	_, err := strconv.Atoi(spec)
	return err == nil
}
