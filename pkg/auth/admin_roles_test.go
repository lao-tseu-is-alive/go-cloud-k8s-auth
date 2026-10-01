package auth

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1"
)

// AddRole grants the role in memory, mirroring the "only when absent" SQL.
func (f *fakeUserStorage) AddRole(_ context.Context, id uuid.UUID, role string) (*User, error) {
	u, ok := f.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	if slices.Contains(u.Roles, role) {
		return nil, nil
	}
	updated := *u
	updated.Roles = append(slices.Clone(u.Roles), role)
	f.usersByID[id] = &updated
	return &updated, nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestParseAdminEmails(t *testing.T) {
	emails := ParseAdminEmails(" Alice@Example.org, ,bob@example.org,alice@example.org ")
	assert.Len(t, emails, 2)
	assert.Contains(t, emails, "alice@example.org")
	assert.Contains(t, emails, "bob@example.org")
	assert.Empty(t, ParseAdminEmails(""))
}

func TestEnsureBootstrapAdmin(t *testing.T) {
	listed := &User{ID: uuid.New(), Email: "Boss@Example.org", Roles: []string{"user"}}
	other := &User{ID: uuid.New(), Email: "someone@example.org", Roles: []string{"user"}}
	store := &fakeUserStorage{usersByID: map[uuid.UUID]*User{listed.ID: listed, other.ID: other}}
	svc := &AuthBusinessService{Store: store, Log: discardLogger(), BootstrapAdminEmails: ParseAdminEmails("boss@example.org")}

	got, err := svc.ensureBootstrapAdmin(context.Background(), listed)
	require.NoError(t, err)
	assert.Equal(t, []string{"user", AdminRole}, got.Roles, "a listed e-mail is granted admin, case-insensitively")

	again, err := svc.ensureBootstrapAdmin(context.Background(), got)
	require.NoError(t, err)
	assert.Equal(t, []string{"user", AdminRole}, again.Roles, "the role is granted once")

	untouched, err := svc.ensureBootstrapAdmin(context.Background(), other)
	require.NoError(t, err)
	assert.Equal(t, []string{"user"}, untouched.Roles, "an unlisted e-mail keeps its roles")
}

func TestEnsureNotSelfDemotion(t *testing.T) {
	self := &User{ID: uuid.New(), AlternateAppID: 7, Email: "me@example.org", Roles: []string{"user", AdminRole}}
	peer := &User{ID: uuid.New(), AlternateAppID: 8, Email: "peer@example.org", Roles: []string{"user", AdminRole}}
	store := &fakeUserStorage{usersByID: map[uuid.UUID]*User{self.ID: self, peer.ID: peer}}
	server := NewUserConnectServer(NewUserBusinessService(store, discardLogger(), 50), discardLogger())
	ctx := context.WithValue(context.Background(), userIDKey, int32(7))

	err := server.ensureNotSelfDemotion(ctx, self.ID, []string{"user"})
	require.Error(t, err)
	assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))

	assert.NoError(t, server.ensureNotSelfDemotion(ctx, self.ID, []string{"user", AdminRole}), "keeping one's own admin role is fine")
	assert.NoError(t, server.ensureNotSelfDemotion(ctx, peer.ID, []string{"user"}), "an admin may demote another admin")
}

// TestUserDirectoryNeedsAdmin covers the user directory: listing and counting
// users need the admin role, reading one user is for itself or an admin.
func TestUserDirectoryNeedsAdmin(t *testing.T) {
	self := &User{ID: uuid.New(), AlternateAppID: 7, Email: "me@example.org", Roles: []string{"user"}}
	peer := &User{ID: uuid.New(), AlternateAppID: 8, Email: "peer@example.org", Roles: []string{"user"}}
	store := &fakeUserStorage{usersByID: map[uuid.UUID]*User{self.ID: self, peer.ID: peer}}
	server := NewUserConnectServer(NewUserBusinessService(store, discardLogger(), 50), discardLogger())
	user := context.WithValue(context.Background(), userIDKey, int32(7))
	admin := context.WithValue(user, isAdminKey, true)

	_, err := server.List(user, connect.NewRequest(&authv1.ListRequest{}))
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err), "listing users needs the admin role")
	_, err = server.Count(user, connect.NewRequest(&authv1.CountRequest{}))
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err), "counting users needs the admin role")
	_, err = server.Get(user, connect.NewRequest(&authv1.GetRequest{Id: peer.ID.String()}))
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err), "reading another user needs the admin role")
	_, err = server.GetByExternalId(user, connect.NewRequest(&authv1.GetByExternalIdRequest{ExternalId: 8}))
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err), "reading another user by external id needs the admin role")

	_, err = server.Get(user, connect.NewRequest(&authv1.GetRequest{Id: self.ID.String()}))
	assert.NoError(t, err, "a user reads itself")
	_, err = server.Get(admin, connect.NewRequest(&authv1.GetRequest{Id: peer.ID.String()}))
	assert.NoError(t, err, "an admin reads anyone")
}
