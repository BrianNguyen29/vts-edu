package admin

// User is the public admin view of a user within an organization.
type User struct {
	ID                 string   `json:"id"`
	DisplayName        string   `json:"display_name"`
	Email              string   `json:"email"`
	LoginName          string   `json:"login_name"`
	Roles              []string `json:"roles"`
	MustChangePassword bool     `json:"must_change_password"`
}

// CreateUserRequest is the payload for POST /api/v1/users.
type CreateUserRequest struct {
	LoginName         string   `json:"login_name"`
	DisplayName       string   `json:"display_name"`
	Email             string   `json:"email"`
	TemporaryPassword string   `json:"temporary_password"`
	Roles             []string `json:"roles"`
}

// UpdateRolesRequest is the payload for PUT /api/v1/users/{user_id}/roles.
type UpdateRolesRequest struct {
	Roles []string `json:"roles"`
}

// ResetPasswordRequest is the payload for POST /api/v1/users/{user_id}/reset-password.
type ResetPasswordRequest struct {
	TemporaryPassword string `json:"temporary_password"`
}

// Organization is the public view of the current organization.
type Organization struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// UpdateOrganizationRequest is the payload for PATCH /api/v1/organizations/current.
type UpdateOrganizationRequest struct {
	Name string `json:"name"`
}

// AuditLogParams is the persistence input for an audit log row.
type AuditLogParams struct {
	OrganizationID string
	ActorUserID    string
	Action         string
	ResourceType   string
	ResourceID     string
	BeforeJSON     []byte
	AfterJSON      []byte
	MetadataJSON   []byte
}

// ListOptions is the optional pagination/search input for list endpoints.
type ListOptions struct {
	Query  string
	Limit  int
	Offset int
}

// PageInfo is returned with paginated list responses.
type PageInfo struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// DataEnvelope wraps successful API responses.
type DataEnvelope struct {
	Data any       `json:"data"`
	Page *PageInfo `json:"page,omitempty"`
}

// ErrorEnvelope wraps API error responses.
type ErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
