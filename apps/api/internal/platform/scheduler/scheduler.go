package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job is a unit of work executed repeatedly by the Scheduler.
type Job interface {
	Name() string
	Run(ctx context.Context) error
}

// JobFunc adapts a plain function into a Job.
type JobFunc struct {
	name string
	run  func(ctx context.Context) error
}

// NewJobFunc creates a Job from a name and run function.
func NewJobFunc(name string, run func(ctx context.Context) error) Job {
	return &JobFunc{name: name, run: run}
}

func (j *JobFunc) Name() string { return j.name }

func (j *JobFunc) Run(ctx context.Context) error { return j.run(ctx) }

// Scheduler runs registered jobs on a fixed interval.
// It is intentionally simple and in-process; scale-out or durability
// requirements should be revisited before adding River/Redis.
type Scheduler struct {
	interval time.Duration
	jobs     []Job
	stop     chan struct{}
	wg       sync.WaitGroup
}

// New creates a Scheduler that ticks every interval.
func New(interval time.Duration) *Scheduler {
	return &Scheduler{
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Register adds a job to the schedule.
func (s *Scheduler) Register(j Job) {
	s.jobs = append(s.jobs, j)
}

// Start begins the scheduler goroutine. It is a no-op if the interval
// is not positive or no jobs are registered.
func (s *Scheduler) Start() {
	if s.interval <= 0 || len(s.jobs) == 0 {
		slog.Warn("scheduler not started", "interval", s.interval, "jobs", len(s.jobs))
		return
	}
	s.wg.Add(1)
	go s.loop()
}

func (s *Scheduler) loop() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	slog.Info("scheduler started", "interval", s.interval.String(), "jobs", len(s.jobs))
	defer slog.Info("scheduler stopped")

	ctx := context.Background()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.runJobs(ctx)
		}
	}
}

func (s *Scheduler) runJobs(ctx context.Context) {
	for _, j := range s.jobs {
		if err := j.Run(ctx); err != nil {
			slog.Error("scheduler job failed", "job", j.Name(), "error", err)
		}
	}
}

// Stop signals the scheduler to stop and waits for the current tick to finish.
func (s *Scheduler) Stop() {
	close(s.stop)
	s.wg.Wait()
}
