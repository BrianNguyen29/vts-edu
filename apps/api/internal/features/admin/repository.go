package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for the admin feature.
type Repository interface {
	ListUsers(ctx context.Context, orgID string) ([]User, error)
	LoginExists(ctx context.Context, orgID, loginName string) (bool, error)
	CreateUser(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error)
	GetMembershipID(ctx context.Context, orgID, userID string) (string, error)
	ReplaceRoles(ctx context.Context, tx pgx.Tx, membershipID string, roles []string) error
	BumpAuthVersion(ctx context.Context, tx pgx.Tx, userID string) error
	ResetPassword(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error
	RevokeUserSessions(ctx context.Context, tx pgx.Tx, userID string) error
	GetOrganization(ctx context.Context, orgID string) (Organization, error)
	UpdateOrganization(ctx context.Context, tx pgx.Tx, orgID, name string) error
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new admin repository backed by a pgx pool.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) ListUsers(ctx context.Context, orgID string) ([]User, error) {
	query := `
		SELECT
			u.id,
			u.display_name,
			u.email,
			ln.username_normalized,
			u.must_change_password,
			array_agg(mr.role) FILTER (WHERE mr.role IS NOT NULL)
		FROM users u
		JOIN organization_memberships m
			ON m.user_id = u.id
		JOIN membership_login_names ln
			ON ln.user_id = u.id AND ln.organization_id = m.organization_id
		LEFT JOIN membership_roles mr
			ON mr.membership_id = m.id
		WHERE m.organization_id = $1
		  AND m.status = 'ACTIVE'
		  AND ln.status = 'ACTIVE'
		GROUP BY u.id, u.display_name, u.email, ln.username_normalized, u.must_change_password
		ORDER BY ln.username_normalized
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID,
			&u.DisplayName,
			&u.Email,
			&u.LoginName,
			&u.MustChangePassword,
			&u.Roles,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		if u.Roles == nil {
			u.Roles = []string{}
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	if users == nil {
		users = []User{}
	}
	return users, nil
}

func (r *repository) LoginExists(ctx context.Context, orgID, loginName string) (bool, error) {
	query := `
		SELECT 1
		FROM membership_login_names
		WHERE organization_id = $1
		  AND lower(username_normalized) = $2
		LIMIT 1
	`

	var n int
	err := r.pool.QueryRow(ctx, query, orgID, strings.ToLower(loginName)).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check login exists: %w", err)
	}
	return true, nil
}

func (r *repository) CreateUser(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error) {
	var userID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (display_name, email, must_change_password)
		VALUES ($1, $2, true)
		RETURNING id
	`, displayName, email).Scan(&userID); err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}

	var membershipID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO organization_memberships (organization_id, user_id)
		VALUES ($1, $2)
		RETURNING id
	`, orgID, userID).Scan(&membershipID); err != nil {
		return User{}, fmt.Errorf("insert membership: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO membership_login_names (organization_id, username_normalized, user_id, password_hash)
		VALUES ($1, $2, $3, $4)
	`, orgID, strings.ToLower(loginName), userID, passwordHash); err != nil {
		return User{}, fmt.Errorf("insert login name: %w", err)
	}

	for _, role := range roles {
		if _, err := tx.Exec(ctx, `
			INSERT INTO membership_roles (membership_id, role)
			VALUES ($1, $2)
		`, membershipID, role); err != nil {
			return User{}, fmt.Errorf("insert role: %w", err)
		}
	}

	return User{
		ID:                 userID,
		DisplayName:        displayName,
		Email:              email,
		LoginName:          loginName,
		Roles:              roles,
		MustChangePassword: true,
	}, nil
}

func (r *repository) GetMembershipID(ctx context.Context, orgID, userID string) (string, error) {
	query := `
		SELECT id
		FROM organization_memberships
		WHERE organization_id = $1
		  AND user_id = $2
		  AND status = 'ACTIVE'
		LIMIT 1
	`

	var id string
	err := r.pool.QueryRow(ctx, query, orgID, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get membership: %w", err)
	}
	return id, nil
}

func (r *repository) ReplaceRoles(ctx context.Context, tx pgx.Tx, membershipID string, roles []string) error {
	if _, err := tx.Exec(ctx, `
		DELETE FROM membership_roles
		WHERE membership_id = $1
	`, membershipID); err != nil {
		return fmt.Errorf("delete roles: %w", err)
	}

	for _, role := range roles {
		if _, err := tx.Exec(ctx, `
			INSERT INTO membership_roles (membership_id, role)
			VALUES ($1, $2)
		`, membershipID, role); err != nil {
			return fmt.Errorf("insert role: %w", err)
		}
	}
	return nil
}

func (r *repository) BumpAuthVersion(ctx context.Context, tx pgx.Tx, userID string) error {
	_, err := tx.Exec(ctx, `
		UPDATE users
		SET auth_version = auth_version + 1
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("bump auth version: %w", err)
	}
	return nil
}

func (r *repository) ResetPassword(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error {
	tag, err := tx.Exec(ctx, `
		UPDATE membership_login_names
		SET password_hash = $1
		WHERE user_id = $2
		  AND organization_id = $3
	`, passwordHash, userID, orgID)
	if err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET must_change_password = true,
		    auth_version = auth_version + 1
		WHERE id = $1
	`, userID); err != nil {
		return fmt.Errorf("update user flags: %w", err)
	}
	return nil
}

func (r *repository) RevokeUserSessions(ctx context.Context, tx pgx.Tx, userID string) error {
	_, err := tx.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = now()
		WHERE user_id = $1
		  AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

func (r *repository) GetOrganization(ctx context.Context, orgID string) (Organization, error) {
	query := `
		SELECT id, code, name
		FROM organizations
		WHERE id = $1
		LIMIT 1
	`

	var o Organization
	err := r.pool.QueryRow(ctx, query, orgID).Scan(&o.ID, &o.Code, &o.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return Organization{}, ErrOrganizationNotFound
	}
	if err != nil {
		return Organization{}, fmt.Errorf("get organization: %w", err)
	}
	return o, nil
}

func (r *repository) UpdateOrganization(ctx context.Context, tx pgx.Tx, orgID, name string) error {
	tag, err := tx.Exec(ctx, `
		UPDATE organizations
		SET name = $1
		WHERE id = $2
	`, name, orgID)
	if err != nil {
		return fmt.Errorf("update organization: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}
