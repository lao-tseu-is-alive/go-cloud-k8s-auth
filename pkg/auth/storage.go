package auth

import (
	"context"

	"github.com/google/uuid"
)

// UserStorage defines the persistence interface for user operations.
type UserStorage interface {
	// UpsertByProvider creates or updates a user based on OAuth provider + provider ID.
	// On conflict (provider, provider_id), updates name, avatar_url, and last_login_at.
	UpsertByProvider(ctx context.Context, email, name, avatarURL, provider, providerID string) (*User, error)

	// GetByID returns the user with the specified UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByAlternateAppID returns the user with the specified legacy integer ID.
	GetByAlternateAppID(ctx context.Context, appID int64) (*User, error)

	// List returns users with pagination and optional active/disabled filter.
	// If disabledFilter is nil, all users are returned.
	List(ctx context.Context, offset, limit int, disabledFilter *bool) ([]*UserList, error)

	// Create saves a new user in the storage.
	Create(ctx context.Context, u User) (*User, error)

	// Update updates the user with given UUID in the storage.
	Update(ctx context.Context, id uuid.UUID, u User) (*User, error)

	// Delete removes the user with given UUID from the storage.
	Delete(ctx context.Context, id uuid.UUID) error

	// Count returns the total number of users, optionally filtered by active status.
	Count(ctx context.Context, disabledFilter *bool) (int32, error)

	// Exist returns true if a user with the specified UUID exists.
	Exist(ctx context.Context, id uuid.UUID) bool

	// UpdateLastLogin updates the last_login_at timestamp for the given user.
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// GetUserGroupIDs returns the group UUIDs for the given user.
	GetUserGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
