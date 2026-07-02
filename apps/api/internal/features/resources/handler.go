package resources

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/storage"
	"github.com/go-chi/chi/v5"
)

// Allowed preview MIME types for inline rendering. Must be a strict
// subset of storage.AllowedDownloadContentTypes and limited to types the
// browser can render without an external plugin.
var inlinePreviewContentTypes = map[string]struct{}{
	"text/plain":      {},
	"text/csv":        {},
	"text/markdown":   {},
	"application/pdf": {},
	"image/png":       {},
	"image/jpeg":      {},
	"image/gif":       {},
	"image/webp":      {},
	"image/svg+xml":   {},
}

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
	case errors.Is(err, errClassAccessUnavailable):
		writeError(w, r, http.StatusBadGateway, "upstream_unavailable", "class access check failed")
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

	filter := ListFilter{
		ContextType: strings.TrimSpace(r.URL.Query().Get("context_type")),
		ContextID:   strings.TrimSpace(r.URL.Query().Get("context_id")),
	}
	resources, err := h.svc.ListResources(r.Context(), actor, filter)
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

// UploadFile handles POST /api/v1/resources/{id}/files. Accepts a
// multipart body with either a `file` field (single) or a `files[]`
// repeated field (multi). The single-field form is preserved for
// backward compatibility.
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

	multi := r.MultipartForm.File["files[]"]
	if len(multi) == 0 {
		multi = r.MultipartForm.File["files"]
	}
	if len(multi) == 0 {
		multi = r.MultipartForm.File["file"]
	}
	if len(multi) == 0 {
		writeError(w, r, http.StatusBadRequest, "bad_request", "missing file field")
		return
	}

	inputs := make([]UploadInput, 0, len(multi))
	opened := make([]io.ReadCloser, 0, len(multi))
	defer func() {
		for _, c := range opened {
			_ = c.Close()
		}
	}()

	for _, fh := range multi {
		f, err := fh.Open()
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "bad_request", "cannot read uploaded file")
			return
		}
		opened = append(opened, f)
		ct := fh.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/octet-stream"
		}
		inputs = append(inputs, UploadInput{
			FileName:    fh.Filename,
			ContentType: ct,
			Data:        f,
			Size:        fh.Size,
		})
	}

	stored, err := h.svc.UploadFiles(r.Context(), actor, id, inputs)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	// Close opened files eagerly now that storage has read them.
	for _, c := range opened {
		_ = c.Close()
	}
	opened = nil
	enveloped := make([]DataEnvelope, len(stored))
	for i, f := range stored {
		enveloped[i] = DataEnvelope{Data: f}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(enveloped)
}

// ListFiles handles GET /api/v1/resources/{id}/files.
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	files, err := h.svc.ListFiles(r.Context(), actor, id)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	enveloped := make([]DataEnvelope, len(files))
	for i, f := range files {
		enveloped[i] = DataEnvelope{Data: f}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(enveloped)
}

// DownloadFile handles GET /api/v1/resources/{id}/download.
// Optional query parameters:
//   - file_id=<uuid>  – download a specific file. Defaults to the active
//     file when omitted (backward compatible).
//   - disposition=inline – render in the browser when the file's MIME is
//     in inlinePreviewContentTypes. Anything else falls back to
//     attachment and the X-Content-Type-Options: nosniff header is kept.
func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	fileID := r.URL.Query().Get("file_id")

	reader, file, err := h.svc.DownloadFile(r.Context(), actor, id, fileID)
	if err != nil {
		h.mapError(w, r, err)
		return
	}
	defer reader.Close()

	contentType := storage.SanitizeContentType(file.ContentType)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.SizeBytes, 10))
	w.Header().Set("X-Content-Type-Options", "nosniff")

	disposition := r.URL.Query().Get("disposition")
	dispType := "attachment"
	if disposition == "inline" {
		if _, ok := inlinePreviewContentTypes[contentType]; ok {
			dispType = "inline"
		}
	}
	w.Header().Set("Content-Disposition", dispType+"; filename=\""+safeFilename(file.OriginalName)+"\"")
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
