// Package auth provides the authentication and user management service for go-cloud-k8s-auth.
package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the go_auth schema.
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	AlternateAppID int64      `json:"alternate_app_id" db:"alternate_app_id"`
	Email          string     `json:"email" db:"email"`
	Name           string     `json:"name" db:"name"`
	AvatarURL      string     `json:"avatar_url" db:"avatar_url"`
	Provider       string     `json:"provider" db:"provider"`
	ProviderID     string     `json:"provider_id" db:"provider_id"`
	Roles          []string   `json:"roles" db:"roles"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      *time.Time `json:"created_at" db:"created_at"`
	LastLoginAt    *time.Time `json:"last_login_at" db:"last_login_at"`
}

// UserList is a lightweight version of User for list endpoints.
type UserList struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	AlternateAppID int64      `json:"alternate_app_id" db:"alternate_app_id"`
	Email          string     `json:"email" db:"email"`
	Name           string     `json:"name" db:"name"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      *time.Time `json:"created_at" db:"created_at"`
	LastLoginAt    *time.Time `json:"last_login_at" db:"last_login_at"`
}

// Group represents a named collection of users.
type Group struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	OwnerID    uuid.UUID  `json:"owner_id" db:"owner_id"`
	IsPersonal bool       `json:"is_personal" db:"is_personal"`
	CreatedAt  *time.Time `json:"created_at" db:"created_at"`
}

// UserGroup represents the many-to-many relationship between users and groups.
type UserGroup struct {
	UserID  uuid.UUID `json:"user_id" db:"user_id"`
	GroupID uuid.UUID `json:"group_id" db:"group_id"`
	Role    string    `json:"role" db:"role"` // owner, admin, member
}

// OAuthUserInfo holds the user profile information returned by an OAuth provider.
type OAuthUserInfo struct {
	Provider   string
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}
