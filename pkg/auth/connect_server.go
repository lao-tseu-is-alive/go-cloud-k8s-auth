// Package go_cloud_auth provides Connect RPC handlers for the AuthService.
package auth

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	go_cloud_authv1 "https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/go_cloud_auth/v1"
	"https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/go_cloud_auth/v1/go_cloud_authv1connect"
)

// AuthConnectServer implements the AuthServiceHandler interface.
// Authentication is handled by the AuthInterceptor, which injects user info into context.
type AuthConnectServer struct {
	BusinessService *BusinessService
	Log             *slog.Logger

	// Embed the unimplemented handler for forward compatibility
	go_cloud_authv1connect.UnimplementedAuthServiceHandler
}

// NewAuthConnectServer creates a new AuthConnectServer.
// Note: Authentication is handled by the AuthInterceptor, not by this server.
func NewAuthConnectServer(business *BusinessService, log *slog.Logger) *AuthConnectServer {
	return &AuthConnectServer{
		BusinessService: business,
		Log:             log,
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// mapErrorToConnect converts business errors to Connect errors
func (s *AuthConnectServer) mapErrorToConnect(err error) *connect.Error {
	switch {
	case errors.Is(err, ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrTypeAuthNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, ErrUnauthorized):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, err)
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
// AuthService RPC Methods
// =============================================================================

// List returns a list of go_cloud_auths
func (s *AuthConnectServer) List(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.ListRequest],
) (*connect.Response[go_cloud_authv1.ListResponse], error) {
	s.Log.Info("Connect: List called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("List", "userId", userId)

	// Build domain params from proto request
	msg := req.Msg
	params := ListParams{}
	if msg.Type != 0 {
		params.Type = &msg.Type
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}
	if msg.Validated {
		params.Validated = &msg.Validated
	}

	// Handle pagination with defaults
	limit := s.BusinessService.ListDefaultLimit
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	// Call business logic
	list, err := s.BusinessService.List(ctx, offset, limit, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	// Convert to proto and return
	response := &go_cloud_authv1.ListResponse{
		Auths: DomainAuthListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}

// Create creates a new go_cloud_auth
func (s *AuthConnectServer) Create(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.CreateRequest],
) (*connect.Response[go_cloud_authv1.CreateResponse], error) {
	s.Log.Info("Connect: Create called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Create", "userId", userId)

	// Convert proto to domain
	protoAuth := req.Msg.Auth
	if protoAuth == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("go_cloud_auth is required"))
	}

	domainAuth, err := ProtoAuthToDomain(protoAuth)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Call business logic
	createdAuth, err := s.BusinessService.Create(ctx, userId, *domainAuth)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	// Convert back to proto
	response := &go_cloud_authv1.CreateResponse{
		Auth: DomainAuthToProto(createdAuth),
	}
	return connect.NewResponse(response), nil
}

// Get retrieves a go_cloud_auth by ID
func (s *AuthConnectServer) Get(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.GetRequest],
) (*connect.Response[go_cloud_authv1.GetResponse], error) {
	s.Log.Info("Connect: Get called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Get", "userId", userId)

	// Parse UUID
	go_cloud_authId, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid go_cloud_auth ID format"))
	}

	// Call business logic
	go_cloud_auth, err := s.BusinessService.Get(ctx, go_cloud_authId)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.GetResponse{
		Auth: DomainAuthToProto(go_cloud_auth),
	}
	return connect.NewResponse(response), nil
}

// Update updates a go_cloud_auth
func (s *AuthConnectServer) Update(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.UpdateRequest],
) (*connect.Response[go_cloud_authv1.UpdateResponse], error) {
	s.Log.Info("Connect: Update called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Update", "userId", userId)

	// Parse UUID
	go_cloud_authId, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid go_cloud_auth ID format"))
	}

	// Convert proto to domain
	protoAuth := req.Msg.Auth
	if protoAuth == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("go_cloud_auth data is required"))
	}

	domainAuth, err := ProtoAuthToDomain(protoAuth)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Call business logic
	updatedAuth, err := s.BusinessService.Update(ctx, userId, go_cloud_authId, *domainAuth)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.UpdateResponse{
		Auth: DomainAuthToProto(updatedAuth),
	}
	return connect.NewResponse(response), nil
}

// Delete deletes a go_cloud_auth
func (s *AuthConnectServer) Delete(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.DeleteRequest],
) (*connect.Response[go_cloud_authv1.DeleteResponse], error) {
	s.Log.Info("Connect: Delete called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Delete", "userId", userId)

	// Parse UUID
	go_cloud_authId, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid go_cloud_auth ID format"))
	}

	// Call business logic
	err = s.BusinessService.Delete(ctx, userId, go_cloud_authId)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&go_cloud_authv1.DeleteResponse{}), nil
}

// Search returns go_cloud_auths based on search criteria
func (s *AuthConnectServer) Search(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.SearchRequest],
) (*connect.Response[go_cloud_authv1.SearchResponse], error) {
	s.Log.Info("Connect: Search called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Search", "userId", userId)

	msg := req.Msg
	params := SearchParams{}
	if msg.Keywords != "" {
		params.Keywords = &msg.Keywords
	}
	if msg.Type != 0 {
		params.Type = &msg.Type
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}
	if msg.Validated {
		params.Validated = &msg.Validated
	}

	limit := s.BusinessService.ListDefaultLimit
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	list, err := s.BusinessService.Search(ctx, offset, limit, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.SearchResponse{
		Auths: DomainAuthListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}

// Count returns the number of go_cloud_auths
func (s *AuthConnectServer) Count(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.CountRequest],
) (*connect.Response[go_cloud_authv1.CountResponse], error) {
	s.Log.Info("Connect: Count called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Count", "userId", userId)

	msg := req.Msg
	params := CountParams{}
	if msg.Keywords != "" {
		params.Keywords = &msg.Keywords
	}
	if msg.Type != 0 {
		params.Type = &msg.Type
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}
	if msg.Validated {
		params.Validated = &msg.Validated
	}

	count, err := s.BusinessService.Count(ctx, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.CountResponse{
		Count: count,
	}
	return connect.NewResponse(response), nil
}

// GeoJson returns a GeoJSON representation of go_cloud_auths
func (s *AuthConnectServer) GeoJson(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.GeoJsonRequest],
) (*connect.Response[go_cloud_authv1.GeoJsonResponse], error) {
	s.Log.Info("Connect: GeoJson called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("GeoJson", "userId", userId)

	msg := req.Msg
	params := GeoJsonParams{}
	if msg.Type != 0 {
		params.Type = &msg.Type
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}
	if msg.Validated {
		params.Validated = &msg.Validated
	}

	limit := s.BusinessService.ListDefaultLimit
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	result, err := s.BusinessService.GeoJson(ctx, offset, limit, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &go_cloud_authv1.GeoJsonResponse{
		Result: result,
	}
	return connect.NewResponse(response), nil
}

// ListByExternalId returns go_cloud_auths filtered by external ID
func (s *AuthConnectServer) ListByExternalId(
	ctx context.Context,
	req *connect.Request[go_cloud_authv1.ListByExternalIdRequest],
) (*connect.Response[go_cloud_authv1.ListByExternalIdResponse], error) {
	s.Log.Info("Connect: ListByExternalId called", "externalId", req.Msg.ExternalId)

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("ListByExternalId", "userId", userId)

	msg := req.Msg
	limit := s.BusinessService.ListDefaultLimit
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	list, err := s.BusinessService.ListByExternalId(ctx, offset, limit, int(msg.ExternalId))
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	// Return NotFound if no results (matching HTTP handler behavior)
	if len(list) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no go_cloud_auths found with this external ID"))
	}

	response := &go_cloud_authv1.ListByExternalIdResponse{
		Auths: DomainAuthListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}
