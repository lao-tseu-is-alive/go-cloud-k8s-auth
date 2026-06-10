package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"golang.org/x/oauth2"
)

// AuthBusinessService handles OAuth flows, JWT signing, and token validation.
type AuthBusinessService struct {
	Store        UserStorage
	JwtCheck     goHttpEcho.JwtChecker
	OAuthConfigs map[string]*oauth2.Config // keyed by provider name: "google", "github", "microsoft"
	StateStore   *StateStore
	Log          *slog.Logger
}

// NewAuthBusinessService creates a new AuthBusinessService.
func NewAuthBusinessService(
	store UserStorage,
	jwtCheck goHttpEcho.JwtChecker,
	oauthConfigs map[string]*oauth2.Config,
	log *slog.Logger,
) *AuthBusinessService {
	return &AuthBusinessService{
		Store:        store,
		JwtCheck:     jwtCheck,
		OAuthConfigs: oauthConfigs,
		StateStore:   NewStateStore(),
		Log:          log,
	}
}

// GetAvailableProviders returns the list of configured OAuth providers.
func (s *AuthBusinessService) GetAvailableProviders() []string {
	providers := make([]string, 0, len(s.OAuthConfigs))
	for name := range s.OAuthConfigs {
		providers = append(providers, name)
	}
	return providers
}

// StartOAuth generates the OAuth authorization URL for the given provider.
// Returns the URL the client should redirect to, and the state for CSRF protection.
func (s *AuthBusinessService) StartOAuth(provider, redirectURL string) (authURL string, state string, err error) {
	cfg, ok := s.OAuthConfigs[provider]
	if !ok {
		return "", "", fmt.Errorf("%w: unsupported provider %q, available: %v", ErrProviderError, provider, s.GetAvailableProviders())
	}

	state, err = s.StateStore.Generate(provider)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// If a custom redirect URL is provided, override the configured one
	if redirectURL != "" {
		cfg = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  redirectURL,
			Scopes:       cfg.Scopes,
			Endpoint:     cfg.Endpoint,
		}
	}

	authURL = cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	s.Log.Info("StartOAuth: generated auth URL", "provider", provider)
	return authURL, state, nil
}

// HandleCallback processes the OAuth callback: validates state, exchanges code for token,
// fetches user info from provider, upserts user in DB, and returns a signed JWT.
func (s *AuthBusinessService) HandleCallback(ctx context.Context, provider, code, state string) (jwtToken string, user *User, err error) {
	// 1. Validate state (CSRF protection)
	if !s.StateStore.Validate(state, provider) {
		return "", nil, ErrInvalidState
	}

	// 2-6. Exchange code, fetch profile, upsert user, check active
	user, err = s.AuthenticateOAuthUser(ctx, provider, code, "")
	if err != nil {
		return "", nil, err
	}

	// 7-8. Generate JWT token with group claims
	token, err := s.IssueJwtForUser(ctx, user)
	if err != nil {
		return "", nil, err
	}

	s.Log.Info("HandleCallback: user authenticated", "email", user.Email, "provider", provider, "userId", user.AlternateAppID)
	return token, user, nil
}

// AuthenticateOAuthUser exchanges an authorization code for the provider's
// access token, fetches the user profile, upserts the user in the database and
// checks that the account is active. It does NOT validate the CSRF state nor
// issue any credential, so both the Connect RPC callback and the browser
// cookie flow can share it. redirectURL overrides the configured OAuth
// redirect URL for the code exchange (required when the authorization request
// used a different callback than the configured one); pass "" to keep the
// configured value.
func (s *AuthBusinessService) AuthenticateOAuthUser(ctx context.Context, provider, code, redirectURL string) (*User, error) {
	cfg, ok := s.OAuthConfigs[provider]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported provider %q", ErrProviderError, provider)
	}
	if redirectURL != "" {
		cfgCopy := *cfg
		cfgCopy.RedirectURL = redirectURL
		cfg = &cfgCopy
	}

	oauthToken, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: token exchange failed: %v", ErrProviderError, err)
	}

	fetchFn, ok := UserInfoFetcher[provider]
	if !ok {
		return nil, fmt.Errorf("%w: no user info fetcher for provider %q", ErrProviderError, provider)
	}
	oauthUserInfo, err := fetchFn(ctx, oauthToken)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch user info: %v", ErrProviderError, err)
	}

	if oauthUserInfo.Email == "" {
		return nil, fmt.Errorf("%w: provider returned no email", ErrProviderError)
	}

	user, err := s.Store.UpsertByProvider(ctx,
		oauthUserInfo.Email,
		oauthUserInfo.Name,
		oauthUserInfo.AvatarURL,
		oauthUserInfo.Provider,
		oauthUserInfo.ProviderID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserDisabled
	}
	return user, nil
}

// IssueJwtForUser generates a signed JWT for the given user, including group claims.
func (s *AuthBusinessService) IssueJwtForUser(ctx context.Context, user *User) (string, error) {
	groupIDs, err := s.getGroupIDsAsInts(ctx, user)
	if err != nil {
		s.Log.Warn("failed to fetch group IDs, using empty", "error", err)
		groupIDs = []int{}
	}
	token, err := s.JwtCheck.GetTokenFromUserInfo(DomainUserToJwtUserInfo(user, groupIDs))
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}
	return token.String(), nil
}

// ValidateToken validates a JWT token and returns the associated user.
func (s *AuthBusinessService) ValidateToken(ctx context.Context, tokenStr string) (bool, *User, error) {
	claims, err := s.JwtCheck.ParseToken(tokenStr)
	if err != nil {
		return false, nil, nil // Invalid token, but not an error
	}

	user, err := s.Store.GetByAlternateAppID(ctx, int64(claims.User.UserId))
	if err != nil {
		return false, nil, fmt.Errorf("user lookup failed: %w", err)
	}

	if !user.IsActive {
		return false, nil, nil
	}

	return true, user, nil
}

// GetCurrentUser retrieves the user by their alternate app ID (from JWT claims).
func (s *AuthBusinessService) GetCurrentUser(ctx context.Context, userAppID int64) (*User, error) {
	user, err := s.Store.GetByAlternateAppID(ctx, userAppID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}
	return user, nil
}

// getGroupIDsAsInts fetches group UUIDs and converts them to integer hashes for JWT.
// For Phase 1, we use a simple conversion of the UUID's least significant bytes.
func (s *AuthBusinessService) getGroupIDsAsInts(ctx context.Context, user *User) ([]int, error) {
	groupUUIDs, err := s.Store.GetUserGroupIDs(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	result := make([]int, len(groupUUIDs))
	for i, gid := range groupUUIDs {
		// Use the last 4 bytes of UUID as a deterministic int
		bytes := gid[12:16]
		result[i] = int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
	}
	return result, nil
}
