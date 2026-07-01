package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	resourcessqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/resources/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for resources and their files.
type Repository interface {
	CreateResource(ctx context.Context, actor auth.Actor, req CreateResourceRequest) (Resource, error)
	ListResources(ctx context.Context, orgID string, statuses []string) ([]Resource, error)
	GetResource(ctx context.Context, orgID, id string) (Resource, error)
	UpdateResourceStatus(ctx context.Context, orgID, id, status string) (Resource, error)
	ArchiveResource(ctx context.Context, orgID, id string) error

	CreateResourceFile(ctx context.Context, file ResourceFile) (ResourceFile, error)
	ListResourceFiles(ctx context.Context, orgID, resourceID string) ([]ResourceFile, error)
	GetActiveResourceFile(ctx context.Context, orgID, resourceID string) (ResourceFile, error)
	ArchiveResourceFile(ctx context.Context, orgID, fileID string) error
}

type sqlcRepository struct {
	queries *resourcessqlc.Queries
}

// NewRepository creates a new resources repository backed by generated sqlc queries.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: resourcessqlc.New(pool)}
}

func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func toText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func textOrEmpty(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func timeOrNil(t pgtype.Timestamptz) *string {
	if t.Valid {
		s := t.Time.UTC().Format(time.RFC3339Nano)
		return &s
	}
	return nil
}

func formatTime(t pgtype.Timestamptz) string {
	if t.Valid {
		return t.Time.UTC().Format(time.RFC3339Nano)
	}
	return ""
}

func (r *sqlcRepository) CreateResource(ctx context.Context, actor auth.Actor, req CreateResourceRequest) (Resource, error) {
	orgUUID, err := toUUID(actor.OrgID)
	if err != nil {
		return Resource{}, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(actor.UserID)
	if err != nil {
		return Resource{}, fmt.Errorf("invalid user id: %w", err)
	}
	contextUUID, err := toUUID(req.ContextID)
	if err != nil {
		return Resource{}, fmt.Errorf("invalid context id: %w", err)
	}

	row, err := r.queries.CreateResource(ctx, resourcessqlc.CreateResourceParams{
		OrganizationID: orgUUID,
		Title:          req.Title,
		Description:    toText(req.Description),
		Column4:        resourcessqlc.ResourceContextType(req.ContextType),
		ContextID:      contextUUID,
		CreatedBy:      userUUID,
	})
	if err != nil {
		return Resource{}, fmt.Errorf("create resource: %w", err)
	}
	return mapResource(row), nil
}

func (r *sqlcRepository) ListResources(ctx context.Context, orgID string, statuses []string) ([]Resource, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	rows, err := r.queries.ListResources(ctx, resourcessqlc.ListResourcesParams{
		OrganizationID: orgUUID,
		Column2:        statuses,
	})
	if err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}

	resources := make([]Resource, len(rows))
	for i, row := range rows {
		resources[i] = mapResource(row)
	}
	return resources, nil
}

func (r *sqlcRepository) GetResource(ctx context.Context, orgID, id string) (Resource, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Resource{}, fmt.Errorf("invalid organization id: %w", err)
	}
	idUUID, err := toUUID(id)
	if err != nil {
		return Resource{}, fmt.Errorf("invalid resource id: %w", err)
	}

	row, err := r.queries.GetResource(ctx, resourcessqlc.GetResourceParams{
		ID:             idUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return Resource{}, mapRepoError(err)
	}
	return mapResource(row), nil
}

func (r *sqlcRepository) UpdateResourceStatus(ctx context.Context, orgID, id, status string) (Resource, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Resource{}, fmt.Errorf("invalid organization id: %w", err)
	}
	idUUID, err := toUUID(id)
	if err != nil {
		return Resource{}, fmt.Errorf("invalid resource id: %w", err)
	}

	row, err := r.queries.UpdateResourceStatus(ctx, resourcessqlc.UpdateResourceStatusParams{
		ID:             idUUID,
		OrganizationID: orgUUID,
		Column3:        resourcessqlc.ResourceStatus(status),
	})
	if err != nil {
		return Resource{}, mapRepoError(err)
	}
	return mapResource(row), nil
}

func (r *sqlcRepository) ArchiveResource(ctx context.Context, orgID, id string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	idUUID, err := toUUID(id)
	if err != nil {
		return fmt.Errorf("invalid resource id: %w", err)
	}

	if err := r.queries.ArchiveResource(ctx, resourcessqlc.ArchiveResourceParams{
		ID:             idUUID,
		OrganizationID: orgUUID,
	}); err != nil {
		return mapRepoError(err)
	}
	return nil
}

func (r *sqlcRepository) CreateResourceFile(ctx context.Context, file ResourceFile) (ResourceFile, error) {
	resourceUUID, err := toUUID(file.ResourceID)
	if err != nil {
		return ResourceFile{}, fmt.Errorf("invalid resource id: %w", err)
	}
	orgUUID, err := toUUID(file.OrgID)
	if err != nil {
		return ResourceFile{}, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(file.CreatedBy)
	if err != nil {
		return ResourceFile{}, fmt.Errorf("invalid user id: %w", err)
	}

	row, err := r.queries.CreateResourceFile(ctx, resourcessqlc.CreateResourceFileParams{
		ResourceID:     resourceUUID,
		OrganizationID: orgUUID,
		OriginalName:   file.OriginalName,
		StorageKey:     file.StorageKey,
		ContentType:    file.ContentType,
		SizeBytes:      file.SizeBytes,
		CreatedBy:      userUUID,
	})
	if err != nil {
		return ResourceFile{}, fmt.Errorf("create resource file: %w", err)
	}
	return mapFile(row), nil
}

func (r *sqlcRepository) ListResourceFiles(ctx context.Context, orgID, resourceID string) ([]ResourceFile, error) {
	resourceUUID, err := toUUID(resourceID)
	if err != nil {
		return nil, fmt.Errorf("invalid resource id: %w", err)
	}
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	rows, err := r.queries.ListResourceFiles(ctx, resourcessqlc.ListResourceFilesParams{
		ResourceID:     resourceUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list resource files: %w", err)
	}

	files := make([]ResourceFile, len(rows))
	for i, row := range rows {
		files[i] = mapFile(row)
	}
	return files, nil
}

func (r *sqlcRepository) GetActiveResourceFile(ctx context.Context, orgID, resourceID string) (ResourceFile, error) {
	resourceUUID, err := toUUID(resourceID)
	if err != nil {
		return ResourceFile{}, fmt.Errorf("invalid resource id: %w", err)
	}
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return ResourceFile{}, fmt.Errorf("invalid organization id: %w", err)
	}

	row, err := r.queries.GetActiveResourceFile(ctx, resourcessqlc.GetActiveResourceFileParams{
		ResourceID:     resourceUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return ResourceFile{}, mapRepoError(err)
	}
	return mapFile(row), nil
}

func (r *sqlcRepository) ArchiveResourceFile(ctx context.Context, orgID, fileID string) error {
	fileUUID, err := toUUID(fileID)
	if err != nil {
		return fmt.Errorf("invalid file id: %w", err)
	}
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}

	if err := r.queries.ArchiveResourceFile(ctx, resourcessqlc.ArchiveResourceFileParams{
		ID:             fileUUID,
		OrganizationID: orgUUID,
	}); err != nil {
		return mapRepoError(err)
	}
	return nil
}

func mapResource(row any) Resource {
	switch r := row.(type) {
	case resourcessqlc.CreateResourceRow:
		return Resource{
			ID:          r.ID.String(),
			OrgID:       r.OrganizationID.String(),
			Title:       r.Title,
			Description: textOrEmpty(r.Description),
			ContextType: r.ContextType,
			ContextID:   r.ContextID.String(),
			Status:      r.Status,
			CreatedBy:   r.CreatedBy.String(),
			CreatedAt:   formatTime(r.CreatedAt),
			UpdatedAt:   formatTime(r.UpdatedAt),
			PublishedAt: timeOrNil(r.PublishedAt),
		}
	case resourcessqlc.GetResourceRow:
		return Resource{
			ID:          r.ID.String(),
			OrgID:       r.OrganizationID.String(),
			Title:       r.Title,
			Description: textOrEmpty(r.Description),
			ContextType: r.ContextType,
			ContextID:   r.ContextID.String(),
			Status:      r.Status,
			CreatedBy:   r.CreatedBy.String(),
			CreatedAt:   formatTime(r.CreatedAt),
			UpdatedAt:   formatTime(r.UpdatedAt),
			PublishedAt: timeOrNil(r.PublishedAt),
		}
	case resourcessqlc.UpdateResourceStatusRow:
		return Resource{
			ID:          r.ID.String(),
			OrgID:       r.OrganizationID.String(),
			Title:       r.Title,
			Description: textOrEmpty(r.Description),
			ContextType: r.ContextType,
			ContextID:   r.ContextID.String(),
			Status:      r.Status,
			CreatedBy:   r.CreatedBy.String(),
			CreatedAt:   formatTime(r.CreatedAt),
			UpdatedAt:   formatTime(r.UpdatedAt),
			PublishedAt: timeOrNil(r.PublishedAt),
		}
	case resourcessqlc.ListResourcesRow:
		return Resource{
			ID:          r.ID.String(),
			OrgID:       r.OrganizationID.String(),
			Title:       r.Title,
			Description: textOrEmpty(r.Description),
			ContextType: r.ContextType,
			ContextID:   r.ContextID.String(),
			Status:      r.Status,
			CreatedBy:   r.CreatedBy.String(),
			CreatedAt:   formatTime(r.CreatedAt),
			UpdatedAt:   formatTime(r.UpdatedAt),
			PublishedAt: timeOrNil(r.PublishedAt),
		}
	default:
		return Resource{}
	}
}

func mapFile(row any) ResourceFile {
	switch f := row.(type) {
	case resourcessqlc.CreateResourceFileRow:
		return ResourceFile{
			ID:           f.ID.String(),
			ResourceID:   f.ResourceID.String(),
			OrgID:        f.OrganizationID.String(),
			OriginalName: f.OriginalName,
			StorageKey:   f.StorageKey,
			ContentType:  f.ContentType,
			SizeBytes:    f.SizeBytes,
			Status:       f.Status,
			CreatedBy:    f.CreatedBy.String(),
			CreatedAt:    formatTime(f.CreatedAt),
		}
	case resourcessqlc.GetActiveResourceFileRow:
		return ResourceFile{
			ID:           f.ID.String(),
			ResourceID:   f.ResourceID.String(),
			OrgID:        f.OrganizationID.String(),
			OriginalName: f.OriginalName,
			StorageKey:   f.StorageKey,
			ContentType:  f.ContentType,
			SizeBytes:    f.SizeBytes,
			Status:       f.Status,
			CreatedBy:    f.CreatedBy.String(),
			CreatedAt:    formatTime(f.CreatedAt),
		}
	default:
		return ResourceFile{}
	}
}
