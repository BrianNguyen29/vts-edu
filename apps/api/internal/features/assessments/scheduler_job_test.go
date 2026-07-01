package assessments_test

import (
	"context"
	"errors"
	"testing"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/assessments"
)

type transitionRepoSpy struct {
	openCalled   bool
	openCount    int64
	openErr      error
	closedCalled bool
	closedCount  int64
	closedErr    error
}

func (s *transitionRepoSpy) TransitionAssessmentsToOpen(ctx context.Context) (int64, error) {
	s.openCalled = true
	return s.openCount, s.openErr
}

func (s *transitionRepoSpy) TransitionAssessmentsToClosed(ctx context.Context) (int64, error) {
	s.closedCalled = true
	return s.closedCount, s.closedErr
}

func TestTransitionJob_RunsBothTransitions(t *testing.T) {
	spy := &transitionRepoSpy{openCount: 3, closedCount: 5}
	job := assessments.NewTransitionJob(spy)

	if got := job.Name(); got != "assessment-transition" {
		t.Fatalf("job name = %q, want assessment-transition", got)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !spy.openCalled {
		t.Fatal("expected TransitionAssessmentsToOpen to be called")
	}
	if !spy.closedCalled {
		t.Fatal("expected TransitionAssessmentsToClosed to be called")
	}
}

func TestTransitionJob_OpenErrorReturned(t *testing.T) {
	spy := &transitionRepoSpy{openErr: errors.New("open failed")}
	job := assessments.NewTransitionJob(spy)

	if err := job.Run(context.Background()); err == nil {
		t.Fatal("expected error from open transition")
	}
	if spy.closedCalled {
		t.Fatal("closed transition should not run when open fails")
	}
}

func TestTransitionJob_ClosedErrorReturned(t *testing.T) {
	spy := &transitionRepoSpy{closedErr: errors.New("closed failed")}
	job := assessments.NewTransitionJob(spy)

	if err := job.Run(context.Background()); err == nil {
		t.Fatal("expected error from closed transition")
	}
	if !spy.openCalled {
		t.Fatal("expected open transition to run before closed transition")
	}
}
