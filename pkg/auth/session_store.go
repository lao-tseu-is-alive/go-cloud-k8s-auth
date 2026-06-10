package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

// ErrSessionNotFound is returned when no valid (existing, unexpired, unrevoked) session matches.
var ErrSessionNotFound = errors.New("session not found or expired")

// Session represents a browser SSO session backed by the go_auth.sessions table.
type Session struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TokenHash  string     `json:"-" db:"token_hash"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	LastSeenAt *time.Time `json:"last_seen_at" db:"last_seen_at"`
	UserAgent  string     `json:"user_agent" db:"user_agent"`
	IP         string     `json:"ip" db:"ip"`
	RevokedAt  *time.Time `json:"revoked_at" db:"revoked_at"`
}

// SessionStorage defines the persistence interface for browser SSO sessions.
type SessionStorage interface {
	// Create stores a new session for the user and returns the raw opaque token
	// destined for the cookie. Only its SHA-256 hash is persisted.
	Create(ctx context.Context, userID uuid.UUID, userAgent, ip string, ttl time.Duration) (string, error)

	// GetValid returns the session matching rawToken if it exists, is not expired
	// and not revoked, bumping last_seen_at. Returns ErrSessionNotFound otherwise.
	GetValid(ctx context.Context, rawToken string) (*Session, error)

	// Revoke marks the session matching rawToken as revoked. Revoking an unknown
	// token is not an error.
	Revoke(ctx context.Context, rawToken string) error

	// RevokeAllForUser revokes every active session of the given user.
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

const (
	insertSession = `
INSERT INTO go_auth.sessions (token_hash, user_id, expires_at, user_agent, ip)
VALUES ($1, $2, $3, $4, $5);
`

	getValidSession = `
UPDATE go_auth.sessions SET last_seen_at = NOW()
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
RETURNING id, token_hash, user_id, created_at, expires_at, last_seen_at, user_agent, ip, revoked_at;
`

	revokeSession = `
UPDATE go_auth.sessions SET revoked_at = NOW()
WHERE token_hash = $1 AND revoked_at IS NULL;
`

	revokeAllUserSessions = `
UPDATE go_auth.sessions SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;
`
)

// PgxSessionStore implements SessionStorage using PostgreSQL via pgx.
type PgxSessionStore struct {
	Conn *pgxpool.Pool
	log  *slog.Logger
}

// NewPgxSessionStore creates a new PgxSessionStore and verifies the sessions table exists.
func NewPgxSessionStore(ctx context.Context, db database.DB, log *slog.Logger) (SessionStorage, error) {
	pgConn, err := db.GetPGConn()
	if err != nil {
		return nil, fmt.Errorf("NewPgxSessionStore: failed to get PG connection: %w", err)
	}
	var tableExists bool
	err = pgConn.QueryRow(ctx,
		`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'go_auth' AND table_name = 'sessions')`,
	).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("NewPgxSessionStore: failed to check schema: %w", err)
	}
	if !tableExists {
		return nil, fmt.Errorf("NewPgxSessionStore: go_auth.sessions table does not exist, run migrations first")
	}
	log.Info("NewPgxSessionStore: connected to go_auth.sessions")
	return &PgxSessionStore{Conn: pgConn, log: log}, nil
}

// newSessionToken generates a 256-bit random opaque token and its storage hash.
func newSessionToken() (rawToken, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("newSessionToken: %w", err)
	}
	rawToken = base64.RawURLEncoding.EncodeToString(b)
	return rawToken, hashSessionToken(rawToken), nil
}

// hashSessionToken returns the sha256 hex digest of a raw session token.
func hashSessionToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

// Create stores a new session and returns the raw cookie token.
func (s *PgxSessionStore) Create(ctx context.Context, userID uuid.UUID, userAgent, ip string, ttl time.Duration) (string, error) {
	rawToken, tokenHash, err := newSessionToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(ttl)
	if _, err = s.Conn.Exec(ctx, insertSession, tokenHash, userID, expiresAt, userAgent, ip); err != nil {
		s.log.Error("Create session failed", "userID", userID, "error", err)
		return "", fmt.Errorf("Create session: %w", err)
	}
	s.log.Info("session created", "userID", userID, "expiresAt", expiresAt)
	return rawToken, nil
}

// GetValid returns the session for rawToken if still valid, bumping last_seen_at.
func (s *PgxSessionStore) GetValid(ctx context.Context, rawToken string) (*Session, error) {
	res := &Session{}
	err := pgxscan.Get(ctx, s.Conn, res, getValidSession, hashSessionToken(rawToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		s.log.Error("GetValid session failed", "error", err)
		return nil, fmt.Errorf("GetValid session: %w", err)
	}
	return res, nil
}

// Revoke marks the session matching rawToken as revoked.
func (s *PgxSessionStore) Revoke(ctx context.Context, rawToken string) error {
	if _, err := s.Conn.Exec(ctx, revokeSession, hashSessionToken(rawToken)); err != nil {
		s.log.Error("Revoke session failed", "error", err)
		return fmt.Errorf("Revoke session: %w", err)
	}
	return nil
}

// RevokeAllForUser revokes every active session of the given user.
func (s *PgxSessionStore) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	if _, err := s.Conn.Exec(ctx, revokeAllUserSessions, userID); err != nil {
		s.log.Error("RevokeAllForUser failed", "userID", userID, "error", err)
		return fmt.Errorf("RevokeAllForUser: %w", err)
	}
	return nil
}
