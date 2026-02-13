package cron

import (
	"context"
	"sync/atomic"
	"time"
)

// clockPadding ensures we don't wake up slightly before the target minute.
// When sleeping until the next minute, we add this small buffer to avoid
// edge cases where time.After might wake us a few milliseconds early.
const clockPadding = 101 * time.Millisecond

// Func represents a function that can be scheduled and executed as a cron job.
type Func func()

// Job represents a cron job that runs at a specified schedule.
type Job struct {
	spec  Spec
	fn    Func
	run   atomic.Bool
	clock Clock
}

// Clock allows the scheduler to be tested with a deterministic time source.
type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
	After(time.Duration) <-chan time.Time
}

type realClock struct{}

func (realClock) Now() time.Time        { return time.Now() }
func (realClock) Sleep(d time.Duration) { time.Sleep(d) }
func (realClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// Run creates and starts a new cron job with the specified schedule and function.
// It returns a pointer to the newly created Job and an error, if any.
// The spec parameter is a string representing the cron schedule.
// The fn parameter is a function that will be executed according to the schedule.
// The returned Job can be stopped by calling the Stop method.
func Run(spec string, fn Func) (*Job, error) {
	return RunWithContext(context.Background(), spec, fn)
}

// RunWithContext creates and starts a new cron job with the specified schedule and function,
// cancelling execution when the provided context is done.
func RunWithContext(ctx context.Context, spec string, fn Func) (*Job, error) {
	return runWithClock(ctx, spec, fn, realClock{})
}

func runWithClock(ctx context.Context, spec string, fn Func, clock Clock) (*Job, error) {
	s, err := Parse(spec)
	if err != nil {
		return nil, err
	}

	c := &Job{
		spec:  s,
		fn:    fn,
		clock: clock,
	}

	c.run.Store(true)

	go c.tick(ctx)

	return c, nil
}

// Stop stops the execution of the Job.
func (c *Job) Stop() {
	c.run.Store(false)
}

// tick sleeps until the top of the minute then potentially runs the job.
func (c *Job) tick(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	if c.clock.Now().Second() > 0 {
		if !wait(ctx, c.clock) {
			return
		}
	}

	for c.run.Load() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if c.spec.Trigger(c.clock.Now()) {
			c.fn()
		}

		if !wait(ctx, c.clock) {
			return
		}
	}
}

// nextMinute calculates the time instance for the start of next minute.
func nextMinute(now time.Time) time.Time {
	// Calculate the next minute then add some padding to ensure we don't miss it.
	return now.Truncate(time.Minute).Add(time.Minute).Add(clockPadding)
}

// sleepUntil sleeps until the specified time, returning false if the context is cancelled.
func sleepUntil(ctx context.Context, clock Clock, t time.Time) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	d := t.Sub(clock.Now())
	if d <= 0 {
		return true
	}

	select {
	case <-ctx.Done():
		return false
	case <-clock.After(d):
		return true
	}
}

// wait sleeps until the top of the minute (0 seconds past the current minute).
func wait(ctx context.Context, clock Clock) bool {
	return sleepUntil(ctx, clock, nextMinute(clock.Now()))
}
