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
	resourceCounter int
	fileCounter     int
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
	f.resourceCounter++
	r := Resource{ID: "res-" + strconvItoa(f.resourceCounter), OrgID: actor.OrgID, Title: req.Title, Description: req.Description, ContextType: req.ContextType, ContextID: req.ContextID, Status: StatusDraft, CreatedBy: actor.UserID}
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
				break
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
		f.fileCounter++
		file.ID = "file-" + strconvItoa(f.fileCounter)
	}
	f.files[file.ResourceID] = append(f.files[file.ResourceID], file)
	return file, nil
}

func strconvItoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := "0123456789"
	out := ""
	for n > 0 {
		out = string(digits[n%10]) + out
		n /= 10
	}
	return out
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

func (f *fakeRepo) GetResourceFile(ctx context.Context, orgID, fileID string) (ResourceFile, error) {
	for _, list := range f.files {
		for _, item := range list {
			if item.ID == fileID && item.OrgID == orgID {
				return item, nil
			}
		}
	}
	return ResourceFile{}, ErrNotFound
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

// fakeClassAccess is a test double for ClassAccessChecker.
type fakeClassAccess struct {
	exists    map[string]bool
	canView   map[string]bool
	canManage map[string]bool
}

func (f *fakeClassAccess) ClassExists(ctx context.Context, orgID, classID string) (bool, error) {
	return f.exists[classID], nil
}
func (f *fakeClassAccess) CanViewClass(ctx context.Context, actor auth.Actor, classID string) (bool, error) {
	return f.canView[classID], nil
}
func (f *fakeClassAccess) CanManageClass(ctx context.Context, actor auth.Actor, classID string) (bool, error) {
	return f.canManage[classID], nil
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

	if _, _, err := svc.DownloadFile(context.Background(), student, created.ID, ""); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized draft download for student, got %v", err)
	}

	published, err := svc.PublishResource(context.Background(), teacher, created.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.Status != StatusPublished {
		t.Fatalf("unexpected status: %q", published.Status)
	}

	r, file, err := svc.DownloadFile(context.Background(), student, created.ID, "")
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
	_, file, err := svc.DownloadFile(context.Background(), teacher, created.ID, "")
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if file.ContentType != "application/octet-stream" {
		t.Fatalf("expected sanitized content type on read-back, got %q", file.ContentType)
	}
}

func TestService_UploadFiles_AllowsMultipleActive(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{storeKey: "k"}
	svc := NewService(repo, store, 1024)
	teacher := newActor("teacher")
	created, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{Title: "Doc", ContextType: ContextTypeOrganization, ContextID: "org-1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	first, err := svc.UploadFile(context.Background(), teacher, created.ID, "a.txt", "text/plain", bytes.NewReader([]byte("aaa")), 3)
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}
	second, err := svc.UploadFile(context.Background(), teacher, created.ID, "b.txt", "text/plain", bytes.NewReader([]byte("bbb")), 3)
	if err != nil {
		t.Fatalf("second upload: %v", err)
	}

	files, err := svc.ListFiles(context.Background(), teacher, created.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 active files, got %d", len(files))
	}
	if first.ID == second.ID {
		t.Fatalf("expected distinct file ids")
	}
}

func TestService_DownloadFile_SpecificFileID(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{storeKey: "k"}
	svc := NewService(repo, store, 1024)
	teacher := newActor("teacher")
	created, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{Title: "Doc", ContextType: ContextTypeOrganization, ContextID: "org-1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	first, err := svc.UploadFile(context.Background(), teacher, created.ID, "a.txt", "text/plain", bytes.NewReader([]byte("aaa")), 3)
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}
	second, err := svc.UploadFile(context.Background(), teacher, created.ID, "b.txt", "text/plain", bytes.NewReader([]byte("bbb")), 3)
	if err != nil {
		t.Fatalf("second upload: %v", err)
	}
	if _, err := svc.PublishResource(context.Background(), teacher, created.ID); err != nil {
		t.Fatalf("publish: %v", err)
	}
	student := newActor("student")
	r, file, err := svc.DownloadFile(context.Background(), student, created.ID, second.ID)
	if err != nil {
		t.Fatalf("download specific: %v", err)
	}
	defer r.Close()
	if file.ID != second.ID {
		t.Fatalf("expected file %s, got %s", second.ID, file.ID)
	}
	_ = first
}

func TestService_CreateResource_ClassScope_RequiresManage(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{storeKey: "k"}
	access := &fakeClassAccess{canManage: map[string]bool{"class-1": false}}
	svc := NewServiceWithAccess(repo, store, 1024, access)
	teacher := newActor("teacher")

	_, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{
		Title: "Class Doc", ContextType: ContextTypeClass, ContextID: "class-1",
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized for non-managing teacher, got %v", err)
	}

	access.canManage["class-1"] = true
	r, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{
		Title: "Class Doc", ContextType: ContextTypeClass, ContextID: "class-1",
	})
	if err != nil {
		t.Fatalf("manage teacher: %v", err)
	}
	if r.ContextType != ContextTypeClass || r.ContextID != "class-1" {
		t.Fatalf("unexpected resource: %+v", r)
	}
}

func TestService_ListResources_FilterByClass(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{storeKey: "k"}
	access := &fakeClassAccess{
		exists:    map[string]bool{"class-1": true, "class-2": true},
		canView:   map[string]bool{"class-1": true, "class-2": false},
		canManage: map[string]bool{"class-1": true, "class-2": true},
	}
	svc := NewServiceWithAccess(repo, store, 1024, access)
	teacher := newActor("teacher")
	for _, ctx := range []struct {
		title, ctxType, ctxID string
	}{
		{"Org", ContextTypeOrganization, "org-1"},
		{"Class 1", ContextTypeClass, "class-1"},
		{"Class 2", ContextTypeClass, "class-2"},
	} {
		r, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{Title: ctx.title, ContextType: ctx.ctxType, ContextID: ctx.ctxID})
		if err != nil {
			t.Fatalf("create %s: %v", ctx.title, err)
		}
		if _, err := svc.PublishResource(context.Background(), teacher, r.ID); err != nil {
			t.Fatalf("publish %s: %v", ctx.title, err)
		}
	}

	student := newActor("student")
	got, err := svc.ListResources(context.Background(), student, ListFilter{ContextType: ContextTypeClass, ContextID: "class-1"})
	if err != nil {
		t.Fatalf("list class-1: %v", err)
	}
	if len(got) != 1 || got[0].ContextID != "class-1" {
		t.Fatalf("expected 1 class-1 resource, got %+v", got)
	}

	got, err = svc.ListResources(context.Background(), student, ListFilter{ContextType: ContextTypeClass, ContextID: "class-2"})
	if err != nil {
		t.Fatalf("list class-2: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 class-2 resources for non-enrolled student, got %+v", got)
	}

	admin := newActor("admin")
	got, err = svc.ListResources(context.Background(), admin, ListFilter{ContextType: ContextTypeClass, ContextID: "class-2"})
	if err != nil {
		t.Fatalf("list class-2 admin: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected admin to see class-2, got %+v", got)
	}
}

func TestService_DownloadFile_ClassScope_RequiresView(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStorage{storeKey: "k"}
	access := &fakeClassAccess{
		exists:  map[string]bool{"class-1": true},
		canView: map[string]bool{"class-1": false},
	}
	svc := NewServiceWithAccess(repo, store, 1024, access)
	teacher := newActor("teacher")
	access.canManage = map[string]bool{"class-1": true}
	created, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{Title: "Class Doc", ContextType: ContextTypeClass, ContextID: "class-1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.UploadFile(context.Background(), teacher, created.ID, "a.txt", "text/plain", bytes.NewReader([]byte("x")), 1); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if _, err := svc.PublishResource(context.Background(), teacher, created.ID); err != nil {
		t.Fatalf("publish: %v", err)
	}

	student := newActor("student")
	access.canView["class-1"] = false
	if _, _, err := svc.DownloadFile(context.Background(), student, created.ID, ""); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized for non-enrolled student, got %v", err)
	}
	access.canView["class-1"] = true
	if _, _, err := svc.DownloadFile(context.Background(), student, created.ID, ""); err != nil {
		t.Fatalf("expected enrolled student to download, got %v", err)
	}
}
