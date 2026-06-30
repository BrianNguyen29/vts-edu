package academics

import "errors"

// Service errors.
var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrNotFound            = errors.New("not found")
	ErrDuplicateCode       = errors.New("code already exists")
	ErrDuplicateEnrollment = errors.New("student already enrolled")
	ErrDuplicateTeacher    = errors.New("teacher already assigned")
	ErrUserNotFound        = errors.New("user not found")
	ErrClassNotFound       = errors.New("class not found")
)
