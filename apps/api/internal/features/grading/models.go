package grading

import (
	"encoding/json"
	"time"
)

// ReviewQueueEntry is a single attempt summary in the per-assessment review queue.
type ReviewQueueEntry struct {
	AttemptID     string     `json:"attempt_id"`
	StudentUserID string     `json:"student_user_id"`
	StudentName   string     `json:"student_name"`
	Status        string     `json:"status"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	MaxScore      *string    `json:"max_score,omitempty"`
	PendingItems  int        `json:"pending_items"`
	TotalNonMcq   int        `json:"total_non_mcq"`
}

// GradingItemDetail is a single attempt item returned to the review detail page,
// including the student's answer and the current manual grade (if any).
type GradingItemDetail struct {
	ID                string                `json:"id"`
	QuestionVersionID string                `json:"question_version_id"`
	Position          int                   `json:"position"`
	Points            string                `json:"points"`
	QuestionType      string                `json:"question_type"`
	Prompt            json.RawMessage       `json:"prompt"`
	Choices           json.RawMessage       `json:"choices"`
	StudentAnswer     *GradingStudentAnswer `json:"student_answer,omitempty"`
	ItemGrade         *GradingItemGrade     `json:"item_grade,omitempty"`
}

// GradingStudentAnswer is the persisted student answer for the item.
type GradingStudentAnswer struct {
	AnswerPayload json.RawMessage `json:"answer_payload"`
	Revision      int64           `json:"revision"`
	AnsweredAt    time.Time       `json:"answered_at"`
}

// GradingItemGrade is the current manual grade row for the item (if any).
type GradingItemGrade struct {
	ID           string    `json:"id"`
	GraderUserID string    `json:"grader_user_id"`
	AwardedScore string    `json:"awarded_score"`
	Feedback     *string   `json:"feedback,omitempty"`
	GradedAt     time.Time `json:"graded_at"`
}

// AttemptGradingContext is the parent of GradingItemDetail when returning a
// full review detail view (assessment + student + items).
type AttemptGradingContext struct {
	AttemptID     string              `json:"attempt_id"`
	AssessmentID  string              `json:"assessment_id"`
	StudentUserID string              `json:"student_user_id"`
	StudentName   string              `json:"student_name"`
	Status        string              `json:"status"`
	Score         *string             `json:"score,omitempty"`
	MaxScore      *string             `json:"max_score,omitempty"`
	GradingStatus string              `json:"grading_status"`
	SubmittedAt   *time.Time          `json:"submitted_at,omitempty"`
	Items         []GradingItemDetail `json:"items"`
}

// GradeItemRequest is the PUT payload for grading a single item.
type GradeItemRequest struct {
	AwardedScore string  `json:"awarded_score"`
	Feedback     *string `json:"feedback,omitempty"`
}

// GradeItemResponse is the response from a successful grade.
type GradeItemResponse struct {
	ItemGrade     GradingItemGrade `json:"item_grade"`
	AttemptScore  string           `json:"attempt_score"`
	AttemptMax    string           `json:"attempt_max_score"`
	GradingStatus string           `json:"grading_status"`
	StillPending  int              `json:"still_pending_items"`
	TotalNonMcq   int              `json:"total_non_mcq_items"`
}

// PageInfo mirrors the attempts pagination shape.
type PageInfo struct {
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
	TotalCount *int64  `json:"total_count,omitempty"`
}

// DataEnvelope wraps a single response payload.
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
