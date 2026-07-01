package resources

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
)

type fakeRepo struct {
	resources       []Resource
	files           map[string][]ResourceFile
	createResource  func(ctx context.Context, actor auth.Actor, req CreateResourceRequest) (Resource, error)
	getResource     func(ctx context.Context, orgID, id string) (Resource, error)
	getActiveFile   func(ctx context.Context, orgID, resourceID string) (ResourceFile, error)
	createFile      func(ctx context.Context, file ResourceFile) (ResourceFile, error)
	updateStatus    func(ctx context.Context, orgID, id, status string) (Resource, error)
	archiveResource func(ctx context.Context, orgID, id string) error
	archiveFile     func(ctx context.Context, orgID, fileID string) error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{files: map[string][]ResourceFile{}}
}

func (f *fakeRepo) CreateResource(ctx context.Context, actor auth.Actor, req CreateResourceRequest) (Resource, error) {
	if f.createResource != nil {
		return f.createResource(ctx, actor, req)
	}
	r := Resource{ID: "res-1", OrgID: actor.OrgID, Title: req.Title, Description: req.Description, ContextType: req.ContextType, ContextID: req.ContextID, Status: StatusDraft, CreatedBy: actor.UserID}
	f.resources = append(f.resources, r)
	return r, nil
}

func (f *fakeRepo) ListResources(ctx context.Context, orgID string, statuses []string) ([]Resource, error) {
	out := []Resource{}
	for _, r := range f.resources {
		if r.OrgID != orgID {
			continue
		}
		for _, s := range statuses {
			if r.Status == s {
				out = append(out, r)
				return out, nil
			}
		}
	}
	return out, nil
}

func (f *fakeRepo) GetResource(ctx context.Context, orgID, id string) (Resource, error) {
	if f.getResource != nil {
		return f.getResource(ctx, orgID, id)
	}
	for _, r := range f.resources {
		if r.OrgID == orgID && r.ID == id {
			return r, nil
		}
	}
	return Resource{}, ErrNotFound
}

func (f *fakeRepo) UpdateResourceStatus(ctx context.Context, orgID, id, status string) (Resource, error) {
	if f.updateStatus != nil {
		return f.updateStatus(ctx, orgID, id, status)
	}
	for i, r := range f.resources {
		if r.OrgID == orgID && r.ID == id {
			r.Status = status
			f.resources[i] = r
			return r, nil
		}
	}
	return Resource{}, ErrNotFound
}

func (f *fakeRepo) ArchiveResource(ctx context.Context, orgID, id string) error {
	if f.archiveResource != nil {
		return f.archiveResource(ctx, orgID, id)
	}
	for i, r := range f.resources {
		if r.OrgID == orgID && r.ID == id {
			r.Status = StatusArchived
			f.resources[i] = r
			return nil
		}
	}
	return ErrNotFound
}

func (f *fakeRepo) CreateResourceFile(ctx context.Context, file ResourceFile) (ResourceFile, error) {
	if f.createFile != nil {
		return f.createFile(ctx, file)
	}
	file.Status = FileStatusActive
	if file.ID == "" {
		file.ID = "file-1"
	}
	f.files[file.ResourceID] = append(f.files[file.ResourceID], file)
	return file, nil
}

func (f *fakeRepo) ListResourceFiles(ctx context.Context, orgID, resourceID string) ([]ResourceFile, error) {
	return f.files[resourceID], nil
}

func (f *fakeRepo) GetActiveResourceFile(ctx context.Context, orgID, resourceID string) (ResourceFile, error) {
	if f.getActiveFile != nil {
		return f.getActiveFile(ctx, orgID, resourceID)
	}
	for _, f := range f.files[resourceID] {
		if f.Status == FileStatusActive {
			return f, nil
		}
	}
	return ResourceFile{}, ErrNoActiveFile
}

func (f *fakeRepo) ArchiveResourceFile(ctx context.Context, orgID, fileID string) error {
	if f.archiveFile != nil {
		return f.archiveFile(ctx, orgID, fileID)
	}
	for rid, list := range f.files {
		for i, item := range list {
			if item.ID == fileID {
				item.Status = FileStatusArchived
				f.files[rid][i] = item
				return nil
			}
		}
	}
	return nil
}

type fakeStorage struct {
	storeKey string
	stored   []byte
}

func (s *fakeStorage) Store(ctx context.Context, r io.Reader, size int64, contentType string) (string, error) {
	buf := make([]byte, size)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	s.stored = buf
	if s.storeKey == "" {
		s.storeKey = "key-1"
	}
	return s.storeKey, nil
}

func (s *fakeStorage) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.stored)), nil
}

func (s *fakeStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func newActor(role string) auth.Actor {
	return auth.Actor{UserID: "user-1", OrgID: "org-1", Roles: []string{role}}
}

func TestService_CreateResource_RequiresManager(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{}
	svc := NewService(repo, store, 1024)

	if _, err := svc.CreateResource(context.Background(), newActor("student"), CreateResourceRequest{Title: "x", ContextType: ContextTypeOrganization, ContextID: "org-1"}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	r, err := svc.CreateResource(context.Background(), newActor("teacher"), CreateResourceRequest{Title: "Lesson 1", ContextType: ContextTypeOrganization, ContextID: "org-1"})
	if err != nil {
		t.Fatalf("teacher create: %v", err)
	}
	if r.Title != "Lesson 1" {
		t.Fatalf("unexpected title: %q", r.Title)
	}
}

func TestService_PublishAndArchive_RequiresManager(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{}
	svc := NewService(repo, store, 1024)

	if _, err := svc.PublishResource(context.Background(), newActor("student"), "x"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	if err := svc.ArchiveResource(context.Background(), newActor("student"), "x"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestService_UploadFile_StoresPayload(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{storeKey: "abc123"}
	svc := NewService(repo, store, 1024)
	teacher := newActor("teacher")

	created, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{Title: "Doc", ContextType: ContextTypeOrganization, ContextID: "org-1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	payload := []byte("hello")
	file, err := svc.UploadFile(context.Background(), teacher, created.ID, "doc.txt", "text/plain", bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if file.StorageKey != "abc123" {
		t.Fatalf("unexpected storage key: %q", file.StorageKey)
	}
	if !bytes.Equal(store.stored, payload) {
		t.Fatalf("payload not stored: %q", store.stored)
	}
}

func TestService_DownloadFile_RequiresPublishedForStudent(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{storeKey: "k"}
	svc := NewService(repo, store, 1024)
	teacher := newActor("teacher")
	student := newActor("student")

	created, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{Title: "Doc", ContextType: ContextTypeOrganization, ContextID: "org-1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = svc.UploadFile(context.Background(), teacher, created.ID, "doc.txt", "text/plain", bytes.NewReader([]byte("x")), 1)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if _, _, err := svc.DownloadFile(context.Background(), student, created.ID); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized draft download for student, got %v", err)
	}

	published, err := svc.PublishResource(context.Background(), teacher, created.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.Status != StatusPublished {
		t.Fatalf("unexpected status: %q", published.Status)
	}

	r, file, err := svc.DownloadFile(context.Background(), student, created.ID)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	r.Close()
	if file.SizeBytes != 1 {
		t.Fatalf("unexpected size: %d", file.SizeBytes)
	}
}

func TestService_UploadFile_SanitizesDisallowedContentType(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{storeKey: "k"}
	svc := NewService(repo, store, 1024)
	teacher := newActor("teacher")
	created, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{Title: "Doc", ContextType: ContextTypeOrganization, ContextID: "org-1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Attacker tries to set an exotic content type. Service should
	// sanitize to application/octet-stream before persisting.
	uploaded, err := svc.UploadFile(context.Background(), teacher, created.ID, "doc.bin", "application/x-evil", bytes.NewReader([]byte("x")), 1)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if uploaded.ContentType != "application/octet-stream" {
		t.Fatalf("expected sanitized content type, got %q", uploaded.ContentType)
	}
	// And on the way out, file.ContentType is also sanitized.
	_, file, err := svc.DownloadFile(context.Background(), teacher, created.ID)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if file.ContentType != "application/octet-stream" {
		t.Fatalf("expected sanitized content type on read-back, got %q", file.ContentType)
	}
}
