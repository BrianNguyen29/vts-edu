package gradebook

import (
	"bytes"
	"encoding/csv"
	"time"
)

func renderAssessmentAttemptsCSV(attempts []AssessmentAttempt) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"attempt_id",
		"assessment_id",
		"student_user_id",
		"student_name",
		"status",
		"started_at",
		"expires_at",
		"submitted_at",
		"score",
		"max_score",
		"grading_status",
	})
	for _, a := range attempts {
		_ = w.Write([]string{
			a.ID,
			a.AssessmentID,
			a.StudentUserID,
			a.StudentName,
			a.Status,
			formatTime(a.StartedAt),
			formatTime(a.ExpiresAt),
			formatTime(a.SubmittedAt),
			ptrString(a.Score),
			ptrString(a.MaxScore),
			ptrString(a.GradingStatus),
		})
	}
	w.Flush()
	return buf.Bytes()
}

func renderClassGradebookCSV(entries []ClassGradebookEntry) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"student_user_id",
		"student_name",
		"assessment_id",
		"assessment_title",
		"attempt_id",
		"status",
		"score",
		"max_score",
		"submitted_at",
	})
	for _, e := range entries {
		_ = w.Write([]string{
			e.StudentUserID,
			e.StudentName,
			e.AssessmentID,
			e.AssessmentTitle,
			ptrString(e.AttemptID),
			ptrString(e.Status),
			ptrString(e.Score),
			ptrString(e.MaxScore),
			formatTime(e.SubmittedAt),
		})
	}
	w.Flush()
	return buf.Bytes()
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func ptrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
