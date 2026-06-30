package assessments

// Assessment is the persistence shape for an assessment row.
type Assessment struct {
	ID              string
	Title           string
	Status          string
	DurationMinutes int
}

// AssessmentListItem is the public list view for GET /api/v1/assessments.
type AssessmentListItem struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	DurationMinutes int    `json:"duration_minutes"`
}

// DataEnvelope wraps successful API responses.
type DataEnvelope struct {
	Data any `json:"data"`
}

// ErrorEnvelope wraps API error responses.
type ErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
