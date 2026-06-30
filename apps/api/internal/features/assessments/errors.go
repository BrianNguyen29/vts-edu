package assessments

import "errors"

// ErrInvalidCursor is returned when a cursor string cannot be decoded.
var ErrInvalidCursor = errors.New("invalid cursor")
