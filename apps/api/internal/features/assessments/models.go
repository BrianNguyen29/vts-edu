package assessments

import "encoding/json"

// Assessment is the persistence shape for an assessment row.
type Assessment struct {
	ID              string
	Title           string
	Status          string
	DurationMinutes int
	CreatedAt       string
}

// AssessmentListItem is the public list view for GET /api/v1/assessments.
type AssessmentListItem struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	DurationMinutes int    `json:"duration_minutes"`
}

// AssessmentDetail is the full builder view for GET /api/v1/assessments/{id}.
type AssessmentDetail struct {
	ID              string          `json:"id"`
	ClassSectionID  *string         `json:"class_section_id,omitempty"`
	Title           string          `json:"title"`
	Status          string          `json:"status"`
	DurationMinutes int             `json:"duration_minutes"`
	MaxAttempts     int             `json:"max_attempts"`
	Revision        int             `json:"revision"`
	Instructions    string          `json:"instructions,omitempty"`
	OpensAt         *string         `json:"opens_at,omitempty"`
	ClosesAt        *string         `json:"closes_at,omitempty"`
	Settings        json.RawMessage `json:"settings,omitempty"`
	Sections        []SectionDetail `json:"sections"`
	Targets         []TargetDetail  `json:"targets"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

// SectionDetail is a builder section with its items.
type SectionDetail struct {
	ID       string          `json:"id"`
	Title    string          `json:"title"`
	Position int             `json:"position"`
	Settings json.RawMessage `json:"settings,omitempty"`
	Items    []ItemDetail    `json:"items"`
}

// ItemDetail is a builder item referencing a question version.
type ItemDetail struct {
	ID                  string `json:"id"`
	AssessmentSectionID string `json:"assessment_section_id"`
	QuestionVersionID   string `json:"question_version_id"`
	Position            int    `json:"position"`
	Points              string `json:"points"`
}

// TargetDetail is a class target for an assessment.
type TargetDetail struct {
	ID             string `json:"id"`
	ClassSectionID string `json:"class_section_id"`
}

// CreateAssessmentRequest is the payload for POST /classes/{class_id}/assessments.
type CreateAssessmentRequest struct {
	Title           string `json:"title"`
	DurationMinutes int    `json:"duration_minutes"`
	MaxAttempts     int    `json:"max_attempts"`
}

// UpdateAssessmentRequest is the payload for PATCH /assessments/{id}.
type UpdateAssessmentRequest struct {
	Title           string          `json:"title,omitempty"`
	DurationMinutes *int            `json:"duration_minutes,omitempty"`
	MaxAttempts     *int            `json:"max_attempts,omitempty"`
	Instructions    string          `json:"instructions,omitempty"`
	OpensAt         string          `json:"opens_at,omitempty"`
	ClosesAt        string          `json:"closes_at,omitempty"`
	Settings        json.RawMessage `json:"settings,omitempty"`
}

// CreateSectionRequest is the payload for POST /assessments/{id}/sections.
type CreateSectionRequest struct {
	Title    string `json:"title"`
	Position int    `json:"position"`
}

// CreateItemRequest is the payload for POST /assessment-sections/{section_id}/items.
type CreateItemRequest struct {
	QuestionVersionID string `json:"question_version_id"`
	Position          int    `json:"position"`
	Points            string `json:"points,omitempty"`
}

// UpdateSectionRequest is the payload for PATCH /assessment-sections/{section_id}.
type UpdateSectionRequest struct {
	Title    string `json:"title,omitempty"`
	Position int    `json:"position,omitempty"`
}

// UpdateItemRequest is the payload for PATCH /assessment-items/{item_id}.
type UpdateItemRequest struct {
	QuestionVersionID string `json:"question_version_id,omitempty"`
	Position          int    `json:"position,omitempty"`
	Points            string `json:"points,omitempty"`
}

// ReorderSectionsRequest is the payload for POST /assessments/{id}/sections/reorder.
type ReorderSectionsRequest struct {
	SectionIDs []string `json:"section_ids"`
}

// ReorderItemsRequest is the payload for POST /assessment-sections/{section_id}/items/reorder.
type ReorderItemsRequest struct {
	ItemIDs []string `json:"item_ids"`
}

// CreateTargetRequest is the payload for POST /assessments/{id}/targets.
type CreateTargetRequest struct {
	ClassSectionID string `json:"class_section_id"`
}

// ListQuestionsOptions filters the question picker list.
type ListQuestionsOptions struct {
	Query  string
	BankID string
	Limit  int
	Offset int
}

// QuestionPickerItem is a question/version choice for the builder picker.
type QuestionPickerItem struct {
	ID                    string `json:"id"`
	QuestionBankID        string `json:"question_bank_id"`
	QuestionVersionID     string `json:"question_version_id"`
	QuestionVersionStatus string `json:"question_version_status"`
	Prompt                string `json:"prompt"`
}

// PublicationSummary is a row in the assessment publication history.
type PublicationSummary struct {
	ID          string `json:"id"`
	Version     int    `json:"version"`
	Status      string `json:"status"`
	PublishedAt string `json:"published_at"`
}

// PublishResult is returned by POST /assessments/{id}/publish.
type PublishResult struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Revision    int    `json:"revision"`
	PublishedAt string `json:"published_at"`
}

// ValidationError lists validation failures.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ListOptions is the optional pagination/search input for list endpoints.
type ListOptions struct {
	Query  string
	Limit  int
	Offset int
	Cursor string
	Count  bool
}

// PageInfo is returned with paginated list responses.
type PageInfo struct {
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
	TotalCount *int64  `json:"total_count,omitempty"`
}

// DataEnvelope wraps successful API responses.
type DataEnvelope struct {
	Data any       `json:"data"`
	Page *PageInfo `json:"page,omitempty"`
}

// ErrorEnvelope wraps API error responses.
type ErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
