package admin

import (
	"context"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/grading"
	"github.com/jackc/pgx/v5"
)

// GradingAuditAdapter exposes a minimal AuditLogger for the grading package.
// It avoids a grading → admin import cycle by implementing the interface in
// the admin package itself.
type GradingAuditAdapter struct {
	Repo Repository
}

// InsertAuditLog converts the grading entry into the admin audit params and
// delegates to the admin repository.
func (a *GradingAuditAdapter) InsertAuditLog(ctx context.Context, tx pgx.Tx, p grading.AuditLogEntry) error {
	return a.Repo.InsertAuditLog(ctx, tx, AuditLogParams{
		OrganizationID: p.OrganizationID,
		ActorUserID:    p.ActorUserID,
		Action:         p.Action,
		ResourceType:   p.ResourceType,
		ResourceID:     p.ResourceID,
		BeforeJSON:     p.BeforeJSON,
		AfterJSON:      p.AfterJSON,
		MetadataJSON:   p.MetadataJSON,
	})
}

// Ensure the adapter satisfies the grading AuditLogger interface.
var _ grading.AuditLogger = (*GradingAuditAdapter)(nil)
