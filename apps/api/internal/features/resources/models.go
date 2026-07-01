package resources

import "errors"

const (
	ContextTypeOrganization = "organization"
	ContextTypeClass        = "class"

	StatusDraft     = "DRAFT"
	StatusPublished = "PUBLISHED"
	StatusArchived  = "ARCHIVED"

	FileStatusActive   = "ACTIVE"
	FileStatusArchived = "ARCHIVED"
)

// Resource is a shared learning material scoped to an organization or class.
type Resource struct {
	ID          string  `json:"id"`
	OrgID       string  `json:"organization_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ContextType string  `json:"context_type"`
	ContextID   string  `json:"context_id"`
	Status      string  `json:"status"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	PublishedAt *string `json:"published_at,omitempty"`
}

// ResourceFile is a stored file attached to a resource.
type ResourceFile struct {
	ID           string `json:"id"`
	ResourceID   string `json:"resource_id"`
	OrgID        string `json:"organization_id"`
	OriginalName string `json:"original_name"`
	StorageKey   string `json:"storage_key"`
	ContentType  string `json:"content_type"`
	SizeBytes    int64  `json:"size_bytes"`
	Status       string `json:"status"`
	CreatedBy    string `json:"created_by"`
	CreatedAt    string `json:"created_at"`
}

// CreateResourceRequest is the payload to create a resource metadata record.
type CreateResourceRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ContextType string `json:"context_type"`
	ContextID   string `json:"context_id"`
}

// UpdateResourceStatusRequest is the payload to publish/archive a resource.
type UpdateResourceStatusRequest struct {
	Status string `json:"status"`
}

// ResourceWithFiles bundles a resource with its attached files.
type ResourceWithFiles struct {
	Resource
	Files []ResourceFile `json:"files"`
}

// Common domain errors.
var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrNotFound        = errors.New("resource not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidStatus   = errors.New("invalid status transition")
	ErrFileTooLarge    = errors.New("file too large")
	ErrNoActiveFile    = errors.New("no active file for download")
	ErrStorageNotFound = errors.New("stored object not found")
	ErrStorageFailure  = errors.New("storage failure")
)
