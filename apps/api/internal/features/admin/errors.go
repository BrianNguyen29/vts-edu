package admin

import "errors"

// Service errors.
var (
	ErrUserNotFound         = errors.New("user not found")
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrDuplicateLogin       = errors.New("login name already exists")
	ErrInvalidInput         = errors.New("invalid input")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInvalidCursor        = errors.New("invalid cursor")
)
