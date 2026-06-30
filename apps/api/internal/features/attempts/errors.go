package attempts

import "errors"

// Service-level errors.
var (
	ErrAttemptNotFound      = errors.New("attempt not found")
	ErrAttemptNotInProgress = errors.New("attempt not in progress")
	ErrAttemptExpired       = errors.New("attempt expired")
	ErrAnswerItemNotFound   = errors.New("attempt item not found")
)
