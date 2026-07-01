package resources

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/go-chi/chi/v5"
)

// Handler exposes the resources HTTP endpoints.
type Handler struct {
	svc    Service
	issuer *auth.TokenIssuer
}

// NewHandler creates a resources HTTP handler.
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
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrNoActiveFile):
		writeError(w, r, http.StatusNotFound, "not_found", "resource or file not found")
	case errors.Is(err, ErrInvalidInput):
		writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
	case errors.Is(err, ErrInvalidStatus):
		writeError(w, r, http.StatusConflict, "conflict", err.Error())
	case errors.Is(err, ErrFileTooLarge):
		writeError(w, r, http.StatusRequestEntityTooLarge, "file_too_large", err.Error())
	case errors.Is(err, ErrStorageNotFound):
		writeError(w, r, http.StatusNotFound, "not_found", "stored object not found")
	case errors.Is(err, ErrStorageFailure):
		writeError(w, r, http.StatusInternalServerError, "storage_error", "failed to process file")
	default:
		writeError(w, r, http.StatusInternalServerError, "internal_error", "resource operation failed")
	}
}

// ListResources handles GET /api/v1/resources.
func (h *Handler) ListResources(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	resources, err := h.svc.ListResources(r.Context(), actor)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	enveloped := make([]DataEnvelope, len(resources))
	for i, res := range resources {
		enveloped[i] = DataEnvelope{Data: res}
	}
	writeData(w, http.StatusOK, enveloped)
}

// CreateResource handles POST /api/v1/resources.
func (h *Handler) CreateResource(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	var req CreateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	resource, err := h.svc.CreateResource(r.Context(), actor, req)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, resource)
}

// PublishResource handles POST /api/v1/resources/{id}/publish.
func (h *Handler) PublishResource(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	resource, err := h.svc.PublishResource(r.Context(), actor, id)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, resource)
}

// ArchiveResource handles DELETE /api/v1/resources/{id}.
func (h *Handler) ArchiveResource(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.ArchiveResource(r.Context(), actor, id); err != nil {
		h.mapError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UploadFile handles POST /api/v1/resources/{id}/files.
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")

	const maxMemory = 32 << 20 // 32 MiB
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid multipart form")
		return
	}
	defer func() {
		_ = r.MultipartForm.RemoveAll()
	}()

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		writeError(w, r, http.StatusBadRequest, "bad_request", "missing file field")
		return
	}
	fh := files[0]

	file, err := fh.Open()
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "cannot read uploaded file")
		return
	}
	defer file.Close()

	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploaded, err := h.svc.UploadFile(r.Context(), actor, id, fh.Filename, contentType, file, fh.Size)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, uploaded)
}

// DownloadFile handles GET /api/v1/resources/{id}/download.
func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	reader, file, err := h.svc.DownloadFile(r.Context(), actor, id)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.SizeBytes, 10))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+safeFilename(file.OriginalName)+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, reader)
}

// safeFilename strips path separators and control characters from a
// user-supplied filename before it is emitted in a header value.
func safeFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "\n", "_")
	name = strings.ReplaceAll(name, "\r", "_")
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return '_'
		}
		return r
	}, name)
	return name
}
