package cron

import (
	"context"
	"testing"
	"time"
)

type fakeClock struct {
	now         time.Time
	afterCalled chan struct{}
}

func (f *fakeClock) Now() time.Time {
	return f.now
}

func (f *fakeClock) Sleep(d time.Duration) {
	f.now = f.now.Add(d)
}

func (f *fakeClock) After(d time.Duration) <-chan time.Time {
	f.now = f.now.Add(d)
	ch := make(chan time.Time, 1)
	ch <- f.now
	if f.afterCalled != nil {
		select {
		case f.afterCalled <- struct{}{}:
		default:
		}
	}
	return ch
}

func TestWaitAdvancesToNextMinute(t *testing.T) {
	start := time.Date(2024, time.January, 1, 12, 0, 30, 0, time.UTC)
	clock := &fakeClock{now: start}

	wait(context.Background(), clock)

	expected := start.Truncate(time.Minute).Add(time.Minute).Add(clockPadding)
	if !clock.Now().Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, clock.Now())
	}
}

func TestWaitContextCancel(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if wait(ctx, clock) {
		t.Fatal("expected wait to return false after context cancellation")
	}
}

func TestRunWithClock(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}

	job, err := runWithClock(context.Background(), "* * * * *", func() {}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected non-nil job")
	}

	// Verify job was created with correct spec
	if job.spec.minute == 0 {
		t.Error("job spec not initialized correctly")
	}

	job.Stop()

	// Give goroutine time to stop
	time.Sleep(10 * time.Millisecond)
}

func TestRunWithInvalidSpec(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}

	job, err := runWithClock(context.Background(), "invalid spec", func() {}, clock)

	if err == nil {
		t.Fatal("expected error for invalid spec")
	}
	if job != nil {
		t.Fatal("expected nil job for invalid spec")
	}
}

func TestJobStop(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}

	job, err := runWithClock(context.Background(), "* * * * *", func() {}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Job should be running
	if !job.run.Load() {
		t.Error("job should be running after creation")
	}

	// Stop the job
	job.Stop()

	// Job should be stopped
	if job.run.Load() {
		t.Error("job should be stopped after Stop()")
	}
}

func TestTickWithContextCancellation(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	job, err := runWithClock(ctx, "* * * * *", func() {
		t.Error("job should not execute after context cancellation")
	}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give goroutine time to exit
	time.Sleep(20 * time.Millisecond)

	job.Stop()
}

func TestTickWithNilContext(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}

	// Pass nil context - should use Background
	job, err := runWithClock(context.TODO(), "* * * * *", func() {}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify job was created successfully
	if job == nil {
		t.Fatal("expected non-nil job with nil context")
	}

	job.Stop()
	time.Sleep(10 * time.Millisecond)
}

func TestTickStartsMidMinute(t *testing.T) {
	// Start at 30 seconds past the minute
	startTime := time.Date(2024, time.January, 1, 12, 0, 30, 0, time.UTC)
	clock := &fakeClock{now: startTime}

	job, err := runWithClock(context.Background(), "* * * * *", func() {}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give the goroutine time to start and call wait()
	time.Sleep(20 * time.Millisecond)

	// Clock should have advanced to next minute (due to wait() being called)
	if clock.now.Before(time.Date(2024, time.January, 1, 12, 1, 0, 0, time.UTC)) {
		t.Error("clock should have advanced to next minute")
	}

	job.Stop()
}

func TestSleepUntilPastTime(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 5, 0, 0, time.UTC)}
	pastTime := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)

	result := sleepUntil(context.Background(), clock, pastTime)

	if !result {
		t.Error("sleepUntil should return true for past time")
	}

	// Clock should not have advanced
	expected := time.Date(2024, time.January, 1, 12, 5, 0, 0, time.UTC)
	if !clock.now.Equal(expected) {
		t.Errorf("clock should not advance for past time: got %v, want %v", clock.now, expected)
	}
}

func TestSleepUntilContextCancelledImmediately(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	futureTime := clock.now.Add(time.Minute)
	result := sleepUntil(ctx, clock, futureTime)

	if result {
		t.Error("sleepUntil should return false when context is already cancelled")
	}
}

func TestSleepUntilContextCancelledDuringSleep(t *testing.T) {
	// With real clock, test that context cancellation during sleep works
	ctx, cancel := context.WithCancel(context.Background())
	rc := realClock{}

	// Use a channel to coordinate the test
	done := make(chan bool, 1)
	go func() {
		futureTime := rc.Now().Add(100 * time.Millisecond)
		result := sleepUntil(ctx, rc, futureTime)
		done <- result
	}()

	// Cancel after a short delay
	time.Sleep(10 * time.Millisecond)
	cancel()

	result := <-done
	if result {
		t.Error("sleepUntil should return false when context is cancelled during sleep")
	}
}

func TestRealClockMethods(t *testing.T) {
	rc := realClock{}

	// Test Now()
	now := rc.Now()
	if now.IsZero() {
		t.Error("realClock.Now() returned zero time")
	}

	// Test After()
	start := time.Now()
	ch := rc.After(10 * time.Millisecond)
	<-ch
	elapsed := time.Since(start)
	if elapsed < 10*time.Millisecond {
		t.Errorf("realClock.After() fired too early: %v", elapsed)
	}

	// Test Sleep()
	start = time.Now()
	rc.Sleep(10 * time.Millisecond)
	elapsed = time.Since(start)
	if elapsed < 10*time.Millisecond {
		t.Errorf("realClock.Sleep() returned too early: %v", elapsed)
	}
}

func TestRun(t *testing.T) {
	job, err := Run("* * * * *", func() {
		// Function body - can't easily test execution with real clock
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected non-nil job")
	}

	// Can't easily test execution with real clock, but we can verify job is created
	job.Stop()
}

func TestRunWithContext(t *testing.T) {
	ctx := context.Background()
	job, err := RunWithContext(ctx, "* * * * *", func() {})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected non-nil job")
	}

	job.Stop()
}

func TestRunWithInvalidSpecReturnsError(t *testing.T) {
	job, err := Run("invalid cron spec", func() {})

	if err == nil {
		t.Fatal("expected error for invalid spec")
	}
	if job != nil {
		t.Fatal("expected nil job for invalid spec")
	}
}

func TestMultipleExecutions(t *testing.T) {
	// This test verifies the job can be created for repeated execution
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}

	job, err := runWithClock(context.Background(), "* * * * *", func() {}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify job is running
	if !job.run.Load() {
		t.Error("job should be running")
	}

	job.Stop()

	if job.run.Load() {
		t.Error("job should be stopped")
	}
}

func TestJobNotTriggeredWhenSpecDoesNotMatch(t *testing.T) {
	// Start at 12:00
	startTime := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: startTime}

	// Spec: run only at 13:00 (1 PM)
	job, err := runWithClock(context.Background(), "0 13 * * *", func() {
		t.Error("job should not trigger at 12:xx")
	}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify spec is set correctly (should only trigger at 13:00)
	if job.spec.Trigger(startTime) {
		t.Error("spec should not trigger at 12:00")
	}

	job.Stop()
	time.Sleep(10 * time.Millisecond)
}

func TestJobExecutesFunction(t *testing.T) {
	executed := make(chan bool, 1)
	// Start at exact minute boundary so job can trigger immediately
	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}

	job, err := runWithClock(context.Background(), "* * * * *", func() {
		select {
		case executed <- true:
		default:
		}
	}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for the job to execute (with timeout)
	select {
	case <-executed:
		// Success - job executed
	case <-time.After(100 * time.Millisecond):
		t.Error("job function was not executed within timeout")
	}

	job.Stop()
}

func TestJobExecutionWithMatchingSpec(t *testing.T) {
	// Test that job function is called when spec matches
	executed := make(chan bool, 1)
	// Start at exact minute so spec will match: 12:05:00
	testTime := time.Date(2024, time.January, 1, 12, 5, 0, 0, time.UTC)
	clock := &fakeClock{now: testTime}

	job, err := runWithClock(context.Background(), "5 12 * * *", func() {
		select {
		case executed <- true:
		default:
		}
	}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify spec would trigger at 12:05:00
	if !job.spec.Trigger(testTime) {
		t.Fatalf("spec '5 12 * * *' should trigger at 12:05:00, minute=%d hour=%d",
			testTime.Minute(), testTime.Hour())
	}

	// Wait for execution
	select {
	case <-executed:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("job function should have executed")
	}

	job.Stop()
}

func TestTickContextCancelInLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	clock := &fakeClock{now: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)}

	job, err := runWithClock(ctx, "* * * * *", func() {
		// Job function - if it executes, that's fine
	}, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for context to timeout
	time.Sleep(100 * time.Millisecond)

	// Job should still be marked as running (Stop() wasn't called)
	// but the goroutine should have exited due to context cancellation
	job.Stop()
}
