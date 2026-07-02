package grading

import "errors"

// Sentinels for grading-domain errors.
var (
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrInvalidScore       = errors.New("awarded_score is invalid")
	ErrScoreExceedsPoints = errors.New("awarded_score exceeds item points")
	ErrNotGradeable       = errors.New("item is not manually gradable")
	ErrItemNotInAttempt   = errors.New("item does not belong to attempt")
)
