package scheduler_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/scheduler"
)

func TestScheduler_RunsRegisteredJob(t *testing.T) {
	var count atomic.Int64
	s := scheduler.New(50 * time.Millisecond)
	s.Register(scheduler.NewJobFunc("counter", func(ctx context.Context) error {
		count.Add(1)
		return nil
	}))

	s.Start()
	time.Sleep(120 * time.Millisecond)
	s.Stop()

	if got := count.Load(); got < 1 {
		t.Fatalf("expected job to run at least once, got %d", got)
	}
}

func TestScheduler_StopWithoutStartDoesNotPanic(t *testing.T) {
	s := scheduler.New(50 * time.Millisecond)
	s.Register(scheduler.NewJobFunc("noop", func(ctx context.Context) error {
		return nil
	}))
	// Stop without Start should be safe.
	s.Stop()
}

func TestScheduler_NoJobsNoGoroutine(t *testing.T) {
	s := scheduler.New(50 * time.Millisecond)
	s.Start()
	// Nothing to wait for; just ensure no panic.
}

func TestScheduler_JobErrorDoesNotCrashScheduler(t *testing.T) {
	var count atomic.Int64
	s := scheduler.New(50 * time.Millisecond)
	s.Register(scheduler.NewJobFunc("flaky", func(ctx context.Context) error {
		count.Add(1)
		return errors.New("intentional failure")
	}))

	s.Start()
	time.Sleep(120 * time.Millisecond)
	s.Stop()

	if got := count.Load(); got < 1 {
		t.Fatalf("expected flaky job to keep running despite errors, got %d", got)
	}
}
