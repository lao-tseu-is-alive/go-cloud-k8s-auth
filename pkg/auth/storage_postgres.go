package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

// PGX implements UserStorage using PostgreSQL via pgx.
type PGX struct {
	Conn *pgxpool.Pool
	dbi  database.DB
	log  *slog.Logger
}

// NewPgxDB creates a new PGX storage instance and verifies that the go_auth schema exists.
func NewPgxDB(ctx context.Context, db database.DB, log *slog.Logger) (UserStorage, error) {
	pgConn, err := db.GetPGConn()
	if err != nil {
		return nil, fmt.Errorf("NewPgxDB: failed to get PG connection: %w", err)
	}

	// Verify the go_auth schema exists by checking the users table
	var tableExists bool
	err = pgConn.QueryRow(ctx,
		`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'go_auth' AND table_name = 'users')`,
	).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("NewPgxDB: failed to check schema: %w", err)
	}
	if !tableExists {
		return nil, fmt.Errorf("NewPgxDB: go_auth.users table does not exist, run migrations first")
	}

	log.Info("NewPgxDB: connected to go_auth schema")
	return &PGX{Conn: pgConn, dbi: db, log: log}, nil
}

// UpsertByProvider creates or updates a user by OAuth provider.
func (db *PGX) UpsertByProvider(ctx context.Context, email, name, avatarURL, provider, providerID string) (*User, error) {
	db.log.Debug("UpsertByProvider", "email", email, "provider", provider, "providerID", providerID)
	res := &User{}
	err := pgxscan.Get(ctx, db.Conn, res, upsertUserByProvider, email, name, avatarURL, provider, providerID)
	if err != nil {
		db.log.Error("UpsertByProvider failed", "error", err)
		return nil, fmt.Errorf("UpsertByProvider: %w", err)
	}
	return res, nil
}

// GetByID returns a user by UUID.
func (db *PGX) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	db.log.Debug("GetByID", "id", id)
	res := &User{}
	err := pgxscan.Get(ctx, db.Conn, res, getUserByID, id)
	if err != nil {
		db.log.Error("GetByID failed", "id", id, "error", err)
		return nil, err
	}
	return res, nil
}

// GetByAlternateAppID returns a user by legacy integer ID.
func (db *PGX) GetByAlternateAppID(ctx context.Context, appID int64) (*User, error) {
	db.log.Debug("GetByAlternateAppID", "appID", appID)
	res := &User{}
	err := pgxscan.Get(ctx, db.Conn, res, getUserByAlternateAppID, appID)
	if err != nil {
		db.log.Error("GetByAlternateAppID failed", "appID", appID, "error", err)
		return nil, err
	}
	return res, nil
}

// List returns a paginated list of users.
func (db *PGX) List(ctx context.Context, offset, limit int, disabledFilter *bool) ([]*UserList, error) {
	db.log.Debug("List", "offset", offset, "limit", limit)
	var res []*UserList
	// Convert disabled filter to is_active filter (inverted logic)
	var isActiveFilter *bool
	if disabledFilter != nil {
		inverted := !*disabledFilter
		isActiveFilter = &inverted
	}
	err := pgxscan.Select(ctx, db.Conn, &res, listUsers, limit, offset, isActiveFilter)
	if err != nil {
		db.log.Error("List failed", "error", err)
		return nil, err
	}
	if res == nil {
		return make([]*UserList, 0), nil
	}
	return res, nil
}

// Create inserts a new user.
func (db *PGX) Create(ctx context.Context, u User) (*User, error) {
	db.log.Debug("Create", "email", u.Email)
	res := &User{}
	roles := u.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}
	err := pgxscan.Get(ctx, db.Conn, res, createUser,
		u.Email, u.Name, u.AvatarURL, u.Provider, u.ProviderID, roles, u.IsActive)
	if err != nil {
		db.log.Error("Create failed", "email", u.Email, "error", err)
		return nil, err
	}
	return res, nil
}

// Update modifies an existing user.
func (db *PGX) Update(ctx context.Context, id uuid.UUID, u User) (*User, error) {
	db.log.Debug("Update", "id", id)
	res := &User{}
	err := pgxscan.Get(ctx, db.Conn, res, updateUser,
		id, u.Email, u.Name, u.AvatarURL, u.Roles, u.IsActive)
	if err != nil {
		db.log.Error("Update failed", "id", id, "error", err)
		return nil, err
	}
	return res, nil
}

// Delete removes a user by UUID.
func (db *PGX) Delete(ctx context.Context, id uuid.UUID) error {
	db.log.Debug("Delete", "id", id)
	rowsAffected, err := db.dbi.ExecActionQuery(ctx, deleteUser, id)
	if err != nil {
		db.log.Error("Delete failed", "id", id, "error", err)
		return fmt.Errorf("Delete: %w", err)
	}
	if rowsAffected < 1 {
		return fmt.Errorf("Delete: no user found with id %v", id)
	}
	return nil
}

// Count returns the number of users.
func (db *PGX) Count(ctx context.Context, disabledFilter *bool) (int32, error) {
	db.log.Debug("Count")
	var isActiveFilter *bool
	if disabledFilter != nil {
		inverted := !*disabledFilter
		isActiveFilter = &inverted
	}
	count, err := db.dbi.GetQueryInt(ctx, countUsers, isActiveFilter)
	if err != nil {
		db.log.Error("Count failed", "error", err)
		return 0, err
	}
	return int32(count), nil
}

// Exist checks if a user exists by UUID.
func (db *PGX) Exist(ctx context.Context, id uuid.UUID) bool {
	count, err := db.dbi.GetQueryInt(ctx, existUser, id)
	if err != nil {
		db.log.Error("Exist failed", "id", id, "error", err)
		return false
	}
	return count > 0
}

// UpdateLastLogin updates the last_login_at timestamp.
func (db *PGX) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := db.dbi.ExecActionQuery(ctx, updateLastLogin, id)
	if err != nil {
		db.log.Error("UpdateLastLogin failed", "id", id, "error", err)
		return err
	}
	return nil
}

// GetUserGroupIDs returns the group UUIDs for the given user.
func (db *PGX) GetUserGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	db.log.Debug("GetUserGroupIDs", "userID", userID)
	var groupIDs []uuid.UUID
	rows, err := db.Conn.Query(ctx, getUserGroupIDs, userID)
	if err != nil {
		db.log.Error("GetUserGroupIDs query failed", "userID", userID, "error", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var gid uuid.UUID
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, gid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if groupIDs == nil {
		return make([]uuid.UUID, 0), nil
	}
	return groupIDs, nil
}
