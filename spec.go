// Package cron provides a simple library to schedule functions to run periodically
// using Unix cron format. It supports the standard 5-field cron syntax:
//
//   - * * * *
//     | | | | |
//     | | | | +- Day of week (0–7) (Sunday=0 or 7) or Sun, Mon, Tue,…
//     | | | +--- Month (1–12) or Jan, Feb,…
//     | | +----- Day of month (1–31)
//     | +------- Hour (0–23)
//     +--------- Minute (0–59)
//
// The package also supports common shortcuts like @hourly, @daily, @weekly, @monthly, and @yearly.
//
// Example usage:
//
//	c, err := cron.Run("*/5 * * * *", func() {
//	    fmt.Println("This runs every 5 minutes")
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer c.Stop()
package cron

import (
	"errors"
	"fmt"
	"time"
)

// ErrCronParse is returned when a cron specification cannot be parsed.
var ErrCronParse = errors.New("cron specification error")

// ErrParseStep is returned when a step value is invalid.
var ErrParseStep = errors.New("invalid step value")

// ErrParseRange is returned when a range value is invalid.
var ErrParseRange = errors.New("invalid range value")

// ErrParseNumber is returned when a numeric value is invalid.
var ErrParseNumber = errors.New("invalid numeric value")

// ErrParseKeyword is returned when a keyword value is invalid.
var ErrParseKeyword = errors.New("invalid keyword value")

// Spec represents a parsed cron schedule.
// https://www.ibm.com/docs/en/db2/11.5?topic=task-unix-cron-format
type Spec struct {
	minute             uint64
	hour               uint64
	dayOfMonth         uint64
	month              uint64
	dayOfWeek          uint64
	daysMatchingModeOR bool // true when both dayOfMonth + dayOfWeek are restricted

	// From doc:
	// The day of a command's execution can be specified by two fields: day of month and day of week.
	// If both fields are restricted by the use of a value other than the asterisk, the command will
	// run when either field matches the current time. For example, the value 30 4 1,15 * 5 causes a
	// command to run at 4:30 AM on the 1st and 15th of each month, plus every Friday.
	// And:
	// "32 18 17,21,29 11 mon,wed"
	// 6.32 PM on the 17th, 21st and 29th of November plus each Monday and Wednesday in November each year
}

// String returns a human-readable representation of the Spec.
func (c Spec) String() string {
	return fmt.Sprintf("minute=%b hour=%b dayOfMonth=%b month=%b dayOfWeek=%b",
		c.minute, c.hour, c.dayOfMonth, c.month, c.dayOfWeek)
}

var (
	// monthValues maps month keywords to their numeric positions (1-12).
	// Index 0 is empty since months are 1-indexed.
	monthValues = []string{"", "jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"}

	// dayValues maps day-of-week keywords to their numeric positions (1-7).
	// Index 0 is empty since Monday=1, ..., Sunday=7 in these keywords.
	dayValues = []string{"", "mon", "tue", "wed", "thu", "fri", "sat", "sun"}

	// shortcuts defines common cron schedule aliases.
	shortcuts = map[string]string{
		"@yearly":   "0 0 1 1 *",
		"@annually": "0 0 1 1 *",
		"@monthly":  "0 0 1 * *",
		"@weekly":   "0 0 * * 0",
		"@daily":    "0 0 * * *",
		"@midnight": "0 0 * * *",
		"@hourly":   "0 * * * *",
	}
)

// Trigger reports whether the spec matches the given time.
// It returns true if the cron job should run at the specified time.
func (c Spec) Trigger(now time.Time) bool {
	return c.matchesTimeFields(now) && c.matchesDayFields(now)
}

// Next returns the next time after the provided instant that matches the spec.
// If no matching time is found within a year, the zero time is returned.
func (c Spec) Next(from time.Time) time.Time {
	t := from.Truncate(time.Minute).Add(time.Minute)
	const maxMinutes = 366 * 24 * 60
	for range maxMinutes {
		if c.Trigger(t) {
			return t
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}
}

// matchesTimeFields checks if the given `now` time matches the specified time fields in the Spec.
// It compares the minute, hour, and month fields of the Spec with the corresponding values of the `now` time.
func (c Spec) matchesTimeFields(now time.Time) bool {
	return isSet(c.minute, uint64(now.Minute())) && isSet(c.hour, uint64(now.Hour())) && isSet(c.month, uint64(now.Month()))
}

// matchesDayFields determines if the current day matches the day fields specified in the cron schedule
func (c Spec) matchesDayFields(now time.Time) bool {
	dayMatch := isSet(c.dayOfMonth, uint64(now.Day()))
	weekday := uint64(now.Weekday())
	weekdayMatch := isSet(c.dayOfWeek, weekday)
	// In Unix cron, Sunday can be represented by 0 or 7.
	if weekday == 0 {
		weekdayMatch = weekdayMatch || isSet(c.dayOfWeek, 7)
	}

	if c.daysMatchingModeOR {
		return dayMatch || weekdayMatch
	}
	return dayMatch && weekdayMatch
}

// setBit sets the specific bit position in the given number and returns the modified number.
func setBit(n uint64, pos uint64) uint64 {
	n |= 1 << pos
	return n
}

// isSet checks if a specific bit position is set in the given number.
func isSet(n uint64, pos uint64) bool {
	val := n & (1 << pos)
	return val > 0
}

// numBits counts the number of set bits (1s) in the binary representation of the given number.
func numBits(n uint64) int {
	var count uint64
	for n != 0 {
		count += n & 1
		n >>= 1
	}
	return int(count)
}
