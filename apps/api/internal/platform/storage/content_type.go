package storage

import "strings"

// AllowedDownloadContentTypes is the conservative allowlist used by the
// resources download path. It is intentionally narrow: text/*, common
// image, PDF, and Office document types. Anything outside the list is
// served as `application/octet-stream` with a sanitized filename.
var AllowedDownloadContentTypes = map[string]struct{}{
	"text/plain":               {},
	"text/csv":                 {},
	"text/markdown":            {},
	"application/pdf":          {},
	"application/json":         {},
	"application/zip":          {},
	"application/octet-stream": {},
	"image/png":                {},
	"image/jpeg":               {},
	"image/gif":                {},
	"image/webp":               {},
	"image/svg+xml":            {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
	"application/vnd.ms-excel":      {},
	"application/vnd.ms-powerpoint": {},
	"application/msword":            {},
}

// SanitizeContentType returns a safe content type for HTTP responses. If
// the input is empty, malformed, or not on the allowlist, it returns
// "application/octet-stream". The lookup is case-insensitive.
func SanitizeContentType(ct string) string {
	ct = strings.TrimSpace(strings.ToLower(ct))
	if ct == "" {
		return "application/octet-stream"
	}
	// Strip parameters like "; charset=utf-8" before comparison.
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	if _, ok := AllowedDownloadContentTypes[ct]; ok {
		return ct
	}
	return "application/octet-stream"
}
