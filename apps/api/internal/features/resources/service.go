package resources

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/storage"
)

// Service defines resource business operations.
type Service interface {
	ListResources(ctx context.Context, actor auth.Actor) ([]Resource, error)
	CreateResource(ctx context.Context, actor auth.Actor, req CreateResourceRequest) (Resource, error)
	PublishResource(ctx context.Context, actor auth.Actor, id string) (Resource, error)
	ArchiveResource(ctx context.Context, actor auth.Actor, id string) error
	UploadFile(ctx context.Context, actor auth.Actor, resourceID, fileName, contentType string, data io.Reader, size int64) (ResourceFile, error)
	DownloadFile(ctx context.Context, actor auth.Actor, resourceID string) (io.ReadCloser, ResourceFile, error)
}

type service struct {
	repo    Repository
	storage storage.Provider
	maxSize int64
}

// NewService creates a resources service.
func NewService(repo Repository, provider storage.Provider, maxSize int64) Service {
	return &service{repo: repo, storage: provider, maxSize: maxSize}
}

func (s *service) ListResources(ctx context.Context, actor auth.Actor) ([]Resource, error) {
	if actor.OrgID == "" {
		return nil, ErrUnauthorized
	}
	var statuses []string
	if hasRole(actor.Roles, "teacher") || hasRole(actor.Roles, "admin") {
		statuses = []string{StatusDraft, StatusPublished}
	} else {
		statuses = []string{StatusPublished}
	}
	return s.repo.ListResources(ctx, actor.OrgID, statuses)
}

func (s *service) CreateResource(ctx context.Context, actor auth.Actor, req CreateResourceRequest) (Resource, error) {
	if !isManager(actor) {
		return Resource{}, ErrUnauthorized
	}
	if req.Title == "" {
		return Resource{}, ErrInvalidInput
	}
	if req.ContextType != ContextTypeOrganization && req.ContextType != ContextTypeClass {
		return Resource{}, ErrInvalidInput
	}
	if req.ContextID == "" {
		return Resource{}, ErrInvalidInput
	}
	return s.repo.CreateResource(ctx, actor, req)
}

func (s *service) PublishResource(ctx context.Context, actor auth.Actor, id string) (Resource, error) {
	if !isManager(actor) {
		return Resource{}, ErrUnauthorized
	}
	return s.repo.UpdateResourceStatus(ctx, actor.OrgID, id, StatusPublished)
}

func (s *service) ArchiveResource(ctx context.Context, actor auth.Actor, id string) error {
	if !isManager(actor) {
		return ErrUnauthorized
	}
	return s.repo.ArchiveResource(ctx, actor.OrgID, id)
}

func (s *service) UploadFile(ctx context.Context, actor auth.Actor, resourceID, fileName, contentType string, data io.Reader, size int64) (ResourceFile, error) {
	if !isManager(actor) {
		return ResourceFile{}, ErrUnauthorized
	}
	if size > s.maxSize {
		return ResourceFile{}, ErrFileTooLarge
	}
	if fileName == "" {
		return ResourceFile{}, ErrInvalidInput
	}
	// Sanitize the uploaded content type against the allowlist. The
	// download path re-validates on the way out, so anything outside
	// the list is replaced with application/octet-stream and stays
	// safe across all storage backends.
	contentType = storage.SanitizeContentType(contentType)

	resource, err := s.repo.GetResource(ctx, actor.OrgID, resourceID)
	if err != nil {
		return ResourceFile{}, err
	}
	if resource.Status == StatusArchived {
		return ResourceFile{}, ErrInvalidStatus
	}

	// Archive any existing active file before attaching a new one.
	if active, err := s.repo.GetActiveResourceFile(ctx, actor.OrgID, resourceID); err == nil {
		_ = s.repo.ArchiveResourceFile(ctx, actor.OrgID, active.ID)
	}

	key, err := s.storage.Store(ctx, data, size, contentType)
	if err != nil {
		return ResourceFile{}, fmt.Errorf("%w: %v", ErrStorageFailure, err)
	}

	file := ResourceFile{
		ResourceID:   resourceID,
		OrgID:        actor.OrgID,
		OriginalName: path.Base(fileName),
		StorageKey:   key,
		ContentType:  contentType,
		SizeBytes:    size,
		CreatedBy:    actor.UserID,
	}
	return s.repo.CreateResourceFile(ctx, file)
}

func (s *service) DownloadFile(ctx context.Context, actor auth.Actor, resourceID string) (io.ReadCloser, ResourceFile, error) {
	if actor.OrgID == "" {
		return nil, ResourceFile{}, ErrUnauthorized
	}

	resource, err := s.repo.GetResource(ctx, actor.OrgID, resourceID)
	if err != nil {
		return nil, ResourceFile{}, err
	}

	if resource.Status == StatusArchived {
		return nil, ResourceFile{}, ErrNotFound
	}
	if resource.Status != StatusPublished && !isManager(actor) {
		return nil, ResourceFile{}, ErrUnauthorized
	}

	file, err := s.repo.GetActiveResourceFile(ctx, actor.OrgID, resourceID)
	if err != nil {
		return nil, ResourceFile{}, ErrNoActiveFile
	}

	reader, err := s.storage.Retrieve(ctx, file.StorageKey)
	if err != nil {
		return nil, ResourceFile{}, fmt.Errorf("%w: %v", ErrStorageFailure, err)
	}
	return reader, file, nil
}

func isManager(actor auth.Actor) bool {
	return hasRole(actor.Roles, "teacher") || hasRole(actor.Roles, "admin")
}

func hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
