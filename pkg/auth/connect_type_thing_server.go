// Package go_cloud_auth provides Connect RPC handlers for the TypeAuthService.
package auth

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	go_cloud_authv1 "https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/go_cloud_auth/v1"
	"https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/go_cloud_auth/v1/go_cloud_authv1connect"
)

// TypeAuthConnectServer implements the TypeAuthServiceHandler interface.
// Authentication is handled by the AuthInterceptor, which injects user info into context.
type TypeAuthConnectServer struct {
	BusinessService *BusinessService
	Log             *slog.Logger

	// Embed the unimplemented handler for forward compatibility
	go_cloud_authv1connect.UnimplementedTypeAuthServiceHandler
}

// NewTypeAuthConnectServer creates a new TypeAuthConnectServer.
// Note: Authentication is handled by the AuthInterceptor, not by this server.
func NewTypeAuthConnectServer(business *BusinessService, log *slog.Logger) *TypeAuthConnectServer {
	return &TypeAuthConnectServer{
		BusinessService: business,
		Log:             log,
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// mapErrorToConnect converts business errors to Connect errors
func (s *TypeAuthConnectServer) mapErrorToConnect(err error) *connect.Error {
	switch {
	case errors.Is(err, ErrTypeAuthNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, ErrAdminRequired):
		return connect.NewError(connect.CodePermissionDenied, errors.New(OnlyAdminCanManageTypeAuths))
	case errors.Is(err, ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, pgx.ErrNoRows):
		return connect.NewError(connect.CodeNotFound, errors.New("not found"))
	default:
		s.Log.Error("internal error", "error", err)
		return connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}
}

// =============================================================================
// TypeAuthService RPC Methods
// =============================================================================

// List returns a list of type go_cloud_auths
func (s *TypeAuthConnectServer) List(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.TypeAuthListRequest],
) (*connect.Response[go_cloud_authv1.TypeAuthListResponse], error) {
	s.Log.Info("Connect: TypeAuth.List called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("TypeAuth.List", "userId", userId)

	msg := req.Msg
	params := TypeAuthListParams{}
	if msg.Keywords != "" {
		params.Keywords = &msg.Keywords
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.ExternalId != 0 {
		params.ExternalId = &msg.ExternalId
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}

	// Handle pagination
	limit := 250 // Default for TypeAuth as in HTTP handler
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	list, err := s.BusinessService.ListTypeAuths(ctx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return empty list instead of error
			return connect.NewResponse(&go_cloud_authv1.TypeAuthListResponse{
				TypeAuths: []*go_cloud_authv1.TypeAuthList{},
			}), nil
		}
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.TypeAuthListResponse{
		TypeAuths: DomainTypeAuthListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}

// Create creates a new type go_cloud_auth
func (s *TypeAuthConnectServer) Create(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.TypeAuthCreateRequest],
) (*connect.Response[go_cloud_authv1.TypeAuthCreateResponse], error) {
	s.Log.Info("Connect: TypeAuth.Create called")

	// User info injected by AuthInterceptor
	userId, isAdmin := GetUserFromContext(ctx)
	s.Log.Info("TypeAuth.Create", "userId", userId, "isAdmin", isAdmin)

	protoTypeAuth := req.Msg.TypeAuth
	if protoTypeAuth == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type_go_cloud_auth is required"))
	}

	domainTypeAuth := ProtoTypeAuthToDomain(protoTypeAuth)

	createdTypeAuth, err := s.BusinessService.CreateTypeAuth(ctx, userId, isAdmin, *domainTypeAuth)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.TypeAuthCreateResponse{
		TypeAuth: DomainTypeAuthToProto(createdTypeAuth),
	}
	return connect.NewResponse(response), nil
}

// Get retrieves a type go_cloud_auth by ID
func (s *TypeAuthConnectServer) Get(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.TypeAuthGetRequest],
) (*connect.Response[go_cloud_authv1.TypeAuthGetResponse], error) {
	s.Log.Info("Connect: TypeAuth.Get called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	_, isAdmin := GetUserFromContext(ctx)

	typeAuth, err := s.BusinessService.GetTypeAuth(ctx, isAdmin, req.Msg.Id)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.TypeAuthGetResponse{
		TypeAuth: DomainTypeAuthToProto(typeAuth),
	}
	return connect.NewResponse(response), nil
}

// Update updates a type go_cloud_auth
func (s *TypeAuthConnectServer) Update(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.TypeAuthUpdateRequest],
) (*connect.Response[go_cloud_authv1.TypeAuthUpdateResponse], error) {
	s.Log.Info("Connect: TypeAuth.Update called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, isAdmin := GetUserFromContext(ctx)
	s.Log.Info("TypeAuth.Update", "userId", userId, "isAdmin", isAdmin)

	protoTypeAuth := req.Msg.TypeAuth
	if protoTypeAuth == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type_go_cloud_auth data is required"))
	}

	domainTypeAuth := ProtoTypeAuthToDomain(protoTypeAuth)

	updatedTypeAuth, err := s.BusinessService.UpdateTypeAuth(ctx, userId, isAdmin, req.Msg.Id, *domainTypeAuth)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.TypeAuthUpdateResponse{
		TypeAuth: DomainTypeAuthToProto(updatedTypeAuth),
	}
	return connect.NewResponse(response), nil
}

// Delete deletes a type go_cloud_auth
func (s *TypeAuthConnectServer) Delete(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.TypeAuthDeleteRequest],
) (*connect.Response[go_cloud_authv1.TypeAuthDeleteResponse], error) {
	s.Log.Info("Connect: TypeAuth.Delete called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, isAdmin := GetUserFromContext(ctx)
	s.Log.Info("TypeAuth.Delete", "userId", userId, "isAdmin", isAdmin)

	err := s.BusinessService.DeleteTypeAuth(ctx, userId, isAdmin, req.Msg.Id)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&go_cloud_authv1.TypeAuthDeleteResponse{}), nil
}

// Count returns the number of type go_cloud_auths
func (s *TypeAuthConnectServer) Count(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.TypeAuthCountRequest],
) (*connect.Response[go_cloud_authv1.TypeAuthCountResponse], error) {
	s.Log.Info("Connect: TypeAuth.Count called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("TypeAuth.Count", "userId", userId)

	msg := req.Msg
	params := TypeAuthCountParams{}
	if msg.Keywords != "" {
		params.Keywords = &msg.Keywords
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}

	count, err := s.BusinessService.CountTypeAuths(ctx, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.TypeAuthCountResponse{
		Count: count,
	}
	return connect.NewResponse(response), nil
}
