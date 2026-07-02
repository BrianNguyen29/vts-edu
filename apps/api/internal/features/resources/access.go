package resources

import (
	"context"
	"errors"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
)

// ClassAccessChecker lets the resources feature ask the academics feature
// whether an actor has access to a class-scoped resource without coupling
// the two packages tightly. Implementations live in cmd/server.
type ClassAccessChecker interface {
	// ClassExists returns true when the class belongs to the actor's org.
	ClassExists(ctx context.Context, orgID, classID string) (bool, error)
	// CanViewClass returns true if the actor is allowed to view class-scoped
	// content (admin in the org, teacher assigned to the class, or
	// student enrolled in the class).
	CanViewClass(ctx context.Context, actor auth.Actor, classID string) (bool, error)
	// CanManageClass returns true if the actor is allowed to create or
	// upload to a class-scoped resource (admin or teacher assigned to the
	// class).
	CanManageClass(ctx context.Context, actor auth.Actor, classID string) (bool, error)
}

// stubChecker is the no-op default used by unit tests that do not exercise
// class-scoped behaviour. It denies all class access so that the test path
// remains an org-scoped one unless explicitly overridden.
type stubChecker struct{}

func (stubChecker) ClassExists(ctx context.Context, orgID, classID string) (bool, error) {
	return false, nil
}
func (stubChecker) CanViewClass(ctx context.Context, actor auth.Actor, classID string) (bool, error) {
	return false, nil
}
func (stubChecker) CanManageClass(ctx context.Context, actor auth.Actor, classID string) (bool, error) {
	return false, nil
}

// errClassAccessUnavailable is returned by service helpers when the checker
// could not determine access (e.g. database error). Surfaced as 502 by the
// handler so that the client gets a clear retryable signal.
var errClassAccessUnavailable = errors.New("class access check unavailable")
