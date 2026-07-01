package auth

import "time"

// LoginRequest is the payload for POST /api/v1/auth/login.
type LoginRequest struct {
	OrganizationCode string `json:"organization_code"`
	Username         string `json:"username"`
	Password         string `json:"password"`
}

// UserInfo is a minimal public user shape returned on login.
type UserInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// LoginResponse is wrapped inside the response envelope on login.
type LoginResponse struct {
	AccessToken        string   `json:"access_token"`
	ExpiresIn          int      `json:"expires_in"`
	User               UserInfo `json:"user"`
	Roles              []string `json:"roles"`
	Permissions        []string `json:"permissions"`
	MustChangePassword bool     `json:"must_change_password"`
}

// MeResponse is wrapped inside the response envelope for GET /api/v1/me.
type MeResponse struct {
	ID                 string   `json:"id"`
	OrganizationID     string   `json:"organization_id"`
	DisplayName        string   `json:"display_name"`
	Roles              []string `json:"roles"`
	Permissions        []string `json:"permissions"`
	MustChangePassword bool     `json:"must_change_password"`
}

// DataEnvelope wraps successful API responses.
type DataEnvelope struct {
	Data any `json:"data"`
}

// ErrorEnvelope wraps API error responses.
type ErrorEnvelope struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

// LoginResult is the service-level result of a successful login.
type LoginResult struct {
	AccessToken        string
	ExpiresIn          int
	RefreshToken       string
	RefreshExpires     time.Time
	User               UserInfo
	Roles              []string
	Permissions        []string
	MustChangePassword bool
}

// RefreshResult is the service-level result of a successful refresh.
type RefreshResult struct {
	AccessToken        string
	ExpiresIn          int
	RefreshToken       string
	RefreshExpires     time.Time
	User               UserInfo
	Roles              []string
	Permissions        []string
	MustChangePassword bool
}

// LogoutResponse is wrapped inside the response envelope on logout.
type LogoutResponse struct {
	Success bool `json:"success"`
}

// LogoutResult is the service-level result of a logout.
type LogoutResult struct {
	Success bool
}

// MeResult is the service-level result for GET /api/v1/me.
type MeResult struct {
	ID                 string
	OrganizationID     string
	DisplayName        string
	Roles              []string
	Permissions        []string
	MustChangePassword bool
}

// ChangePasswordRequest is the payload for POST /api/v1/auth/change-password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePasswordResponse is wrapped inside the response envelope on password change.
type ChangePasswordResponse struct {
	Success bool `json:"success"`
}
