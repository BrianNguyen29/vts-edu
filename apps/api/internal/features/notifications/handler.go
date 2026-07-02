package notifications

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
	"github.com/go-chi/chi/v5"
)

// DataEnvelope is the standard success response shape.
type DataEnvelope struct {
	Data any `json:"data"`
}

// ErrorEnvelope is the standard error response shape.
type ErrorEnvelope struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

// Handler exposes the notifications HTTP endpoints. All routes
// require a valid access token; the auth provider enforces that
// before we reach the actor. Tenant isolation is enforced via
// actor.OrgID in the service layer.
type Handler struct {
	svc    Service
	issuer *auth.TokenIssuer
}

// NewHandler creates a notifications HTTP handler.
func NewHandler(svc Service, issuer *auth.TokenIssuer) *Handler {
	return &Handler{svc: svc, issuer: issuer}
}

func (h *Handler) actor(w http.ResponseWriter, r *http.Request) (auth.Actor, bool) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return auth.Actor{}, false
	}
	return actor, true
}

func (h *Handler) mapError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		writeError(w, r, http.StatusForbidden, "forbidden", "insufficient permissions")
	case errors.Is(err, ErrNotFound):
		writeError(w, r, http.StatusNotFound, "not_found", "notification not found")
	case errors.Is(err, ErrInvalidInput):
		writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
	default:
		writeError(w, r, http.StatusInternalServerError, "internal_error", "notification operation failed")
	}
}

// List handles GET /api/v1/me/notifications?limit=&before=
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	var before *time.Time
	if v := r.URL.Query().Get("before"); v != "" {
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			before = &t
		}
	}
	rows, err := h.svc.List(r.Context(), actor, before, limit)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	enveloped := make([]DataEnvelope, len(rows))
	for i, row := range rows {
		enveloped[i] = DataEnvelope{Data: toWire(row)}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": enveloped})
}

// UnreadCount handles GET /api/v1/me/notifications/unread-count
func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	n, err := h.svc.UnreadCount(r.Context(), actor)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]int{"count": n}})
}

// MarkRead handles POST /api/v1/me/notifications/{id}/read
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, r, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	id := chi.URLParam(r, "id")
	row, err := h.svc.MarkRead(r.Context(), actor, id)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(DataEnvelope{Data: toWire(row)})
}

// notificationWire is the public JSON shape. The raw jsonb bytes are
// surfaced as a parsed map so the client does not have to re-decode.
type notificationWire struct {
	ID          string         `json:"id"`
	OrgID       string         `json:"organization_id"`
	RecipientID string         `json:"recipient_user_id"`
	EventType   string         `json:"event_type"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	Metadata    map[string]any `json:"metadata"`
	IsRead      bool           `json:"is_read"`
	ReadAt      *string        `json:"read_at,omitempty"`
	CreatedAt   string         `json:"created_at"`
}

func toWire(n Notification) notificationWire {
	meta, _ := DecodeMetadata(n.MetadataJSON)
	return notificationWire{
		ID:          n.ID,
		OrgID:       n.OrgID,
		RecipientID: n.RecipientID,
		EventType:   n.EventType,
		Title:       n.Title,
		Body:        n.Body,
		Metadata:    meta,
		IsRead:      n.IsRead,
		ReadAt:      n.ReadAt,
		CreatedAt:   n.CreatedAt,
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorEnvelope{}
	resp.Error.Code = code
	resp.Error.Message = message
	_ = json.NewEncoder(w).Encode(resp)
}
