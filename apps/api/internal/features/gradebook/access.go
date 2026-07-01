package gradebook

import (
	"context"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/academics"
)

// AcademicAccessAdapter adapts the academics repository to the ClassAccessChecker contract.
type AcademicAccessAdapter struct {
	Repo academics.Repository
}

// GetMembershipByUserID returns the membership for the user.
func (a *AcademicAccessAdapter) GetMembershipByUserID(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
	m, err := a.Repo.GetMembershipByUserID(ctx, orgID, userID)
	if err != nil {
		return MembershipInfo{}, err
	}
	return MembershipInfo{ID: m.ID}, nil
}

// IsClassTeacher checks whether the membership is a teacher of the class.
func (a *AcademicAccessAdapter) IsClassTeacher(ctx context.Context, orgID, classID, membershipID string) (bool, error) {
	return a.Repo.IsClassTeacher(ctx, orgID, classID, membershipID)
}

// ClassExists checks whether the class exists in the organization.
func (a *AcademicAccessAdapter) ClassExists(ctx context.Context, orgID, classID string) (bool, error) {
	return a.Repo.ClassExists(ctx, orgID, classID)
}

// Ensure the adapter satisfies the interface.
var _ ClassAccessChecker = (*AcademicAccessAdapter)(nil)
