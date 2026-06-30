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
	Items          []AttemptItem `json:"items"`
}

// AttemptItem is a single question inside an attempt.
type AttemptItem struct {
	ID                string          `json:"id"`
	QuestionVersionID string          `json:"question_version_id"`
	Position          int             `json:"position"`
	Points            string          `json:"points"`
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
