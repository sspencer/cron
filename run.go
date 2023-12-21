package cron

import (
	"sync/atomic"
	"time"
)

const durationPadding = 101 * time.Millisecond

// Func represents a function that can be scheduled and executed as a cron job.
type Func func()

// Job represents a cron job that runs at a specified schedule.
type Job struct {
	spec cronSpec
	fn   Func
	run  atomic.Bool
}

// Run creates and starts a new cron job with the specified schedule and function.
// It returns a pointer to the newly created Job and an error, if any.
// The spec parameter is a string representing the cron schedule.
// The fn parameter is a function that will be executed according to the schedule.
// The returned Job can be stopped by calling the Stop method.
func Run(spec string, fn Func) (*Job, error) {
	s, err := parse(spec)
	if err != nil {
		return nil, err
	}

	c := &Job{
		spec: s,
		fn:   fn,
	}

	c.run.Store(true)

	go c.tick()

	return c, nil
}

// Stop stops the execution of the Job.
func (c *Job) Stop() {
	c.run.Store(false)
}

// tick sleeps until top of minute then runs potentially runs job
func (c *Job) tick() {
	if time.Now().Second() > 0 {
		wait()
	}

	for c.run.Load() {
		if c.spec.trigger(time.Now()) {
			c.fn()
		}

		wait()
	}
}

// nextMinute calculates the time instance for the start of next minute
func nextMinute() time.Time {
	// calculate the nextMinute then add some padding to ensure we don't miss it
	return time.Now().Truncate(time.Minute).Add(time.Minute).Add(durationPadding)
}

// sleepUntil takes a time instance and causes the current routine to sleep until that time
func sleepUntil(t time.Time) {
	time.Sleep(t.Sub(time.Now()))
}

// wait sleeps until top of the minute (0 seconds past current minute)
func wait() {
	sleepUntil(nextMinute())
}
