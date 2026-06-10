package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	// PatTokenPrefix marks a bearer token as a personal access token, letting
	// downstream services route it to introspection instead of JWT parsing.
	PatTokenPrefix = "pat_"
	patRandomChars = 32
	patBase62      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// patDisplayPrefixLen is how many characters of the token are kept for display.
	patDisplayPrefixLen = 12
)

// DefaultPatScopes are granted when a creation request carries no scopes.
var DefaultPatScopes = []string{"notes:read", "notes:write", "notes:mcp"}

// patScopePattern validates scope strings like "notes:read".
var patScopePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*:[a-z][a-z0-9-]*$`)

// PatIntrospection is the result of a successful token introspection.
type PatIntrospection struct {
	Active    bool
	UserID    int64 // alternate_app_id, what modules use as owner id
	Email     string
	Name      string
	Scopes    []string
	ExpiresAt *time.Time
}

// PatBusinessService handles creation, listing, revocation and introspection
// of personal access tokens.
type PatBusinessService struct {
	Pats  PatStorage
	Users UserStorage
	Log   *slog.Logger
}

// NewPatBusinessService creates a new PatBusinessService.
func NewPatBusinessService(pats PatStorage, users UserStorage, log *slog.Logger) *PatBusinessService {
	return &PatBusinessService{Pats: pats, Users: users, Log: log}
}

// generatePatToken returns a new random "pat_..." token value.
func generatePatToken() (string, error) {
	token := make([]byte, 0, len(PatTokenPrefix)+patRandomChars)
	token = append(token, PatTokenPrefix...)
	max := big.NewInt(int64(len(patBase62)))
	for i := 0; i < patRandomChars; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("generatePatToken: %w", err)
		}
		token = append(token, patBase62[n.Int64()])
	}
	return string(token), nil
}

// HashPatToken returns the sha256 hex digest used to store and look up tokens.
func HashPatToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// validatePatScopes checks scope syntax, returning the normalized list.
func validatePatScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return append([]string(nil), DefaultPatScopes...), nil
	}
	for _, s := range scopes {
		if !patScopePattern.MatchString(s) {
			return nil, fmt.Errorf("%w: invalid scope %q (expected form module:action)", ErrInvalidInput, s)
		}
	}
	return scopes, nil
}

// Create generates and stores a new PAT for the user. The returned string is
// the full token value, shown to the user exactly once.
func (s *PatBusinessService) Create(ctx context.Context, userID uuid.UUID, name string, scopes []string, expiresInDays int32) (string, *PersonalAccessToken, error) {
	if name == "" || len(name) > 100 {
		return "", nil, fmt.Errorf("%w: name must be 1-100 characters", ErrInvalidInput)
	}
	if expiresInDays < 0 {
		return "", nil, fmt.Errorf("%w: expires_in_days cannot be negative", ErrInvalidInput)
	}
	normalizedScopes, err := validatePatScopes(scopes)
	if err != nil {
		return "", nil, err
	}

	token, err := generatePatToken()
	if err != nil {
		return "", nil, err
	}
	var expiresAt *time.Time
	if expiresInDays > 0 {
		t := time.Now().AddDate(0, 0, int(expiresInDays))
		expiresAt = &t
	}

	pat, err := s.Pats.Create(ctx, userID, HashPatToken(token), token[:patDisplayPrefixLen], name, normalizedScopes, expiresAt)
	if err != nil {
		return "", nil, err
	}
	s.Log.Info("PAT created", "userID", userID, "patID", pat.ID, "name", name, "scopes", normalizedScopes)
	return token, pat, nil
}

// List returns the user's PATs (metadata only).
func (s *PatBusinessService) List(ctx context.Context, userID uuid.UUID) ([]*PersonalAccessToken, error) {
	return s.Pats.ListByUser(ctx, userID)
}

// Revoke revokes one of the user's PATs.
func (s *PatBusinessService) Revoke(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.Pats.Revoke(ctx, id, userID); err != nil {
		return err
	}
	s.Log.Info("PAT revoked", "userID", userID, "patID", id)
	return nil
}

// Introspect verifies a raw "pat_..." token and returns the attached identity
// and scopes. An unknown, revoked or expired token (or a disabled owner)
// yields Active=false without error, so callers can return RFC 7662 style
// responses without leaking why the token is inactive.
func (s *PatBusinessService) Introspect(ctx context.Context, token string) (*PatIntrospection, error) {
	inactive := &PatIntrospection{Active: false}
	if len(token) < len(PatTokenPrefix) || token[:len(PatTokenPrefix)] != PatTokenPrefix {
		return inactive, nil
	}

	pat, err := s.Pats.GetByHash(ctx, HashPatToken(token))
	if err != nil {
		if err == ErrPatNotFound {
			return inactive, nil
		}
		return nil, err
	}
	if pat.RevokedAt != nil {
		return inactive, nil
	}
	if pat.ExpiresAt != nil && time.Now().After(*pat.ExpiresAt) {
		return inactive, nil
	}

	user, err := s.Users.GetByID(ctx, pat.UserID)
	if err != nil {
		s.Log.Warn("Introspect: PAT owner lookup failed", "patID", pat.ID, "error", err)
		return inactive, nil
	}
	if !user.IsActive {
		return inactive, nil
	}

	// Best-effort usage tracking; never block the request on it.
	go func(id uuid.UUID) {
		touchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Pats.TouchLastUsed(touchCtx, id); err != nil {
			s.Log.Warn("Introspect: TouchLastUsed failed", "patID", id, "error", err)
		}
	}(pat.ID)

	return &PatIntrospection{
		Active:    true,
		UserID:    user.AlternateAppID,
		Email:     user.Email,
		Name:      user.Name,
		Scopes:    pat.Scopes,
		ExpiresAt: pat.ExpiresAt,
	}, nil
}
