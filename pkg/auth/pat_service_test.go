package auth

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePatStore implements PatStorage in memory.
type fakePatStore struct {
	byHash map[string]*PersonalAccessToken
}

func newFakePatStore() *fakePatStore {
	return &fakePatStore{byHash: make(map[string]*PersonalAccessToken)}
}

func (f *fakePatStore) Create(_ context.Context, userID uuid.UUID, tokenHash, prefix, name string, scopes []string, expiresAt *time.Time) (*PersonalAccessToken, error) {
	pat := &PersonalAccessToken{
		ID: uuid.New(), UserID: userID, TokenHash: tokenHash, Prefix: prefix,
		Name: name, Scopes: scopes, CreatedAt: time.Now(), ExpiresAt: expiresAt,
	}
	f.byHash[tokenHash] = pat
	return pat, nil
}

func (f *fakePatStore) GetByHash(_ context.Context, tokenHash string) (*PersonalAccessToken, error) {
	pat, ok := f.byHash[tokenHash]
	if !ok {
		return nil, ErrPatNotFound
	}
	return pat, nil
}

func (f *fakePatStore) ListByUser(_ context.Context, userID uuid.UUID) ([]*PersonalAccessToken, error) {
	var res []*PersonalAccessToken
	for _, p := range f.byHash {
		if p.UserID == userID {
			res = append(res, p)
		}
	}
	return res, nil
}

func (f *fakePatStore) Revoke(_ context.Context, id, userID uuid.UUID) error {
	for _, p := range f.byHash {
		if p.ID == id && p.UserID == userID && p.RevokedAt == nil {
			now := time.Now()
			p.RevokedAt = &now
			return nil
		}
	}
	return ErrPatNotFound
}

func (f *fakePatStore) TouchLastUsed(_ context.Context, _ uuid.UUID) error { return nil }

func newTestPatService(t *testing.T) (*PatBusinessService, *fakePatStore, *User) {
	t.Helper()
	log := golog.NewLogger("simple", io.Discard, golog.ErrorLevel, "pat_service_test")
	user := &User{
		ID:             uuid.New(),
		AlternateAppID: 4242,
		Email:          "jane@example.com",
		Name:           "Jane Doe",
		Roles:          []string{"user"},
		IsActive:       true,
	}
	users := &fakeUserStorage{usersByID: map[uuid.UUID]*User{user.ID: user}}
	pats := newFakePatStore()
	return NewPatBusinessService(pats, users, log), pats, user
}

func TestPatCreate(t *testing.T) {
	svc, _, user := newTestPatService(t)
	ctx := context.Background()

	t.Run("generates a pat_ token with display prefix and default scopes", func(t *testing.T) {
		token, pat, err := svc.Create(ctx, user.ID, "my mcp token", nil, 0)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(token, PatTokenPrefix))
		assert.Len(t, token, len(PatTokenPrefix)+patRandomChars)
		assert.Equal(t, token[:patDisplayPrefixLen], pat.Prefix)
		assert.Equal(t, DefaultPatScopes, pat.Scopes)
		assert.Nil(t, pat.ExpiresAt, "expires_in_days=0 means no expiry")
	})

	t.Run("sets expiry when requested", func(t *testing.T) {
		_, pat, err := svc.Create(ctx, user.ID, "expiring", nil, 30)
		require.NoError(t, err)
		require.NotNil(t, pat.ExpiresAt)
		assert.WithinDuration(t, time.Now().AddDate(0, 0, 30), *pat.ExpiresAt, time.Minute)
	})

	t.Run("rejects empty name and bad scopes", func(t *testing.T) {
		_, _, err := svc.Create(ctx, user.ID, "", nil, 0)
		assert.ErrorIs(t, err, ErrInvalidInput)
		_, _, err = svc.Create(ctx, user.ID, "x", []string{"NOT A SCOPE"}, 0)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})
}

func TestPatIntrospect(t *testing.T) {
	svc, pats, user := newTestPatService(t)
	ctx := context.Background()

	token, pat, err := svc.Create(ctx, user.ID, "mcp", []string{"notes:read", "notes:write"}, 0)
	require.NoError(t, err)

	t.Run("active token", func(t *testing.T) {
		res, err := svc.Introspect(ctx, token)
		require.NoError(t, err)
		assert.True(t, res.Active)
		assert.Equal(t, int64(4242), res.UserID)
		assert.Equal(t, "jane@example.com", res.Email)
		assert.Equal(t, []string{"notes:read", "notes:write"}, res.Scopes)
	})

	t.Run("unknown token is inactive", func(t *testing.T) {
		res, err := svc.Introspect(ctx, "pat_doesNotExistAtAll1234567890ab")
		require.NoError(t, err)
		assert.False(t, res.Active)
	})

	t.Run("non-pat token is inactive", func(t *testing.T) {
		res, err := svc.Introspect(ctx, "eyJhbGci.something.else")
		require.NoError(t, err)
		assert.False(t, res.Active)
	})

	t.Run("expired token is inactive", func(t *testing.T) {
		expToken, expPat, err := svc.Create(ctx, user.ID, "old", nil, 1)
		require.NoError(t, err)
		past := time.Now().Add(-time.Hour)
		expPat.ExpiresAt = &past
		res, err := svc.Introspect(ctx, expToken)
		require.NoError(t, err)
		assert.False(t, res.Active)
	})

	t.Run("revoked token is inactive", func(t *testing.T) {
		require.NoError(t, pats.Revoke(ctx, pat.ID, user.ID))
		res, err := svc.Introspect(ctx, token)
		require.NoError(t, err)
		assert.False(t, res.Active)
	})

	t.Run("disabled owner makes token inactive", func(t *testing.T) {
		user.IsActive = false
		t.Cleanup(func() { user.IsActive = true })
		freshToken, _, err := svc.Create(ctx, user.ID, "fresh", nil, 0)
		require.NoError(t, err)
		res, err := svc.Introspect(ctx, freshToken)
		require.NoError(t, err)
		assert.False(t, res.Active)
	})
}

func TestPatRevokeWrongOwner(t *testing.T) {
	svc, _, user := newTestPatService(t)
	ctx := context.Background()
	_, pat, err := svc.Create(ctx, user.ID, "mine", nil, 0)
	require.NoError(t, err)

	err = svc.Revoke(ctx, pat.ID, uuid.New())
	assert.ErrorIs(t, err, ErrPatNotFound, "revoking another user's PAT must look like not-found")
}
