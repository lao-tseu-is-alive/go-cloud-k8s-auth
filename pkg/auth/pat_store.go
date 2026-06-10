package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

// ErrPatNotFound is returned when no PAT matches the given hash or id.
var ErrPatNotFound = errors.New("personal access token not found")

// PersonalAccessToken represents a row of go_auth.personal_access_tokens.
type PersonalAccessToken struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash  string     `json:"-" db:"token_hash"`
	Prefix     string     `json:"prefix" db:"prefix"`
	Name       string     `json:"name" db:"name"`
	Scopes     []string   `json:"scopes" db:"scopes"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at" db:"last_used_at"`
	RevokedAt  *time.Time `json:"revoked_at" db:"revoked_at"`
}

// PatStorage defines the persistence interface for personal access tokens.
type PatStorage interface {
	// Create stores a new PAT (already hashed) and returns the created row.
	Create(ctx context.Context, userID uuid.UUID, tokenHash, prefix, name string, scopes []string, expiresAt *time.Time) (*PersonalAccessToken, error)

	// GetByHash returns the PAT matching the given sha256 hex hash regardless
	// of its revoked/expired status (callers decide); ErrPatNotFound otherwise.
	GetByHash(ctx context.Context, tokenHash string) (*PersonalAccessToken, error)

	// ListByUser returns all PATs of a user, newest first.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*PersonalAccessToken, error)

	// Revoke marks the PAT as revoked; the userID must match the owner.
	Revoke(ctx context.Context, id, userID uuid.UUID) error

	// TouchLastUsed updates last_used_at for the given PAT.
	TouchLastUsed(ctx context.Context, id uuid.UUID) error
}

const (
	insertPat = `
INSERT INTO go_auth.personal_access_tokens (user_id, token_hash, prefix, name, scopes, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, token_hash, prefix, name, scopes, created_at, expires_at, last_used_at, revoked_at;
`

	getPatByHash = `
SELECT id, user_id, token_hash, prefix, name, scopes, created_at, expires_at, last_used_at, revoked_at
FROM go_auth.personal_access_tokens
WHERE token_hash = $1;
`

	listPatsByUser = `
SELECT id, user_id, token_hash, prefix, name, scopes, created_at, expires_at, last_used_at, revoked_at
FROM go_auth.personal_access_tokens
WHERE user_id = $1
ORDER BY created_at DESC;
`

	revokePat = `
UPDATE go_auth.personal_access_tokens SET revoked_at = NOW()
WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL;
`

	touchPatLastUsed = `
UPDATE go_auth.personal_access_tokens SET last_used_at = NOW() WHERE id = $1;
`
)

// PgxPatStore implements PatStorage using PostgreSQL via pgx.
type PgxPatStore struct {
	Conn *pgxpool.Pool
	log  *slog.Logger
}

// NewPgxPatStore creates a new PgxPatStore and verifies the table exists.
func NewPgxPatStore(ctx context.Context, db database.DB, log *slog.Logger) (PatStorage, error) {
	pgConn, err := db.GetPGConn()
	if err != nil {
		return nil, fmt.Errorf("NewPgxPatStore: failed to get PG connection: %w", err)
	}
	var tableExists bool
	err = pgConn.QueryRow(ctx,
		`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'go_auth' AND table_name = 'personal_access_tokens')`,
	).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("NewPgxPatStore: failed to check schema: %w", err)
	}
	if !tableExists {
		return nil, fmt.Errorf("NewPgxPatStore: go_auth.personal_access_tokens table does not exist, run migrations first")
	}
	log.Info("NewPgxPatStore: connected to go_auth.personal_access_tokens")
	return &PgxPatStore{Conn: pgConn, log: log}, nil
}

// Create stores a new PAT row.
func (s *PgxPatStore) Create(ctx context.Context, userID uuid.UUID, tokenHash, prefix, name string, scopes []string, expiresAt *time.Time) (*PersonalAccessToken, error) {
	res := &PersonalAccessToken{}
	err := pgxscan.Get(ctx, s.Conn, res, insertPat, userID, tokenHash, prefix, name, scopes, expiresAt)
	if err != nil {
		s.log.Error("Create PAT failed", "userID", userID, "error", err)
		return nil, fmt.Errorf("Create PAT: %w", err)
	}
	return res, nil
}

// GetByHash returns the PAT matching the given hash.
func (s *PgxPatStore) GetByHash(ctx context.Context, tokenHash string) (*PersonalAccessToken, error) {
	res := &PersonalAccessToken{}
	err := pgxscan.Get(ctx, s.Conn, res, getPatByHash, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPatNotFound
		}
		s.log.Error("GetByHash PAT failed", "error", err)
		return nil, fmt.Errorf("GetByHash PAT: %w", err)
	}
	return res, nil
}

// ListByUser returns all PATs of a user, newest first.
func (s *PgxPatStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]*PersonalAccessToken, error) {
	var res []*PersonalAccessToken
	err := pgxscan.Select(ctx, s.Conn, &res, listPatsByUser, userID)
	if err != nil {
		s.log.Error("ListByUser PAT failed", "userID", userID, "error", err)
		return nil, fmt.Errorf("ListByUser PAT: %w", err)
	}
	if res == nil {
		return make([]*PersonalAccessToken, 0), nil
	}
	return res, nil
}

// Revoke marks the PAT as revoked if owned by userID.
func (s *PgxPatStore) Revoke(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := s.Conn.Exec(ctx, revokePat, id, userID)
	if err != nil {
		s.log.Error("Revoke PAT failed", "id", id, "error", err)
		return fmt.Errorf("Revoke PAT: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPatNotFound
	}
	return nil
}

// TouchLastUsed updates last_used_at for the given PAT.
func (s *PgxPatStore) TouchLastUsed(ctx context.Context, id uuid.UUID) error {
	if _, err := s.Conn.Exec(ctx, touchPatLastUsed, id); err != nil {
		s.log.Error("TouchLastUsed PAT failed", "id", id, "error", err)
		return fmt.Errorf("TouchLastUsed PAT: %w", err)
	}
	return nil
}
