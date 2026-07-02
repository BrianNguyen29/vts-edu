package notifications

import "errors"

// Common domain errors.
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("notification not found")
	ErrInvalidInput = errors.New("invalid input")
)
