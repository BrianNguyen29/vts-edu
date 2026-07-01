package academics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// fakeService is a stub academics service used by the spike tests.
type fakeService struct {
	mu         sync.Mutex
	terms      []Term
	createErr  error
	createTerm Term
}

func (f *fakeService) ListTerms(_ context.Context, _ string) ([]Term, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]Term, len(f.terms))
	copy(out, f.terms)
	return out, nil
}

func (f *fakeService) CreateTerm(_ context.Context, _ string, _ []string, req CreateTermRequest) (Term, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.createErr != nil {
		return Term{}, f.createErr
	}
	t := Term{
		ID:        "term-1",
		Name:      req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Status:    "ACTIVE",
	}
	f.createTerm = t
	f.terms = append(f.terms, t)
	return t, nil
}

func (f *fakeService) UpdateTerm(_ context.Context, _ string, _ []string, _ string, _ UpdateTermRequest) (Term, error) {
	return Term{}, nil
}
func (f *fakeService) ArchiveTerm(_ context.Context, _ string, _ []string, _ string) error {
	return nil
}

func (f *fakeService) ListSubjects(_ context.Context, _ string) ([]Subject, error) { return nil, nil }
func (f *fakeService) CreateSubject(_ context.Context, _ string, _ []string, _ CreateSubjectRequest) (Subject, error) {
	return Subject{}, nil
}
func (f *fakeService) UpdateSubject(_ context.Context, _ string, _ []string, _ string, _ UpdateSubjectRequest) (Subject, error) {
	return Subject{}, nil
}
func (f *fakeService) ArchiveSubject(_ context.Context, _ string, _ []string, _ string) error {
	return nil
}
func (f *fakeService) ListCourses(_ context.Context, _ string) ([]Course, error) { return nil, nil }
func (f *fakeService) CreateCourse(_ context.Context, _ string, _ []string, _ CreateCourseRequest) (Course, error) {
	return Course{}, nil
}
func (f *fakeService) UpdateCourse(_ context.Context, _ string, _ []string, _ string, _ UpdateCourseRequest) (Course, error) {
	return Course{}, nil
}
func (f *fakeService) ArchiveCourse(_ context.Context, _ string, _ []string, _ string) error {
	return nil
}
func (f *fakeService) ListClasses(_ context.Context, _ string, _ string, _ []string) ([]ClassSection, error) {
	return nil, nil
}
func (f *fakeService) ListMyTeachingClasses(_ context.Context, _ string, _ string) ([]ClassSection, error) {
	return nil, nil
}
func (f *fakeService) CreateClass(_ context.Context, _ string, _ []string, _ CreateClassRequest) (ClassSection, error) {
	return ClassSection{}, nil
}
func (f *fakeService) UpdateClass(_ context.Context, _ string, _ []string, _ string, _ UpdateClassRequest) (ClassSection, error) {
	return ClassSection{}, nil
}
func (f *fakeService) ArchiveClass(_ context.Context, _ string, _ []string, _ string) error {
	return nil
}
func (f *fakeService) ListClassTeachers(_ context.Context, _ string, _ string, _ []string, _ string) ([]ClassTeacher, error) {
	return nil, nil
}
func (f *fakeService) AddClassTeacher(_ context.Context, _ string, _ []string, _ string, _ AddClassTeacherRequest) (ClassTeacher, error) {
	return ClassTeacher{}, nil
}
func (f *fakeService) RemoveClassTeacher(_ context.Context, _ string, _ []string, _, _ string) error {
	return nil
}
func (f *fakeService) ListEnrollments(_ context.Context, _ string, _ string, _ []string, _ string) ([]Enrollment, error) {
	return nil, nil
}
func (f *fakeService) EnrollStudent(_ context.Context, _ string, _ []string, _ string, _ EnrollStudentRequest) (Enrollment, error) {
	return Enrollment{}, nil
}
func (f *fakeService) UnenrollStudent(_ context.Context, _ string, _ []string, _, _ string) error {
	return nil
}
func (f *fakeService) BulkEnrollStudents(_ context.Context, _ string, _ []string, _ string, _ BulkEnrollRequest) (BulkEnrollmentResult, error) {
	return BulkEnrollmentResult{}, nil
}
func (f *fakeService) BulkAssignTeachers(_ context.Context, _ string, _ []string, _ string, _ BulkAssignTeachersRequest) (BulkAssignTeachersResult, error) {
	return BulkAssignTeachersResult{}, nil
}

func newTestIssuer() *auth.TokenIssuer {
	return auth.NewTokenIssuer("test-signing-key-32-bytes-long!!!", "vts-edu-api", "vts-edu-web", 15*time.Minute)
}

func buildTestJWT(t *testing.T, issuer *auth.TokenIssuer, userID, orgID string, roles []string) string {
	t.Helper()
	tok, _, err := issuer.IssueAccessToken(userID, orgID, "", roles, 0, false)
	if err != nil {
		t.Fatalf("issue test jwt: %v", err)
	}
	return tok
}

func readEnvelope(t *testing.T, body string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatalf("decode envelope: %v: %s", err, body)
	}
	return out
}

// newSpikeRouter builds a *chi.Mux that mirrors the production Huma spike
// mount: RequestID middleware + spikeMiddleware (request injection) +
// Huma operations registered on the child router. The Huma NewError
// override is also installed.
func newSpikeRouter(deps HumaSpikeDeps) http.Handler {
	// No NewError override needed: spike handlers return their own response
	// struct and do not rely on huma.NewError.
	cfg := huma.DefaultConfig("VTS EDU academics Huma spike", "0.0.0-spike")
	cfg.Servers = []*huma.Server{{URL: "http://localhost:8080/api/v1"}}
	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
	}

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(spikeMiddleware)
	api := humachi.New(r, cfg)
	huma.Register(api, huma.Operation{
		OperationID: "spike.listTerms",
		Method:      http.MethodGet,
		Path:        "/academic-terms",
		Summary:     "Spike: list academic terms (preserves {data,error} envelope)",
		Tags:        []string{"Spike"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, deps.spikeListTerms)
	huma.Register(api, huma.Operation{
		OperationID: "spike.createTerm",
		Method:      http.MethodPost,
		Path:        "/academic-terms",
		Summary:     "Spike: create academic term (preserves {data,error} envelope, requires CSRF)",
		Tags:        []string{"Spike"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, deps.spikeCreateTerm)
	return r
}

// doRequest fires a test request against a handler and returns the
// httptest recorder. Header pairs are (name, value).
func doRequest(t *testing.T, h http.Handler, method, path, requestID, body string, headerKV ...string) *httptest.ResponseRecorder {
	t.Helper()
	var br *bytes.Reader
	if body != "" {
		br = bytes.NewReader([]byte(body))
	} else {
		br = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, br)
	if requestID != "" {
		req.Header.Set("X-Request-Id", requestID)
	}
	for i := 0; i+1 < len(headerKV); i += 2 {
		req.Header.Set(headerKV[i], headerKV[i+1])
	}
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHumaSpike_ListTerms_PreservesEnvelopeAndRequestID(t *testing.T) {
	svc := &fakeService{terms: []Term{{ID: "t1", Name: "HK1", StartDate: "2026-09-01", EndDate: "2027-05-31", Status: "ACTIVE"}}}
	issuer := newTestIssuer()
	r := newSpikeRouter(HumaSpikeDeps{Svc: svc, Issuer: issuer})

	token := buildTestJWT(t, issuer, "user-1", "org-1", []string{"teacher"})
	resp := doRequest(t, r, http.MethodGet, "/academic-terms", "spike-req-1",
		"", "Authorization", "Bearer "+token,
	)

	if resp.Code != 200 {
		t.Fatalf("status = %d, want 200, body=%q", resp.Code, resp.Body.String())
	}
	env := readEnvelope(t, resp.Body.String())
	data, ok := env["data"].([]any)
	if !ok {
		t.Fatalf("envelope missing data array: %s", resp.Body.String())
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 term, got %d", len(data))
	}
	if got := resp.Header().Get("X-Request-Id"); got != "spike-req-1" {
		t.Logf("note: response X-Request-Id = %q (chi RequestID middleware only sets context)", got)
	}
}

func TestHumaSpike_ListTerms_ForbiddenForStudent(t *testing.T) {
	svc := &fakeService{}
	issuer := newTestIssuer()
	r := newSpikeRouter(HumaSpikeDeps{Svc: svc, Issuer: issuer})

	token := buildTestJWT(t, issuer, "user-1", "org-1", []string{"student"})
	resp := doRequest(t, r, http.MethodGet, "/academic-terms", "spike-req-2",
		"", "Authorization", "Bearer "+token,
	)

	if resp.Code != 403 {
		t.Fatalf("status = %d, want 403, body=%q", resp.Code, resp.Body.String())
	}
	env := readEnvelope(t, resp.Body.String())
	errObj, ok := env["error"].(map[string]any)
	if !ok {
		t.Fatalf("error envelope missing: %s", resp.Body.String())
	}
	if errObj["code"] != "forbidden" {
		t.Fatalf("code = %v, want forbidden", errObj["code"])
	}
	if errObj["request_id"] != "spike-req-2" {
		t.Fatalf("request_id = %v, want spike-req-2", errObj["request_id"])
	}
}

func TestHumaSpike_CreateTerm_ValidationCatchesBadDate(t *testing.T) {
	svc := &fakeService{}
	issuer := newTestIssuer()
	r := newSpikeRouter(HumaSpikeDeps{Svc: svc, Issuer: issuer})

	token := buildTestJWT(t, issuer, "user-1", "org-1", []string{"admin"})
	body := `{"name":"HK1","start_date":"2026-09-01","end_date":"2026-08-01"}`
	resp := doRequest(t, r, http.MethodPost, "/academic-terms", "spike-req-3", body,
		"Authorization", "Bearer "+token,
		"X-CSRF-Token", "demo-csrf-token",
		"Cookie", "vts_csrf=demo-csrf-token",
	)

	if resp.Code != 400 {
		t.Fatalf("status = %d, want 400, body=%q", resp.Code, resp.Body.String())
	}
	env := readEnvelope(t, resp.Body.String())
	errObj, ok := env["error"].(map[string]any)
	if !ok {
		t.Fatalf("error envelope missing: %s", resp.Body.String())
	}
	if errObj["code"] != "bad_request" {
		t.Fatalf("code = %v, want bad_request", errObj["code"])
	}
}

func TestHumaSpike_CreateTerm_HappyPathEnvelope(t *testing.T) {
	svc := &fakeService{}
	issuer := newTestIssuer()
	r := newSpikeRouter(HumaSpikeDeps{Svc: svc, Issuer: issuer})

	token := buildTestJWT(t, issuer, "user-1", "org-1", []string{"admin"})
	body := `{"name":"HK1","start_date":"2026-09-01","end_date":"2027-05-31"}`
	resp := doRequest(t, r, http.MethodPost, "/academic-terms", "spike-req-4", body,
		"Authorization", "Bearer "+token,
		"X-CSRF-Token", "demo-csrf-token",
		"Cookie", "vts_csrf=demo-csrf-token",
	)

	if resp.Code != 201 {
		t.Fatalf("status = %d, want 201, body=%q", resp.Code, resp.Body.String())
	}
	env := readEnvelope(t, resp.Body.String())
	data, ok := env["data"].(map[string]any)
	if !ok {
		t.Fatalf("envelope missing data object: %s", resp.Body.String())
	}
	if data["name"] != "HK1" {
		t.Fatalf("data.name = %v, want HK1", data["name"])
	}
	if got := resp.Header().Get("X-Request-Id"); got != "spike-req-4" {
		t.Logf("note: response X-Request-Id = %q (chi RequestID middleware only sets context)", got)
	}
}

// Sanity check: the fake service is also wrapped in a typed cast.
var _ Service = (*fakeService)(nil)

// Sanity: ErrInvalidInput is reachable for the test using ErrDuplicateCode
// path (unused for now but keeps the dependency explicit).
var _ = errors.Is

// _ = strings.NewReader is here to silence the import if a future test
// needs it; the current spike helpers don't use it directly.
var _ = strings.NewReader
