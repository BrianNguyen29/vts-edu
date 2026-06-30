package assessments

import "errors"

// ErrInvalidCursor is returned when a cursor string cannot be decoded.
var ErrInvalidCursor = errors.New("invalid cursor")

// Builder errors.
var (
	ErrNotFound         = errors.New("assessment not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidInput     = errors.New("invalid input")
	ErrNotDraft         = errors.New("assessment is not in draft status")
	ErrValidationFailed = errors.New("validation failed")
	ErrDuplicateTarget  = errors.New("target already exists")
	ErrDuplicateSection = errors.New("section position already exists")
	ErrDuplicateItem    = errors.New("item position already exists")
)
