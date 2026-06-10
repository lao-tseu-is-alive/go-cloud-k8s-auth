// Package auth provides Connect RPC handlers for personal access tokens.
package auth

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	authv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// patInfoToProto converts a PersonalAccessToken to its proto metadata form.
func patInfoToProto(p *PersonalAccessToken) *authv1.PersonalAccessTokenInfo {
	if p == nil {
		return nil
	}
	info := &authv1.PersonalAccessTokenInfo{
		Id:        p.ID.String(),
		Name:      p.Name,
		Prefix:    p.Prefix,
		Scopes:    p.Scopes,
		CreatedAt: timestamppb.New(p.CreatedAt),
		Revoked:   p.RevokedAt != nil,
	}
	if p.ExpiresAt != nil {
		info.ExpiresAt = timestamppb.New(*p.ExpiresAt)
	}
	if p.LastUsedAt != nil {
		info.LastUsedAt = timestamppb.New(*p.LastUsedAt)
	}
	return info
}

// currentUserUUID resolves the interceptor-injected integer user id to the
// user's UUID, required by the PAT storage layer.
func (s *AuthConnectServer) currentUserUUID(ctx context.Context, log *slog.Logger) (uuid.UUID, *connect.Error) {
	userId, _ := GetUserFromContext(ctx)
	if userId == 0 {
		return uuid.Nil, connect.NewError(connect.CodeUnauthenticated, errors.New("not authenticated"))
	}
	user, err := s.AuthService.Store.GetByAlternateAppID(ctx, int64(userId))
	if err != nil {
		log.Warn("currentUserUUID: user lookup failed", "userId", userId, "error", err)
		return uuid.Nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unknown user"))
	}
	return user.ID, nil
}

// IntrospectToken verifies an opaque token (PAT) and returns identity + scopes.
// Public procedure: the token itself is the credential.
func (s *AuthConnectServer) IntrospectToken(
	ctx context.Context,
	req *connect.Request[authv1.IntrospectTokenRequest],
) (*connect.Response[authv1.IntrospectTokenResponse], error) {
	s.Log.Debug("Connect: IntrospectToken called")

	result, err := s.PatService.Introspect(ctx, req.Msg.Token)
	if err != nil {
		s.Log.Error("IntrospectToken failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}

	resp := &authv1.IntrospectTokenResponse{Active: result.Active}
	if result.Active {
		resp.UserId = result.UserID
		resp.Email = result.Email
		resp.Name = result.Name
		resp.Scopes = result.Scopes
		if result.ExpiresAt != nil {
			resp.ExpiresAt = timestamppb.New(*result.ExpiresAt)
		}
	}
	return connect.NewResponse(resp), nil
}

// CreatePersonalAccessToken creates a PAT for the authenticated user.
func (s *AuthConnectServer) CreatePersonalAccessToken(
	ctx context.Context,
	req *connect.Request[authv1.CreatePersonalAccessTokenRequest],
) (*connect.Response[authv1.CreatePersonalAccessTokenResponse], error) {
	s.Log.Info("Connect: CreatePersonalAccessToken called", "name", req.Msg.Name)

	userUUID, cErr := s.currentUserUUID(ctx, s.Log)
	if cErr != nil {
		return nil, cErr
	}

	token, pat, err := s.PatService.Create(ctx, userUUID, req.Msg.Name, req.Msg.Scopes, req.Msg.ExpiresInDays)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		s.Log.Error("CreatePersonalAccessToken failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}

	return connect.NewResponse(&authv1.CreatePersonalAccessTokenResponse{
		Token: token,
		Info:  patInfoToProto(pat),
	}), nil
}

// ListPersonalAccessTokens lists the authenticated user's PATs.
func (s *AuthConnectServer) ListPersonalAccessTokens(
	ctx context.Context,
	req *connect.Request[authv1.ListPersonalAccessTokensRequest],
) (*connect.Response[authv1.ListPersonalAccessTokensResponse], error) {
	s.Log.Debug("Connect: ListPersonalAccessTokens called")

	userUUID, cErr := s.currentUserUUID(ctx, s.Log)
	if cErr != nil {
		return nil, cErr
	}

	pats, err := s.PatService.List(ctx, userUUID)
	if err != nil {
		s.Log.Error("ListPersonalAccessTokens failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}

	infos := make([]*authv1.PersonalAccessTokenInfo, len(pats))
	for i, p := range pats {
		infos[i] = patInfoToProto(p)
	}
	return connect.NewResponse(&authv1.ListPersonalAccessTokensResponse{Tokens: infos}), nil
}

// RevokePersonalAccessToken revokes one of the authenticated user's PATs.
func (s *AuthConnectServer) RevokePersonalAccessToken(
	ctx context.Context,
	req *connect.Request[authv1.RevokePersonalAccessTokenRequest],
) (*connect.Response[authv1.RevokePersonalAccessTokenResponse], error) {
	s.Log.Info("Connect: RevokePersonalAccessToken called", "id", req.Msg.Id)

	userUUID, cErr := s.currentUserUUID(ctx, s.Log)
	if cErr != nil {
		return nil, cErr
	}

	patID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id must be a valid UUID"))
	}

	if err := s.PatService.Revoke(ctx, patID, userUUID); err != nil {
		if errors.Is(err, ErrPatNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		s.Log.Error("RevokePersonalAccessToken failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}
	return connect.NewResponse(&authv1.RevokePersonalAccessTokenResponse{}), nil
}
