package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse Google user info: %w", err)
	}

	return &OAuthUserInfo{
		Provider:   "google",
		ProviderID: data.ID,
		Email:      data.Email,
		Name:       data.Name,
		AvatarURL:  data.Picture,
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

	// If email is not public, fetch from emails endpoint
	email := data.Email
	if email == "" {
		email, _ = fetchGitHubPrimaryEmail(ctx, client)
	}
	name := data.Name
	if name == "" {
		name = data.Login
	}

	return &OAuthUserInfo{
		Provider:   "github",
		ProviderID: fmt.Sprintf("%d", data.ID),
		Email:      email,
		Name:       name,
		AvatarURL:  data.AvatarURL,
	}, nil
}

// fetchGitHubPrimaryEmail fetches the primary email from GitHub's emails API.
func fetchGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", fmt.Errorf("no email found")
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
	}, nil
}

// UserInfoFetcher maps provider names to their user info fetching functions.
var UserInfoFetcher = map[string]func(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error){
	"google":    FetchGoogleUserInfo,
	"github":    FetchGitHubUserInfo,
	"microsoft": FetchMicrosoftUserInfo,
}
