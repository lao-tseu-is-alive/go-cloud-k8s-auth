package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// fakeUserStorage implements UserStorage for browser handler tests.
type fakeUserStorage struct {
	UserStorage
	usersByID map[uuid.UUID]*User
}

func (f *fakeUserStorage) GetByID(_ context.Context, id uuid.UUID) (*User, error) {
	u, ok := f.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (f *fakeUserStorage) GetUserGroupIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{}, nil
}

// fakeSessionStore implements SessionStorage in memory.
type fakeSessionStore struct {
	sessions map[string]*Session
	revoked  []string
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{sessions: make(map[string]*Session)}
}

func (f *fakeSessionStore) Create(_ context.Context, userID uuid.UUID, _, _ string, ttl time.Duration) (string, error) {
	raw := "session-" + uuid.NewString()
	f.sessions[raw] = &Session{ID: uuid.New(), UserID: userID, ExpiresAt: time.Now().Add(ttl)}
	return raw, nil
}

func (f *fakeSessionStore) GetValid(_ context.Context, rawToken string) (*Session, error) {
	s, ok := f.sessions[rawToken]
	if !ok || time.Now().After(s.ExpiresAt) {
		return nil, ErrSessionNotFound
	}
	return s, nil
}

func (f *fakeSessionStore) Revoke(_ context.Context, rawToken string) error {
	delete(f.sessions, rawToken)
	f.revoked = append(f.revoked, rawToken)
	return nil
}

func (f *fakeSessionStore) RevokeAllForUser(_ context.Context, _ uuid.UUID) error { return nil }

func newTestBrowserHandlers(t *testing.T) (*BrowserHandlers, *fakeSessionStore, *fakeUserStorage, *User) {
	t.Helper()
	log := golog.NewLogger("simple", io.Discard, golog.ErrorLevel, "browser_handlers_test")
	jwtCheck := goHttpEcho.NewJwtChecker("test-secret", "test-issuer", "test-subject", "jwtdata", 15, log)

	user := &User{
		ID:             uuid.New(),
		AlternateAppID: 4242,
		Email:          "jane@example.com",
		Name:           "Jane Doe",
		Roles:          []string{"user"},
		IsActive:       true,
	}
	store := &fakeUserStorage{usersByID: map[uuid.UUID]*User{user.ID: user}}
	service := NewAuthBusinessService(store, jwtCheck, map[string]*oauth2.Config{
		"github": {
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			Endpoint:     oauth2.Endpoint{AuthURL: "https://github.example/authorize", TokenURL: "https://github.example/token"},
		},
	}, log)

	sessions := newFakeSessionStore()
	handlers, err := NewBrowserHandlers(service, sessions, BrowserConfig{
		PublicBaseURL:       "http://localhost:9090",
		CookieName:          "goSession",
		SessionTTL:          time.Hour,
		AllowedRedirectURIs: []string{"http://localhost:8080"},
	}, log)
	require.NoError(t, err)
	return handlers, sessions, store, user
}

func doRequest(h *BrowserHandlers, req *http.Request) *httptest.ResponseRecorder {
	e := echo.New()
	h.RegisterRoutes(e)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestIsAllowedRedirect(t *testing.T) {
	h, _, _, _ := newTestBrowserHandlers(t)

	assert.True(t, h.isAllowedRedirect("http://localhost:8080"))
	assert.True(t, h.isAllowedRedirect("http://localhost:8080/"))
	assert.True(t, h.isAllowedRedirect("http://localhost:8080/some/page?x=1"))
	assert.True(t, h.isAllowedRedirect("/tokens.html"), "same-origin paths are always allowed")

	assert.False(t, h.isAllowedRedirect(""))
	assert.False(t, h.isAllowedRedirect("http://localhost:8080.evil.com/"), "prefix must end on a path boundary")
	assert.False(t, h.isAllowedRedirect("http://evil.com/"))
	assert.False(t, h.isAllowedRedirect("//evil.com/page"), "scheme-relative URLs are rejected")
	assert.False(t, h.isAllowedRedirect("javascript:alert(1)"))
}

func TestHandleLogin(t *testing.T) {
	h, sessions, _, user := newTestBrowserHandlers(t)

	t.Run("rejects disallowed redirect_uri", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/login?redirect_uri=http://evil.com/", nil)
		rec := doRequest(h, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("renders provider buttons without a session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/login?redirect_uri=http://localhost:8080/", nil)
		rec := doRequest(h, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "/auth/oauth/github/start")
		assert.Contains(t, body, "Continue with GitHub")
	})

	t.Run("redirects back immediately with a valid session", func(t *testing.T) {
		raw, err := sessions.Create(context.Background(), user.ID, "", "", time.Hour)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/auth/login?redirect_uri=http://localhost:8080/", nil)
		req.AddCookie(&http.Cookie{Name: "goSession", Value: raw})
		rec := doRequest(h, req)
		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "http://localhost:8080/", rec.Header().Get("Location"))
	})
}

func TestHandleOAuthStart(t *testing.T) {
	h, _, _, _ := newTestBrowserHandlers(t)

	t.Run("unknown provider", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/oauth/gitlab/start", nil)
		rec := doRequest(h, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("redirects to the provider with browser callback URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/oauth/github/start?redirect_uri=http://localhost:8080/", nil)
		rec := doRequest(h, req)
		assert.Equal(t, http.StatusFound, rec.Code)
		location := rec.Header().Get("Location")
		assert.True(t, strings.HasPrefix(location, "https://github.example/authorize"), location)
		assert.Contains(t, location, "state=")
		assert.Contains(t, location, "redirect_uri=http%3A%2F%2Flocalhost%3A9090%2Fauth%2Foauth%2Fgithub%2Fcallback")
	})
}

func TestHandleOAuthCallbackInvalidState(t *testing.T) {
	h, _, _, _ := newTestBrowserHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/github/callback?code=abc&state=bogus", nil)
	rec := doRequest(h, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleToken(t *testing.T) {
	h, sessions, store, user := newTestBrowserHandlers(t)

	t.Run("401 without cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/token", nil)
		rec := doRequest(h, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("401 with bogus cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/token", nil)
		req.AddCookie(&http.Cookie{Name: "goSession", Value: "bogus"})
		rec := doRequest(h, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("mints a JWT for a valid session", func(t *testing.T) {
		raw, err := sessions.Create(context.Background(), user.ID, "", "", time.Hour)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/auth/token", nil)
		req.AddCookie(&http.Cookie{Name: "goSession", Value: raw})
		rec := doRequest(h, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp tokenResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, 15*60, resp.ExpiresInSeconds)
		assert.Equal(t, int64(4242), resp.User.UserID)
		assert.Equal(t, "jane@example.com", resp.User.Email)
		assert.False(t, resp.User.IsAdmin)

		// The minted token must parse with the same checker.
		claims, err := h.service.JwtCheck.ParseToken(resp.Token)
		require.NoError(t, err)
		assert.Equal(t, 4242, claims.User.UserId)
	})

	t.Run("401 for a disabled user", func(t *testing.T) {
		disabled := &User{ID: uuid.New(), AlternateAppID: 777, Email: "off@example.com", IsActive: false}
		store.usersByID[disabled.ID] = disabled
		raw, err := sessions.Create(context.Background(), disabled.ID, "", "", time.Hour)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/auth/token", nil)
		req.AddCookie(&http.Cookie{Name: "goSession", Value: raw})
		rec := doRequest(h, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestHandleLogout(t *testing.T) {
	h, sessions, _, user := newTestBrowserHandlers(t)
	raw, err := sessions.Create(context.Background(), user.ID, "", "", time.Hour)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "goSession", Value: raw})
	rec := doRequest(h, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, sessions.revoked, raw)

	// The Set-Cookie header must expire the session cookie.
	setCookie := rec.Header().Get("Set-Cookie")
	assert.Contains(t, setCookie, "goSession=")
	assert.Contains(t, setCookie, "Max-Age=0")

	// Session is gone: token mint must now fail.
	req = httptest.NewRequest(http.MethodGet, "/auth/token", nil)
	req.AddCookie(&http.Cookie{Name: "goSession", Value: raw})
	rec = doRequest(h, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
