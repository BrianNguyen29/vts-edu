package academics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// HumaSpikeDeps wires dependencies needed by the bounded Huma spike.
type HumaSpikeDeps struct {
	Svc    Service
	Issuer *auth.TokenIssuer
}

// spikeSubRouter is the prefix under which the Huma spike is mounted.
// Existing routes are untouched.
const spikeSubRouter = "/_spike/huma"

// spikeErrBody is the inner shape of the spike's error envelope.
type spikeErrBody struct {
	Code      string `json:"code" required:"true"`
	Message   string `json:"message" required:"true"`
	RequestID string `json:"request_id,omitempty"`
}

// spikeListResp is the Huma response for ListTerms. The Status field
// drives the HTTP status code; the Body carries either a data array
// (success) or an error object. Pointers + omitempty ensure exactly one
// of {data, error} is serialised per response.
type spikeListResp struct {
	Status int
	Body   struct {
		Data  []Term        `json:"data,omitempty"`
		Error *spikeErrBody `json:"error,omitempty"`
	}
}

// spikeCreateResp is the Huma response for CreateTerm.
type spikeCreateResp struct {
	Status int
	Body   struct {
		Data  *Term         `json:"data,omitempty"`
		Error *spikeErrBody `json:"error,omitempty"`
	}
}

// newSpikeListErr builds a spikeListResp carrying an error envelope. The
// request_id is pulled from the spike context (set by spikeMiddleware).
func newSpikeListErr(ctx context.Context, status int, code, message string) *spikeListResp {
	r := &spikeListResp{Status: status}
	r.Body.Error = &spikeErrBody{Code: code, Message: message}
	if reqID := spikeRequestIDFromContext(ctx); reqID != "" {
		r.Body.Error.RequestID = reqID
	}
	return r
}

// newSpikeCreateErr builds a spikeCreateResp carrying an error envelope.
func newSpikeCreateErr(ctx context.Context, status int, code, message string) *spikeCreateResp {
	r := &spikeCreateResp{Status: status}
	r.Body.Error = &spikeErrBody{Code: code, Message: message}
	if reqID := spikeRequestIDFromContext(ctx); reqID != "" {
		r.Body.Error.RequestID = reqID
	}
	return r
}

// spikeRequestIDFromContext returns the X-Request-Id header value if set
// on the request, or an empty string.
func spikeRequestIDFromContext(ctx context.Context) string {
	if r, ok := spikeRequestFromContext(ctx); ok {
		return r.Header.Get("X-Request-Id")
	}
	return ""
}

// MountHumaSpike wires a Huma v2 sub-router under spikeSubRouter on the
// given *chi.Mux. The spike covers only ListTerms and CreateTerm from
// academics. It preserves the {data} success envelope and the
// {error: {code,message}} error envelope (request_id is on the header).
func MountHumaSpike(r *chi.Mux, deps HumaSpikeDeps) huma.API {
	humaConfig := huma.DefaultConfig("VTS EDU academics Huma spike", "0.0.0-spike")
	humaConfig.Info.Description = "Bounded Huma v2 feasibility spike on academics. NOT a production migration."
	humaConfig.Servers = []*huma.Server{{
		URL:         "http://localhost:8080/api/v1",
		Description: "Local dev",
	}}
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	// Disable the default $schema embed so spike responses match the
	// production {data} / {error} envelope shape exactly. The spike
	// report calls this out as a Huma config knob that would need
	// similar handling for any production migration.
	humaConfig.CreateHooks = nil
	humaConfig.Transformers = nil

	// Wrap a child chi.Mux so humachi can mount cleanly. The child
	// carries the request-id middleware so the X-Request-Id header is
	// available for the spike's error envelope.
	child := chi.NewRouter()
	child.Use(middleware.RequestID)
	child.Use(spikeMiddleware)
	api := humachi.New(child, humaConfig)
	r.Mount(spikeSubRouter, child)

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

	return api
}

// spikeRequestKey is the context value used to expose the *http.Request to
// spike handlers. The chi-level spikeMiddleware sets it on production;
// the test helper sets it manually.
type spikeRequestKey struct{}

// spikeMiddleware is the chi-level middleware that injects the request
// into the context for spike handlers. It runs on the child router in
// production.
func spikeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), spikeRequestKey{}, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// spikeRequestFromContext returns the *http.Request set on the context.
func spikeRequestFromContext(ctx context.Context) (*http.Request, bool) {
	if r, ok := ctx.Value(spikeRequestKey{}).(*http.Request); ok {
		return r, true
	}
	return nil, false
}

// actorFromContext extracts the actor from the request's Authorization
// header via the existing auth.ActorFromRequest helper. CSRF is checked
// separately for write operations.
func (d HumaSpikeDeps) actorFromContext(ctx context.Context) (auth.Actor, error) {
	r, ok := spikeRequestFromContext(ctx)
	if !ok {
		return auth.Actor{}, fmt.Errorf("no request on context")
	}
	actor, err := auth.ActorFromRequest(r, d.Issuer)
	if err != nil {
		return auth.Actor{}, err
	}
	return actor, nil
}

// csrfFromContext validates CSRF using the existing csrf package.
func csrfFromContext(ctx context.Context) bool {
	r, ok := spikeRequestFromContext(ctx)
	if !ok {
		return false
	}
	return csrf.Validate(r)
}

// spikeListTerms handles GET /spike/huma/academic-terms.
func (d HumaSpikeDeps) spikeListTerms(ctx context.Context, _ *struct{}) (*spikeListResp, error) {
	actor, err := d.actorFromContext(ctx)
	if err != nil {
		return newSpikeListErr(ctx, http.StatusUnauthorized, "unauthorized", "missing or invalid access token"), nil
	}
	if !actorHasAny(actor.Roles, []string{"teacher", "admin"}) {
		return newSpikeListErr(ctx, http.StatusForbidden, "forbidden", "teacher or admin access required"), nil
	}
	terms, err := d.Svc.ListTerms(ctx, actor.OrgID)
	if err != nil {
		return newSpikeListErr(ctx, http.StatusInternalServerError, "internal_error", "academics operation failed"), nil
	}
	if terms == nil {
		terms = []Term{}
	}
	resp := &spikeListResp{Status: http.StatusOK}
	resp.Body.Data = terms
	return resp, nil
}

// spikeCreateTermInput is the Huma-friendly create body.
type spikeCreateTermInput struct {
	Body struct {
		Name      string `json:"name" minLength:"1" maxLength:"255"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	} `json:"body"`
}

// spikeCreateTerm handles POST /spike/huma/academic-terms.
func (d HumaSpikeDeps) spikeCreateTerm(ctx context.Context, in *spikeCreateTermInput) (*spikeCreateResp, error) {
	actor, err := d.actorFromContext(ctx)
	if err != nil {
		return newSpikeCreateErr(ctx, http.StatusUnauthorized, "unauthorized", "missing or invalid access token"), nil
	}
	if !actorHasAny(actor.Roles, []string{"admin"}) {
		return newSpikeCreateErr(ctx, http.StatusForbidden, "forbidden", "admin access required"), nil
	}
	if !csrfFromContext(ctx) {
		return newSpikeCreateErr(ctx, http.StatusForbidden, "invalid_csrf", "invalid csrf token"), nil
	}
	startDate, err := time.Parse("2006-01-02", in.Body.StartDate)
	if err != nil {
		return newSpikeCreateErr(ctx, http.StatusBadRequest, "bad_request", "invalid start_date"), nil
	}
	endDate, err := time.Parse("2006-01-02", in.Body.EndDate)
	if err != nil {
		return newSpikeCreateErr(ctx, http.StatusBadRequest, "bad_request", "invalid end_date"), nil
	}
	if startDate.After(endDate) {
		return newSpikeCreateErr(ctx, http.StatusBadRequest, "bad_request", "start_date must be on or before end_date"), nil
	}
	term, err := d.Svc.CreateTerm(ctx, actor.OrgID, actor.Roles, CreateTermRequest{
		Name:      in.Body.Name,
		StartDate: in.Body.StartDate,
		EndDate:   in.Body.EndDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrUnauthorized):
			return newSpikeCreateErr(ctx, http.StatusForbidden, "forbidden", err.Error()), nil
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrDuplicateCode):
			return newSpikeCreateErr(ctx, http.StatusBadRequest, "bad_request", err.Error()), nil
		default:
			return newSpikeCreateErr(ctx, http.StatusInternalServerError, "internal_error", "academics operation failed"), nil
		}
	}
	_ = startDate
	_ = endDate
	resp := &spikeCreateResp{Status: http.StatusCreated}
	resp.Body.Data = &term
	return resp, nil
}

// actorHasAny returns true if the actor has any of the wanted roles.
func actorHasAny(roles []string, wanted []string) bool {
	for _, r := range roles {
		for _, w := range wanted {
			if r == w {
				return true
			}
		}
	}
	return false
}
