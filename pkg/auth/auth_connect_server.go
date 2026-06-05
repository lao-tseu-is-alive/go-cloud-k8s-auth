// Package auth provides Connect RPC handlers for the AuthService.
package auth

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	authv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1/authv1connect"
)

// AuthConnectServer implements the AuthServiceHandler interface for ConnectRPC.
type AuthConnectServer struct {
	AuthService *AuthBusinessService
	Log         *slog.Logger

	// Embed the unimplemented handler for forward compatibility
	authv1connect.UnimplementedAuthServiceHandler
}

// NewAuthConnectServer creates a new AuthConnectServer.
func NewAuthConnectServer(authService *AuthBusinessService, log *slog.Logger) *AuthConnectServer {
	return &AuthConnectServer{
		AuthService: authService,
		Log:         log,
	}
}

// StartOAuth initiates the OAuth flow for the specified provider.
func (s *AuthConnectServer) StartOAuth(
	ctx context.Context,
	req *connect.Request[authv1.StartOAuthRequest],
) (*connect.Response[authv1.StartOAuthResponse], error) {
	s.Log.Info("Connect: StartOAuth called", "provider", req.Msg.Provider)

	authURL, state, err := s.AuthService.StartOAuth(req.Msg.Provider, req.Msg.RedirectUrl)
	if err != nil {
		if errors.Is(err, ErrProviderError) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		s.Log.Error("StartOAuth failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}

	return connect.NewResponse(&authv1.StartOAuthResponse{
		AuthUrl: authURL,
		State:   state,
	}), nil
}

// OAuthCallback handles the OAuth callback, exchanges the code, and returns a JWT.
func (s *AuthConnectServer) OAuthCallback(
	ctx context.Context,
	req *connect.Request[authv1.OAuthCallbackRequest],
) (*connect.Response[authv1.OAuthCallbackResponse], error) {
	s.Log.Info("Connect: OAuthCallback called", "provider", req.Msg.Provider)

	jwtToken, user, err := s.AuthService.HandleCallback(ctx, req.Msg.Provider, req.Msg.Code, req.Msg.State)
	if err != nil {
		if errors.Is(err, ErrInvalidState) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		if errors.Is(err, ErrUserDisabled) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		if errors.Is(err, ErrProviderError) {
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		s.Log.Error("OAuthCallback failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}

	return connect.NewResponse(&authv1.OAuthCallbackResponse{
		Jwt:  jwtToken,
		User: DomainUserToProto(user),
	}), nil
}

// ValidateToken validates a JWT token and returns the associated user.
func (s *AuthConnectServer) ValidateToken(
	ctx context.Context,
	req *connect.Request[authv1.ValidateTokenRequest],
) (*connect.Response[authv1.ValidateTokenResponse], error) {
	s.Log.Info("Connect: ValidateToken called")

	valid, user, err := s.AuthService.ValidateToken(ctx, req.Msg.Token)
	if err != nil {
		s.Log.Error("ValidateToken failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}

	return connect.NewResponse(&authv1.ValidateTokenResponse{
		Valid: valid,
		User:  DomainUserToProto(user),
	}), nil
}

// GetCurrentUser returns the currently authenticated user from JWT claims.
func (s *AuthConnectServer) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[authv1.GetCurrentUserRequest],
) (*connect.Response[authv1.GetCurrentUserResponse], error) {
	s.Log.Info("Connect: GetCurrentUser called")

	// User info is injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	if userId == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("not authenticated"))
	}

	user, err := s.AuthService.GetCurrentUser(ctx, int64(userId))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		s.Log.Error("GetCurrentUser failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}

	return connect.NewResponse(&authv1.GetCurrentUserResponse{
		User: DomainUserToProto(user),
	}), nil
}
