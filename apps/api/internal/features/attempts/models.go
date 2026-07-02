package attempts

import (
	"encoding/json"
	"time"
)

// AttemptSnapshot is the runtime view of an attempt and its items/answers.
type AttemptSnapshot struct {
	ID             string        `json:"id"`
	OrganizationID string        `json:"organization_id"`
	AssessmentID   string        `json:"assessment_id"`
	PublicationID  *string       `json:"publication_id,omitempty"`
	Status         string        `json:"status"`
	StartedAt      *time.Time    `json:"started_at,omitempty"`
	ExpiresAt      *time.Time    `json:"expires_at,omitempty"`
	SubmittedAt    *time.Time    `json:"submitted_at,omitempty"`
	ServerTime     time.Time     `json:"server_time"`
	Items          []AttemptItem `json:"items"`
}

// AssignedAssessment is a published assessment available to the current student.
type AssignedAssessment struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Status          string  `json:"status"`
	Availability    string  `json:"availability"`
	DurationMinutes int     `json:"duration_minutes"`
	MaxAttempts     int     `json:"max_attempts"`
	AttemptsUsed    int     `json:"attempts_used"`
	Revision        int     `json:"revision"`
	PublicationID   string  `json:"publication_id"`
	PublishedAt     string  `json:"published_at"`
	OpensAt         *string `json:"opens_at,omitempty"`
	ClosesAt        *string `json:"closes_at,omitempty"`
}

// StudentAttempt is a single entry in a student's attempt history.
type StudentAttempt struct {
	ID              string     `json:"id"`
	AssessmentID    string     `json:"assessment_id"`
	AssessmentTitle string     `json:"assessment_title"`
	Status          string     `json:"status"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
	Score           *string    `json:"score,omitempty"`
	MaxScore        *string    `json:"max_score,omitempty"`
	GradingStatus   *string    `json:"grading_status,omitempty"`
}

// AttemptResult is the graded review view of a submitted or expired attempt.
type AttemptResult struct {
	ID            string              `json:"id"`
	AssessmentID  string              `json:"assessment_id"`
	Status        string              `json:"status"`
	SubmittedAt   *time.Time          `json:"submitted_at,omitempty"`
	Score         string              `json:"score"`
	MaxScore      string              `json:"max_score"`
	GradingStatus string              `json:"grading_status"`
	ServerTime    time.Time           `json:"server_time"`
	Items         []AttemptResultItem `json:"items"`
}

// AttemptResultItem is a single question in an attempt result review.
type AttemptResultItem struct {
	ID                string               `json:"id"`
	QuestionVersionID string               `json:"question_version_id"`
	QuestionType      string               `json:"question_type"`
	Position          int                  `json:"position"`
	Points            string               `json:"points"`
	Prompt            json.RawMessage      `json:"prompt"`
	Choices           json.RawMessage      `json:"choices"`
	CorrectAnswer     json.RawMessage      `json:"correct_answer"`
	StudentAnswer     *AttemptResultAnswer `json:"student_answer,omitempty"`
	GradingStatus     string               `json:"grading_status"`
	IsCorrect         *bool                `json:"is_correct,omitempty"`
	AwardedScore      *string              `json:"awarded_score,omitempty"`
	Feedback          *string              `json:"feedback,omitempty"`
}

// AttemptResultAnswer is the student's answer for a result item.
type AttemptResultAnswer struct {
	AnswerPayload json.RawMessage `json:"answer_payload"`
	AnsweredAt    time.Time       `json:"answered_at"`
}

// PublicationSnapshot mirrors the JSON stored in assessment_publications.snapshot_json.
type PublicationSnapshot struct {
	ID              string               `json:"id"`
	Title           string               `json:"title"`
	DurationMinutes int                  `json:"duration_minutes"`
	MaxAttempts     int                  `json:"max_attempts"`
	Instructions    string               `json:"instructions"`
	OpensAt         *string              `json:"opens_at"`
	ClosesAt        *string              `json:"closes_at"`
	Revision        int                  `json:"revision"`
	Sections        []PublicationSection `json:"sections"`
}

// PublicationSection is a section inside a published assessment snapshot.
type PublicationSection struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Position int               `json:"position"`
	Items    []PublicationItem `json:"items"`
}

// PublicationItem is an item inside a published assessment snapshot.
type PublicationItem struct {
	ID                string          `json:"id"`
	QuestionVersionID string          `json:"question_version_id"`
	QuestionType      string          `json:"question_type"`
	Position          int             `json:"position"`
	Points            string          `json:"points"`
	Prompt            json.RawMessage `json:"prompt"`
	Choices           json.RawMessage `json:"choices"`
	AnswerKey         json.RawMessage `json:"answer_key"`
	MaxScore          string          `json:"max_score"`
}

// AttemptItem is a single question inside an attempt.
type AttemptItem struct {
	ID                string          `json:"id"`
	QuestionVersionID string          `json:"question_version_id"`
	QuestionType      string          `json:"question_type"`
	Position          int             `json:"position"`
	Points            string          `json:"points"`
	Prompt            json.RawMessage `json:"prompt"`
	Choices           json.RawMessage `json:"choices"`
	Answer            *AnswerSnapshot `json:"answer,omitempty"`
}

// AnswerSnapshot is the stored answer for an attempt item.
type AnswerSnapshot struct {
	AnswerPayload json.RawMessage `json:"answer_payload"`
	Revision      int64           `json:"revision"`
	AnsweredAt    time.Time       `json:"answered_at"`
}

// SaveAnswerRequest is the payload for PUT .../answers/{item_id}.
type SaveAnswerRequest struct {
	AnswerPayload json.RawMessage `json:"answer_payload"`
}

// AnswerSaved is the data envelope for a successful answer save.
//
// ServerTime and ExpiresAt are the authoritative server clock and the
// attempt's current expiration; the front-end uses them to recalibrate
// its countdown offset and to refresh its snapshot without an extra
// round trip.
type AnswerSaved struct {
	AttemptItemID string          `json:"attempt_item_id"`
	Revision      int64           `json:"revision"`
	AnswerPayload json.RawMessage `json:"answer_payload"`
	AnsweredAt    time.Time       `json:"answered_at"`
	ServerTime    time.Time       `json:"server_time"`
	ExpiresAt     *time.Time      `json:"expires_at,omitempty"`
}

// AttemptSubmitted is the data envelope for a successful submit.
type AttemptSubmitted struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"`
	SubmittedAt   time.Time `json:"submitted_at"`
	Score         string    `json:"score"`
	MaxScore      string    `json:"max_score"`
	GradingStatus string    `json:"grading_status"`
}

// ListOptions filters and paginates list queries.
type ListOptions struct {
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
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}
