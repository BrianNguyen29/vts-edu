package gradebook

import (
	"time"
)

// AssessmentAttempt is a single submission for an assessment.
type AssessmentAttempt struct {
	ID            string     `json:"id"`
	AssessmentID  string     `json:"assessment_id"`
	StudentUserID string     `json:"student_user_id"`
	StudentName   string     `json:"student_name"`
	Status        string     `json:"status"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	Score         *string    `json:"score,omitempty"`
	MaxScore      *string    `json:"max_score,omitempty"`
	GradingStatus *string    `json:"grading_status,omitempty"`
}

// AssessmentResult is a numeric summary of attempts for an assessment.
type AssessmentResult struct {
	AssessmentID    string  `json:"assessment_id"`
	TotalAttempts   int64   `json:"total_attempts"`
	SubmittedCount  int64   `json:"submitted_count"`
	InProgressCount int64   `json:"in_progress_count"`
	ExpiredCount    int64   `json:"expired_count"`
	AverageScore    *string `json:"average_score,omitempty"`
	MaxScore        *string `json:"max_score,omitempty"`
}

// ClassGradebookEntry is one row in a class gradebook.
type ClassGradebookEntry struct {
	StudentUserID   string     `json:"student_user_id"`
	StudentName     string     `json:"student_name"`
	AssessmentID    string     `json:"assessment_id"`
	AssessmentTitle string     `json:"assessment_title"`
	AttemptID       *string    `json:"attempt_id,omitempty"`
	Status          *string    `json:"status,omitempty"`
	Score           *string    `json:"score,omitempty"`
	MaxScore        *string    `json:"max_score,omitempty"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
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
