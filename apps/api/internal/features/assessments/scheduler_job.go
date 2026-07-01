package assessments

import (
	"context"
	"fmt"
	"log/slog"
)

// TransitionRepository is the minimal persistence surface needed by the
// assessment transition scheduler job.
type TransitionRepository interface {
	TransitionAssessmentsToOpen(ctx context.Context) (int64, error)
	TransitionAssessmentsToClosed(ctx context.Context) (int64, error)
}

// TransitionJob moves assessments between scheduled/published → open and
// open → closed based on opens_at/closes_at. It is safe to run concurrently
// with other operations because each execution uses a single UPDATE per
// transition; overlapping runs are idempotent.
type TransitionJob struct {
	repo TransitionRepository
}

// NewTransitionJob creates a scheduler job that transitions assessment statuses.
func NewTransitionJob(repo TransitionRepository) *TransitionJob {
	return &TransitionJob{repo: repo}
}

// Name returns the scheduler job name.
func (j *TransitionJob) Name() string {
	return "assessment-transition"
}

// Run executes the open and closed transitions and logs the row counts.
func (j *TransitionJob) Run(ctx context.Context) error {
	opened, err := j.repo.TransitionAssessmentsToOpen(ctx)
	if err != nil {
		return fmt.Errorf("open transition: %w", err)
	}

	closed, err := j.repo.TransitionAssessmentsToClosed(ctx)
	if err != nil {
		return fmt.Errorf("closed transition: %w", err)
	}

	slog.Info("assessment transitions completed", "opened", opened, "closed", closed)
	return nil
}
