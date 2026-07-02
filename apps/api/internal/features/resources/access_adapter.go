package resources

import (
	"context"
	"errors"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/academics"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
)

// AcademicAccessAdapter implements ClassAccessChecker by delegating to the
// academics repository. The resources package imports a small concrete
// adapter (rather than depending on a global interface) so the
// dependency direction stays one-way.
type AcademicAccessAdapter struct {
	Repo               academics.Repository
	MembershipResolver func(ctx context.Context, orgID, userID string) (string, error)
}

// NewAcademicAccessAdapter wires a default adapter that resolves the
// actor's membership id via academics.GetMembershipByUserID.
func NewAcademicAccessAdapter(repo academics.Repository) *AcademicAccessAdapter {
	return &AcademicAccessAdapter{
		Repo: repo,
		MembershipResolver: func(ctx context.Context, orgID, userID string) (string, error) {
			info, err := repo.GetMembershipByUserID(ctx, orgID, userID)
			if err != nil {
				if errors.Is(err, academics.ErrUserNotFound) {
					return "", nil
				}
				return "", err
			}
			return info.ID, nil
		},
	}
}

func (a *AcademicAccessAdapter) ClassExists(ctx context.Context, orgID, classID string) (bool, error) {
	return a.Repo.ClassExists(ctx, orgID, classID)
}

func (a *AcademicAccessAdapter) CanViewClass(ctx context.Context, actor auth.Actor, classID string) (bool, error) {
	if actor.OrgID == "" || classID == "" {
		return false, nil
	}
	if hasRole(actor.Roles, "admin") {
		return true, nil
	}
	if hasRole(actor.Roles, "teacher") {
		membershipID, err := a.membershipID(ctx, actor)
		if err != nil {
			return false, err
		}
		if membershipID == "" {
			return false, nil
		}
		return a.Repo.IsClassTeacher(ctx, actor.OrgID, classID, membershipID)
	}
	// Student: must be enrolled in the class.
	return a.Repo.IsStudentEnrolled(ctx, actor.OrgID, classID, actor.UserID)
}

func (a *AcademicAccessAdapter) CanManageClass(ctx context.Context, actor auth.Actor, classID string) (bool, error) {
	if actor.OrgID == "" || classID == "" {
		return false, nil
	}
	if hasRole(actor.Roles, "admin") {
		return true, nil
	}
	if !hasRole(actor.Roles, "teacher") {
		return false, nil
	}
	membershipID, err := a.membershipID(ctx, actor)
	if err != nil {
		return false, err
	}
	if membershipID == "" {
		return false, nil
	}
	return a.Repo.IsClassTeacher(ctx, actor.OrgID, classID, membershipID)
}

func (a *AcademicAccessAdapter) membershipID(ctx context.Context, actor auth.Actor) (string, error) {
	if a.MembershipResolver == nil {
		return "", errors.New("membership resolver not configured")
	}
	return a.MembershipResolver(ctx, actor.OrgID, actor.UserID)
}

// ensure interface satisfaction
var _ ClassAccessChecker = (*AcademicAccessAdapter)(nil)
