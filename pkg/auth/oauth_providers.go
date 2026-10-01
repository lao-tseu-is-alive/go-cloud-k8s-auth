package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

// OAuthProviderConfig holds the configuration for building oauth2.Config instances.
type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string // e.g. "http://localhost:9090/auth/google/callback"
}

// BuildGoogleConfig builds an oauth2.Config for Google.
func BuildGoogleConfig(cfg OAuthProviderConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// BuildGitHubConfig builds an oauth2.Config for GitHub.
func BuildGitHubConfig(cfg OAuthProviderConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}
}

// BuildMicrosoftConfig builds an oauth2.Config for Microsoft.
func BuildMicrosoftConfig(cfg OAuthProviderConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{"openid", "email", "profile", "User.Read"},
		Endpoint:     microsoft.AzureADEndpoint("common"),
	}
}

// FetchGoogleUserInfo fetches user profile from Google's userinfo endpoint.
func FetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Google user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Google response: %w", err)
	}

	var data struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		// VerifiedEmail is Google's assertion that the user owns Email.
		VerifiedEmail bool `json:"verified_email"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse Google user info: %w", err)
	}

	return &OAuthUserInfo{
		Provider:      "google",
		ProviderID:    data.ID,
		Email:         data.Email,
		Name:          data.Name,
		AvatarURL:     data.Picture,
		EmailVerified: data.VerifiedEmail,
	}, nil
}

// FetchGitHubUserInfo fetches user profile from GitHub's API.
func FetchGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GitHub response: %w", err)
	}

	var data struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub user info: %w", err)
	}

	// The emails endpoint tells whether an address is verified, also for the
	// public profile address; without it nothing is verified.
	emails, err := fetchGitHubEmails(client)
	if err != nil {
		emails = nil
	}
	email, verified := pickGitHubEmail(data.Email, emails)
	name := data.Name
	if name == "" {
		name = data.Login
	}

	return &OAuthUserInfo{
		Provider:      "github",
		ProviderID:    fmt.Sprintf("%d", data.ID),
		Email:         email,
		Name:          name,
		AvatarURL:     data.AvatarURL,
		EmailVerified: verified,
	}, nil
}

// githubEmail is one address of the GitHub emails API.
type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// fetchGitHubEmails reads the user's addresses from GitHub's emails API.
func fetchGitHubEmails(client *http.Client) ([]githubEmail, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var emails []githubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return nil, err
	}
	return emails, nil
}

// pickGitHubEmail chooses the login e-mail and whether GitHub verified it: the
// public profile address when set, otherwise the primary address, otherwise
// a verified one.
func pickGitHubEmail(public string, emails []githubEmail) (string, bool) {
	if public != "" {
		for _, e := range emails {
			if strings.EqualFold(e.Email, public) {
				return public, e.Verified
			}
		}
		return public, false
	}
	for _, e := range emails {
		if e.Primary {
			return e.Email, e.Verified
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, true
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, false
	}
	return "", false
}

// FetchMicrosoftUserInfo fetches user profile from Microsoft Graph.
func FetchMicrosoftUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Microsoft user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Microsoft response: %w", err)
	}

	var data struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Mail        string `json:"mail"`
		UPN         string `json:"userPrincipalName"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse Microsoft user info: %w", err)
	}

	email := data.Mail
	if email == "" {
		email = data.UPN
	}

	return &OAuthUserInfo{
		Provider:   "microsoft",
		ProviderID: data.ID,
		Email:      email,
		Name:       data.DisplayName,
		AvatarURL:  "", // Microsoft Graph requires separate photo endpoint
		// Graph does not say whether mail is verified (a personal or guest
		// account may show any address): it never links to another account.
		EmailVerified: false,
	}, nil
}

// UserInfoFetcher maps provider names to their user info fetching functions.
var UserInfoFetcher = map[string]func(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error){
	"google":    FetchGoogleUserInfo,
	"github":    FetchGitHubUserInfo,
	"microsoft": FetchMicrosoftUserInfo,
}
