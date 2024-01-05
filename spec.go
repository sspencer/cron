package cron

import (
	"errors"
	"fmt"
	"time"
)

var ErrCronParse = errors.New("cron specification error")
var ErrParseStep = errors.New("invalid step value")
var ErrParseRange = errors.New("invalid range value")
var ErrParseNumber = errors.New("invalid numeric value")
var ErrParseKeyword = errors.New("invalid keyword value")

// https://www.ibm.com/docs/en/db2/11.5?topic=task-unix-cron-format
type cronSpec struct {
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

func (c cronSpec) String() string {
	return fmt.Sprintf("minute=%b hour=%b dayOfMonth=%b month=%b dayOfWeek=%b",
		c.minute, c.hour, c.dayOfMonth, c.month, c.dayOfWeek)
}

var (
	monthValues = []string{"", "jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"}
	dayValues   = []string{"", "mon", "tue", "wed", "thu", "fri", "sat", "sun"}
	shortcuts   = map[string]string{
		"@yearly":   "0 0 1 1 *",
		"@annually": "0 0 1 1 *",
		"@monthly":  "0 0 1 * *",
		"@weekly":   "0 0 * * 0",
		"@daily":    "0 0 * * *",
		"@midnight": "0 0 * * *",
		"@hourly":   "0 * * * *",
	}
)

// trigger determines if the cronJob should run at specified time
func (c cronSpec) trigger(now time.Time) bool {
	return c.matchesTimeFields(now) && c.matchesDayFields(now)
}

// matchesTimeFields checks if the given `now` time matches the specified time fields in the cronSpec.
// It compares the minute, hour, and month fields of the cronSpec with the corresponding values of the `now` time.
func (c cronSpec) matchesTimeFields(now time.Time) bool {
	return isSet(c.minute, uint64(now.Minute())) && isSet(c.hour, uint64(now.Hour())) && isSet(c.month, uint64(now.Month()))
}

// matchesDayFields determines if the current day matches the day fields specified in the cron schedule
func (c cronSpec) matchesDayFields(now time.Time) bool {
	dayMatch := isSet(c.dayOfMonth, uint64(now.Day()))
	weekdayMatch := isSet(c.dayOfWeek, uint64(now.Weekday()))

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
