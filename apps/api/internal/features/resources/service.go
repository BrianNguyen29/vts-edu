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
	ListResources(ctx context.Context, actor auth.Actor, filter ListFilter) ([]Resource, error)
	CreateResource(ctx context.Context, actor auth.Actor, req CreateResourceRequest) (Resource, error)
	PublishResource(ctx context.Context, actor auth.Actor, id string) (Resource, error)
	ArchiveResource(ctx context.Context, actor auth.Actor, id string) error
	UploadFile(ctx context.Context, actor auth.Actor, resourceID, fileName, contentType string, data io.Reader, size int64) (ResourceFile, error)
	UploadFiles(ctx context.Context, actor auth.Actor, resourceID string, files []UploadInput) ([]ResourceFile, error)
	ListFiles(ctx context.Context, actor auth.Actor, resourceID string) ([]ResourceFile, error)
	DownloadFile(ctx context.Context, actor auth.Actor, resourceID string, fileID string) (io.ReadCloser, ResourceFile, error)
}

// ListFilter narrows which resources are returned to a caller.
type ListFilter struct {
	ContextType string
	ContextID   string
}

// UploadInput is one multipart file entry. The service stores each
// independently and returns a list of persisted files (partial success is
// acceptable: a failure on a later entry does not roll back earlier ones).
type UploadInput struct {
	FileName    string
	ContentType string
	Data        io.Reader
	Size        int64
}

type service struct {
	repo    Repository
	storage storage.Provider
	access  ClassAccessChecker
	maxSize int64
}

// NewService creates a resources service.
func NewService(repo Repository, provider storage.Provider, maxSize int64) Service {
	return NewServiceWithAccess(repo, provider, maxSize, stubChecker{})
}

// NewServiceWithAccess creates a resources service with a class access
// checker. When access is nil the stub is used and class-scoped checks
// always deny (the caller must be an org-scoped admin in that case).
func NewServiceWithAccess(repo Repository, provider storage.Provider, maxSize int64, access ClassAccessChecker) Service {
	if access == nil {
		access = stubChecker{}
	}
	return &service{repo: repo, storage: provider, access: access, maxSize: maxSize}
}

func (s *service) ListResources(ctx context.Context, actor auth.Actor, filter ListFilter) ([]Resource, error) {
	if actor.OrgID == "" {
		return nil, ErrUnauthorized
	}
	var statuses []string
	if hasRole(actor.Roles, "teacher") || hasRole(actor.Roles, "admin") {
		statuses = []string{StatusDraft, StatusPublished}
	} else {
		statuses = []string{StatusPublished}
	}

	// Org-scoped fast path: nothing to narrow by class enrolment.
	if filter.ContextType == "" || filter.ContextType == ContextTypeOrganization {
		return s.repo.ListResources(ctx, actor.OrgID, statuses)
	}

	if filter.ContextType != ContextTypeClass {
		return nil, ErrInvalidInput
	}
	if filter.ContextID == "" {
		return nil, ErrInvalidInput
	}
	exists, err := s.access.ClassExists(ctx, actor.OrgID, filter.ContextID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errClassAccessUnavailable, err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	all, err := s.repo.ListResources(ctx, actor.OrgID, statuses)
	if err != nil {
		return nil, err
	}

	out := make([]Resource, 0, len(all))
	for _, r := range all {
		if r.ContextType != ContextTypeClass || r.ContextID != filter.ContextID {
			continue
		}
		if hasRole(actor.Roles, "admin") || hasRole(actor.Roles, "teacher") {
			out = append(out, r)
			continue
		}
		// Student: include only if enrolled in this class.
		ok, err := s.access.CanViewClass(ctx, actor, r.ContextID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errClassAccessUnavailable, err)
		}
		if ok {
			out = append(out, r)
		}
	}
	return out, nil
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
	if req.ContextType == ContextTypeClass {
		ok, err := s.access.CanManageClass(ctx, actor, req.ContextID)
		if err != nil {
			return Resource{}, fmt.Errorf("%w: %v", errClassAccessUnavailable, err)
		}
		if !ok {
			return Resource{}, ErrUnauthorized
		}
	}
	return s.repo.CreateResource(ctx, actor, req)
}

func (s *service) PublishResource(ctx context.Context, actor auth.Actor, id string) (Resource, error) {
	if !isManager(actor) {
		return Resource{}, ErrUnauthorized
	}
	resource, err := s.repo.GetResource(ctx, actor.OrgID, id)
	if err != nil {
		return Resource{}, err
	}
	if resource.ContextType == ContextTypeClass {
		ok, err := s.access.CanManageClass(ctx, actor, resource.ContextID)
		if err != nil {
			return Resource{}, fmt.Errorf("%w: %v", errClassAccessUnavailable, err)
		}
		if !ok {
			return Resource{}, ErrUnauthorized
		}
	}
	return s.repo.UpdateResourceStatus(ctx, actor.OrgID, id, StatusPublished)
}

func (s *service) ArchiveResource(ctx context.Context, actor auth.Actor, id string) error {
	if !isManager(actor) {
		return ErrUnauthorized
	}
	resource, err := s.repo.GetResource(ctx, actor.OrgID, id)
	if err != nil {
		return err
	}
	if resource.ContextType == ContextTypeClass {
		ok, err := s.access.CanManageClass(ctx, actor, resource.ContextID)
		if err != nil {
			return fmt.Errorf("%w: %v", errClassAccessUnavailable, err)
		}
		if !ok {
			return ErrUnauthorized
		}
	}
	return s.repo.ArchiveResource(ctx, actor.OrgID, id)
}

func (s *service) UploadFile(ctx context.Context, actor auth.Actor, resourceID, fileName, contentType string, data io.Reader, size int64) (ResourceFile, error) {
	files, err := s.UploadFiles(ctx, actor, resourceID, []UploadInput{{
		FileName:    fileName,
		ContentType: contentType,
		Data:        data,
		Size:        size,
	}})
	if err != nil {
		return ResourceFile{}, err
	}
	if len(files) == 0 {
		return ResourceFile{}, ErrStorageFailure
	}
	return files[0], nil
}

// UploadFiles stores all entries as ACTIVE. Existing ACTIVE files are kept
// (multi-file resources): the caller can archive the older file via
// ArchiveResourceFile or simply leave it as historical context. Partial
// success is acceptable; per-file failures are returned in
// ErrPartialUpload and the call site decides whether to surface them.
func (s *service) UploadFiles(ctx context.Context, actor auth.Actor, resourceID string, inputs []UploadInput) ([]ResourceFile, error) {
	if !isManager(actor) {
		return nil, ErrUnauthorized
	}
	if len(inputs) == 0 {
		return nil, ErrInvalidInput
	}
	if size := totalSize(inputs); size > s.maxSize {
		return nil, ErrFileTooLarge
	}

	resource, err := s.repo.GetResource(ctx, actor.OrgID, resourceID)
	if err != nil {
		return nil, err
	}
	if resource.Status == StatusArchived {
		return nil, ErrInvalidStatus
	}
	if resource.ContextType == ContextTypeClass {
		ok, err := s.access.CanManageClass(ctx, actor, resource.ContextID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errClassAccessUnavailable, err)
		}
		if !ok {
			return nil, ErrUnauthorized
		}
	}

	stored := make([]ResourceFile, 0, len(inputs))
	var firstErr error
	for _, in := range inputs {
		if in.FileName == "" {
			if firstErr == nil {
				firstErr = ErrInvalidInput
			}
			continue
		}
		if in.Size > s.maxSize {
			if firstErr == nil {
				firstErr = ErrFileTooLarge
			}
			continue
		}
		ct := storage.SanitizeContentType(in.ContentType)
		key, err := s.storage.Store(ctx, in.Data, in.Size, ct)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%w: %v", ErrStorageFailure, err)
			}
			continue
		}
		file := ResourceFile{
			ResourceID:   resourceID,
			OrgID:        actor.OrgID,
			OriginalName: path.Base(in.FileName),
			StorageKey:   key,
			ContentType:  ct,
			SizeBytes:    in.Size,
			CreatedBy:    actor.UserID,
		}
		persisted, err := s.repo.CreateResourceFile(ctx, file)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		stored = append(stored, persisted)
	}
	if len(stored) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return stored, nil
}

func (s *service) ListFiles(ctx context.Context, actor auth.Actor, resourceID string) ([]ResourceFile, error) {
	if actor.OrgID == "" {
		return nil, ErrUnauthorized
	}
	resource, err := s.repo.GetResource(ctx, actor.OrgID, resourceID)
	if err != nil {
		return nil, err
	}
	if resource.ContextType == ContextTypeClass {
		ok, err := s.access.CanViewClass(ctx, actor, resource.ContextID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errClassAccessUnavailable, err)
		}
		if !ok {
			return nil, ErrUnauthorized
		}
	}
	if resource.Status != StatusPublished && !isManager(actor) {
		return nil, ErrUnauthorized
	}
	files, err := s.repo.ListResourceFiles(ctx, actor.OrgID, resourceID)
	if err != nil {
		return nil, err
	}
	out := make([]ResourceFile, 0, len(files))
	for _, f := range files {
		if f.Status == FileStatusActive {
			out = append(out, f)
		}
	}
	return out, nil
}

func (s *service) DownloadFile(ctx context.Context, actor auth.Actor, resourceID string, fileID string) (io.ReadCloser, ResourceFile, error) {
	if actor.OrgID == "" {
		return nil, ResourceFile{}, ErrUnauthorized
	}

	resource, err := s.repo.GetResource(ctx, actor.OrgID, resourceID)
	if err != nil {
		return nil, ResourceFile{}, err
	}
	if resource.ContextType == ContextTypeClass {
		ok, err := s.access.CanViewClass(ctx, actor, resource.ContextID)
		if err != nil {
			return nil, ResourceFile{}, fmt.Errorf("%w: %v", errClassAccessUnavailable, err)
		}
		if !ok {
			return nil, ResourceFile{}, ErrUnauthorized
		}
	}

	if resource.Status == StatusArchived {
		return nil, ResourceFile{}, ErrNotFound
	}
	if resource.Status != StatusPublished && !isManager(actor) {
		return nil, ResourceFile{}, ErrUnauthorized
	}

	var file ResourceFile
	if fileID != "" {
		file, err = s.repo.GetResourceFile(ctx, actor.OrgID, fileID)
		if err != nil {
			return nil, ResourceFile{}, err
		}
		if file.ResourceID != resourceID {
			return nil, ResourceFile{}, ErrNotFound
		}
		if file.Status != FileStatusActive {
			return nil, ResourceFile{}, ErrNoActiveFile
		}
	} else {
		file, err = s.repo.GetActiveResourceFile(ctx, actor.OrgID, resourceID)
		if err != nil {
			return nil, ResourceFile{}, err
		}
	}

	reader, err := s.storage.Retrieve(ctx, file.StorageKey)
	if err != nil {
		return nil, ResourceFile{}, fmt.Errorf("%w: %v", ErrStorageFailure, err)
	}
	return reader, file, nil
}

func totalSize(inputs []UploadInput) int64 {
	var sum int64
	for _, in := range inputs {
		sum += in.Size
	}
	return sum
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
