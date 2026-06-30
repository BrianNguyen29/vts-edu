package attempts

import "errors"

// Service-level errors.
var (
	ErrAttemptNotFound       = errors.New("attempt not found")
	ErrAttemptNotInProgress  = errors.New("attempt not in progress")
	ErrAttemptExpired        = errors.New("attempt expired")
	ErrAnswerItemNotFound    = errors.New("attempt item not found")
	ErrAssessmentNotFound    = errors.New("assessment not found")
	ErrNoPublication         = errors.New("assessment has no published version")
	ErrAssessmentUnavailable = errors.New("assessment is not currently available")
	ErrAttemptLimitReached   = errors.New("maximum number of attempts reached")
	ErrNotAssigned           = errors.New("assessment is not assigned to this student")
)
