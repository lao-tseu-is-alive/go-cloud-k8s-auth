package auth

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// BrowserConfig holds the cookie and allowlist settings for the browser SSO flow.
type BrowserConfig struct {
	// PublicBaseURL is the externally reachable base URL of this auth service,
	// used to build the OAuth callback URL (e.g. http://localhost:9090).
	PublicBaseURL string
	// CookieName is the name of the session cookie (e.g. goSession).
	CookieName string
	// CookieDomain is the Domain attribute of the session cookie. Empty means
	// host-only, which on localhost is shared across ports (cookies are not
	// port-scoped). In production set the parent domain (e.g. .example.com).
	CookieDomain string
	// CookieSecure sets the Secure attribute; must be false on plain-http localhost.
	CookieSecure bool
	// SessionTTL is the lifetime of new sessions and their cookie.
	SessionTTL time.Duration
	// AllowedRedirectURIs is the list of URL prefixes modules may redirect to.
	AllowedRedirectURIs []string
}

// BrowserHandlers exposes the cookie-based SSO endpoints used by module
// frontends: hosted login page, OAuth redirects, silent JWT mint and logout.
// Everything cookie- or redirect-based lives here as plain Echo handlers;
// bearer-token APIs stay on the Connect/Vanguard side.
type BrowserHandlers struct {
	service  *AuthBusinessService
	sessions SessionStorage
	cfg      BrowserConfig
	log      *slog.Logger
}

// NewBrowserHandlers creates the browser SSO handlers.
func NewBrowserHandlers(service *AuthBusinessService, sessions SessionStorage, cfg BrowserConfig, log *slog.Logger) (*BrowserHandlers, error) {
	if service == nil || sessions == nil || log == nil {
		return nil, errors.New("NewBrowserHandlers: service, sessions and log are required")
	}
	if cfg.PublicBaseURL == "" || cfg.CookieName == "" || cfg.SessionTTL <= 0 {
		return nil, errors.New("NewBrowserHandlers: PublicBaseURL, CookieName and a positive SessionTTL are required")
	}
	cfg.PublicBaseURL = strings.TrimRight(cfg.PublicBaseURL, "/")
	return &BrowserHandlers{service: service, sessions: sessions, cfg: cfg, log: log}, nil
}

// RegisterRoutes mounts the browser SSO endpoints on the given Echo instance.
func (h *BrowserHandlers) RegisterRoutes(e *echo.Echo) {
	e.GET("/auth/login", h.handleLogin)
	e.GET("/auth/oauth/:provider/start", h.handleOAuthStart)
	e.GET("/auth/oauth/:provider/callback", h.handleOAuthCallback)
	e.GET("/auth/token", h.handleToken)
	e.POST("/auth/logout", h.handleLogout)
}

// isAllowedRedirect reports whether uri may be used as a post-login redirect.
// Same-origin paths ("/...") are always allowed; absolute URLs must match one
// of the configured prefixes on a path-segment boundary.
func (h *BrowserHandlers) isAllowedRedirect(uri string) bool {
	if uri == "" {
		return false
	}
	// Same-origin relative path (reject scheme-relative "//host" URLs).
	if strings.HasPrefix(uri, "/") && !strings.HasPrefix(uri, "//") {
		return true
	}
	u, err := url.Parse(uri)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}
	for _, prefix := range h.cfg.AllowedRedirectURIs {
		if prefix == "" {
			continue
		}
		if !strings.HasPrefix(uri, prefix) {
			continue
		}
		// Avoid http://localhost:8080.evil.com matching prefix http://localhost:8080
		rest := uri[len(prefix):]
		if strings.HasSuffix(prefix, "/") || rest == "" || rest[0] == '/' || rest[0] == '?' || rest[0] == '#' {
			return true
		}
	}
	return false
}

// callbackURL returns the browser OAuth callback URL for a provider. This URL
// must be registered in the provider's developer console.
func (h *BrowserHandlers) callbackURL(provider string) string {
	return fmt.Sprintf("%s/auth/oauth/%s/callback", h.cfg.PublicBaseURL, provider)
}

func (h *BrowserHandlers) setSessionCookie(c echo.Context, rawToken string) {
	c.SetCookie(&http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    rawToken,
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		MaxAge:   int(h.cfg.SessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *BrowserHandlers) clearSessionCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// currentSession returns the valid session for the request cookie, or nil.
func (h *BrowserHandlers) currentSession(c echo.Context) *Session {
	cookie, err := c.Cookie(h.cfg.CookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}
	session, err := h.sessions.GetValid(c.Request().Context(), cookie.Value)
	if err != nil {
		return nil
	}
	return session
}

var loginPageTemplate = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
    <title>Sign in</title>
    <style>
        body { font-family: system-ui, sans-serif; background: #0f172a; color: #f8fafc;
               min-height: 100vh; display: flex; align-items: center; justify-content: center; margin: 0; }
        .card { background: #1e293b; border: 1px solid rgba(255,255,255,.08); border-radius: 16px;
                padding: 2.5rem; width: 100%; max-width: 380px; text-align: center; }
        h1 { font-size: 1.4rem; margin: 0 0 .5rem; }
        p { color: #94a3b8; margin: 0 0 2rem; font-size: .95rem; }
        a.provider { display: block; margin: .75rem 0; padding: .8rem 1rem; border-radius: 10px;
                     background: #3b82f6; color: #fff; text-decoration: none; font-weight: 600; }
        a.provider.github { background: #24292f; }
        a.provider.microsoft { background: #5e5e5e; }
        a.provider:hover { filter: brightness(1.15); }
    </style>
</head>
<body>
<div class="card">
    <h1>Sign in</h1>
    <p>Choose a provider to continue</p>
    {{- range .Providers }}
    <a class="provider {{ .Name }}" href="/auth/oauth/{{ .Name }}/start?redirect_uri={{ $.RedirectURI }}">Continue with {{ .Label }}</a>
    {{- end }}
</div>
</body>
</html>
`))

type loginPageProvider struct {
	Name  string
	Label string
}

var providerLabels = map[string]string{
	"google":    "Google",
	"github":    "GitHub",
	"microsoft": "Microsoft",
}

// handleLogin renders the hosted login page, or redirects straight back to the
// module when a valid session cookie already exists.
func (h *BrowserHandlers) handleLogin(c echo.Context) error {
	redirectURI := c.QueryParam("redirect_uri")
	if redirectURI == "" {
		redirectURI = "/"
	}
	if !h.isAllowedRedirect(redirectURI) {
		h.log.Warn("login: redirect_uri not in allowlist", "redirect_uri", redirectURI)
		return c.String(http.StatusBadRequest, "redirect_uri is not allowed")
	}

	if h.currentSession(c) != nil {
		return c.Redirect(http.StatusFound, redirectURI)
	}

	providers := make([]loginPageProvider, 0, len(h.service.OAuthConfigs))
	for name := range h.service.OAuthConfigs {
		label, ok := providerLabels[name]
		if !ok {
			label = name
		}
		providers = append(providers, loginPageProvider{Name: name, Label: label})
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].Name < providers[j].Name })

	var sb strings.Builder
	err := loginPageTemplate.Execute(&sb, map[string]any{
		"Providers":   providers,
		"RedirectURI": url.QueryEscape(redirectURI),
	})
	if err != nil {
		h.log.Error("login: template execution failed", "error", err)
		return c.String(http.StatusInternalServerError, "internal error")
	}
	return c.HTML(http.StatusOK, sb.String())
}

// handleOAuthStart begins the OAuth dance: generates a CSRF state bound to the
// module redirect URI and sends the browser to the provider's consent page.
func (h *BrowserHandlers) handleOAuthStart(c echo.Context) error {
	provider := c.Param("provider")
	cfg, ok := h.service.OAuthConfigs[provider]
	if !ok {
		return c.String(http.StatusNotFound, "unknown oauth provider")
	}

	redirectURI := c.QueryParam("redirect_uri")
	if redirectURI == "" {
		redirectURI = "/"
	}
	if !h.isAllowedRedirect(redirectURI) {
		h.log.Warn("oauth start: redirect_uri not in allowlist", "redirect_uri", redirectURI)
		return c.String(http.StatusBadRequest, "redirect_uri is not allowed")
	}

	state, err := h.service.StateStore.GenerateWithRedirect(provider, redirectURI)
	if err != nil {
		h.log.Error("oauth start: state generation failed", "error", err)
		return c.String(http.StatusInternalServerError, "internal error")
	}

	cfgCopy := *cfg
	cfgCopy.RedirectURL = h.callbackURL(provider)
	return c.Redirect(http.StatusFound, cfgCopy.AuthCodeURL(state, oauth2.AccessTypeOffline))
}

// handleOAuthCallback finishes the OAuth dance: validates the state, exchanges
// the code, upserts the user, creates a DB-backed session, sets the session
// cookie and redirects the browser back to the module.
func (h *BrowserHandlers) handleOAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	if providerErr := c.QueryParam("error"); providerErr != "" {
		h.log.Warn("oauth callback: provider returned error", "provider", provider, "error", providerErr)
		return c.String(http.StatusBadRequest, fmt.Sprintf("oauth provider error: %s", providerErr))
	}

	redirectURI, ok := h.service.StateStore.ValidateAndConsume(c.QueryParam("state"), provider)
	if !ok {
		return c.String(http.StatusBadRequest, "invalid or expired oauth state, please retry the login")
	}
	if redirectURI == "" {
		redirectURI = "/"
	}
	// The state already proves we issued it, but re-check defensively in case
	// the allowlist changed between start and callback.
	if !h.isAllowedRedirect(redirectURI) {
		return c.String(http.StatusBadRequest, "redirect_uri is not allowed")
	}

	code := c.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "missing oauth code")
	}

	ctx := c.Request().Context()
	user, err := h.service.AuthenticateOAuthUser(ctx, provider, code, h.callbackURL(provider))
	if err != nil {
		h.log.Error("oauth callback: authentication failed", "provider", provider, "error", err)
		if errors.Is(err, ErrUserDisabled) {
			return c.String(http.StatusForbidden, "this account is disabled")
		}
		return c.String(http.StatusBadGateway, "oauth authentication failed, please retry the login")
	}

	rawToken, err := h.sessions.Create(ctx, user.ID, c.Request().UserAgent(), c.RealIP(), h.cfg.SessionTTL)
	if err != nil {
		h.log.Error("oauth callback: session creation failed", "error", err)
		return c.String(http.StatusInternalServerError, "internal error")
	}
	h.setSessionCookie(c, rawToken)
	h.log.Info("oauth callback: session established", "provider", provider, "email", user.Email, "userId", user.AlternateAppID)
	return c.Redirect(http.StatusFound, redirectURI)
}

// tokenResponse is the JSON payload of the silent token mint endpoint.
type tokenResponse struct {
	Token            string            `json:"token"`
	ExpiresInSeconds int               `json:"expires_in_seconds"`
	User             tokenResponseUser `json:"user"`
}

type tokenResponseUser struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	IsAdmin   bool   `json:"is_admin"`
}

// handleToken silently mints a short-lived JWT for the session cookie owner.
// Module frontends call it with credentials:'include' whenever they need a
// fresh token; 401 means the user must go through /auth/login again.
func (h *BrowserHandlers) handleToken(c echo.Context) error {
	session := h.currentSession(c)
	if session == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "no valid session"})
	}

	ctx := c.Request().Context()
	user, err := h.service.Store.GetByID(ctx, session.UserID)
	if err != nil {
		h.log.Warn("token: session user lookup failed", "userId", session.UserID, "error", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unknown user"})
	}
	if !user.IsActive {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user account is disabled"})
	}

	token, err := h.service.IssueJwtForUser(ctx, user)
	if err != nil {
		h.log.Error("token: JWT issuance failed", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	isAdmin := false
	for _, role := range user.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}
	return c.JSON(http.StatusOK, tokenResponse{
		Token:            token,
		ExpiresInSeconds: h.service.JwtCheck.GetJwtDuration() * 60,
		User: tokenResponseUser{
			UserID:    user.AlternateAppID,
			Email:     user.Email,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
			IsAdmin:   isAdmin,
		},
	})
}

// handleLogout revokes the current session and clears the cookie.
func (h *BrowserHandlers) handleLogout(c echo.Context) error {
	if cookie, err := c.Cookie(h.cfg.CookieName); err == nil && cookie.Value != "" {
		if err := h.sessions.Revoke(c.Request().Context(), cookie.Value); err != nil {
			h.log.Error("logout: session revocation failed", "error", err)
		}
	}
	h.clearSessionCookie(c)
	return c.JSON(http.StatusOK, map[string]string{"status": "logged out"})
}
