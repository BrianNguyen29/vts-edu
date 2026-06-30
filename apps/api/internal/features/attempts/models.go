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
	ID              string `json:"id"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	DurationMinutes int    `json:"duration_minutes"`
	MaxAttempts     int    `json:"max_attempts"`
	Revision        int    `json:"revision"`
	PublicationID   string `json:"publication_id"`
	PublishedAt     string `json:"published_at"`
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
type AnswerSaved struct {
	AttemptItemID string          `json:"attempt_item_id"`
	Revision      int64           `json:"revision"`
	AnswerPayload json.RawMessage `json:"answer_payload"`
	AnsweredAt    time.Time       `json:"answered_at"`
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
