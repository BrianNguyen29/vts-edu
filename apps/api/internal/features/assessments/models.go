package assessments

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
