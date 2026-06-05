// Package template_4_your_project_name provides Connect RPC handlers for the TypeTemplate4ServiceNameService.
package template4gopackage

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	template_4_your_project_namev1 "github.com/your-github-account/template-4-your-project-name/gen/template_4_your_project_name/v1"
	"github.com/your-github-account/template-4-your-project-name/gen/template_4_your_project_name/v1/template_4_your_project_namev1connect"
)

// TypeTemplate4ServiceNameConnectServer implements the TypeTemplate4ServiceNameServiceHandler interface.
// Authentication is handled by the AuthInterceptor, which injects user info into context.
type TypeTemplate4ServiceNameConnectServer struct {
	BusinessService *BusinessService
	Log             *slog.Logger

	// Embed the unimplemented handler for forward compatibility
	template_4_your_project_namev1connect.UnimplementedTypeTemplate4ServiceNameServiceHandler
}

// NewTypeTemplate4ServiceNameConnectServer creates a new TypeTemplate4ServiceNameConnectServer.
// Note: Authentication is handled by the AuthInterceptor, not by this server.
func NewTypeTemplate4ServiceNameConnectServer(business *BusinessService, log *slog.Logger) *TypeTemplate4ServiceNameConnectServer {
	return &TypeTemplate4ServiceNameConnectServer{
		BusinessService: business,
		Log:             log,
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// mapErrorToConnect converts business errors to Connect errors
func (s *TypeTemplate4ServiceNameConnectServer) mapErrorToConnect(err error) *connect.Error {
	switch {
	case errors.Is(err, ErrTypeTemplate4ServiceNameNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, ErrAdminRequired):
		return connect.NewError(connect.CodePermissionDenied, errors.New(OnlyAdminCanManageTypeTemplate4ServiceNames))
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
// TypeTemplate4ServiceNameService RPC Methods
// =============================================================================

// List returns a list of type template_4_your_project_names
func (s *TypeTemplate4ServiceNameConnectServer) List(
	ctx context.Context,
	req *connect.Request[template_4_your_project_namev1.TypeTemplate4ServiceNameListRequest],
) (*connect.Response[template_4_your_project_namev1.TypeTemplate4ServiceNameListResponse], error) {
	s.Log.Info("Connect: TypeTemplate4ServiceName.List called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("TypeTemplate4ServiceName.List", "userId", userId)

	msg := req.Msg
	params := TypeTemplate4ServiceNameListParams{}
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
	limit := 250 // Default for TypeTemplate4ServiceName as in HTTP handler
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	list, err := s.BusinessService.ListTypeTemplate4ServiceNames(ctx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return empty list instead of error
			return connect.NewResponse(&template_4_your_project_namev1.TypeTemplate4ServiceNameListResponse{
				TypeTemplate4ServiceNames: []*template_4_your_project_namev1.TypeTemplate4ServiceNameList{},
			}), nil
		}
		return nil, s.mapErrorToConnect(err)
	}

	response := &template_4_your_project_namev1.TypeTemplate4ServiceNameListResponse{
		TypeTemplate4ServiceNames: DomainTypeTemplate4ServiceNameListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}

// Create creates a new type template_4_your_project_name
func (s *TypeTemplate4ServiceNameConnectServer) Create(
	ctx context.Context,
	req *connect.Request[template_4_your_project_namev1.TypeTemplate4ServiceNameCreateRequest],
) (*connect.Response[template_4_your_project_namev1.TypeTemplate4ServiceNameCreateResponse], error) {
	s.Log.Info("Connect: TypeTemplate4ServiceName.Create called")

	// User info injected by AuthInterceptor
	userId, isAdmin := GetUserFromContext(ctx)
	s.Log.Info("TypeTemplate4ServiceName.Create", "userId", userId, "isAdmin", isAdmin)

	protoTypeTemplate4ServiceName := req.Msg.TypeTemplate4ServiceName
	if protoTypeTemplate4ServiceName == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type_template_4_your_project_name is required"))
	}

	domainTypeTemplate4ServiceName := ProtoTypeTemplate4ServiceNameToDomain(protoTypeTemplate4ServiceName)

	createdTypeTemplate4ServiceName, err := s.BusinessService.CreateTypeTemplate4ServiceName(ctx, userId, isAdmin, *domainTypeTemplate4ServiceName)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &template_4_your_project_namev1.TypeTemplate4ServiceNameCreateResponse{
		TypeTemplate4ServiceName: DomainTypeTemplate4ServiceNameToProto(createdTypeTemplate4ServiceName),
	}
	return connect.NewResponse(response), nil
}

// Get retrieves a type template_4_your_project_name by ID
func (s *TypeTemplate4ServiceNameConnectServer) Get(
	ctx context.Context,
	req *connect.Request[template_4_your_project_namev1.TypeTemplate4ServiceNameGetRequest],
) (*connect.Response[template_4_your_project_namev1.TypeTemplate4ServiceNameGetResponse], error) {
	s.Log.Info("Connect: TypeTemplate4ServiceName.Get called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	_, isAdmin := GetUserFromContext(ctx)

	typeTemplate4ServiceName, err := s.BusinessService.GetTypeTemplate4ServiceName(ctx, isAdmin, req.Msg.Id)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &template_4_your_project_namev1.TypeTemplate4ServiceNameGetResponse{
		TypeTemplate4ServiceName: DomainTypeTemplate4ServiceNameToProto(typeTemplate4ServiceName),
	}
	return connect.NewResponse(response), nil
}

// Update updates a type template_4_your_project_name
func (s *TypeTemplate4ServiceNameConnectServer) Update(
	ctx context.Context,
	req *connect.Request[template_4_your_project_namev1.TypeTemplate4ServiceNameUpdateRequest],
) (*connect.Response[template_4_your_project_namev1.TypeTemplate4ServiceNameUpdateResponse], error) {
	s.Log.Info("Connect: TypeTemplate4ServiceName.Update called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, isAdmin := GetUserFromContext(ctx)
	s.Log.Info("TypeTemplate4ServiceName.Update", "userId", userId, "isAdmin", isAdmin)

	protoTypeTemplate4ServiceName := req.Msg.TypeTemplate4ServiceName
	if protoTypeTemplate4ServiceName == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type_template_4_your_project_name data is required"))
	}

	domainTypeTemplate4ServiceName := ProtoTypeTemplate4ServiceNameToDomain(protoTypeTemplate4ServiceName)

	updatedTypeTemplate4ServiceName, err := s.BusinessService.UpdateTypeTemplate4ServiceName(ctx, userId, isAdmin, req.Msg.Id, *domainTypeTemplate4ServiceName)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &template_4_your_project_namev1.TypeTemplate4ServiceNameUpdateResponse{
		TypeTemplate4ServiceName: DomainTypeTemplate4ServiceNameToProto(updatedTypeTemplate4ServiceName),
	}
	return connect.NewResponse(response), nil
}

// Delete deletes a type template_4_your_project_name
func (s *TypeTemplate4ServiceNameConnectServer) Delete(
	ctx context.Context,
	req *connect.Request[template_4_your_project_namev1.TypeTemplate4ServiceNameDeleteRequest],
) (*connect.Response[template_4_your_project_namev1.TypeTemplate4ServiceNameDeleteResponse], error) {
	s.Log.Info("Connect: TypeTemplate4ServiceName.Delete called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, isAdmin := GetUserFromContext(ctx)
	s.Log.Info("TypeTemplate4ServiceName.Delete", "userId", userId, "isAdmin", isAdmin)

	err := s.BusinessService.DeleteTypeTemplate4ServiceName(ctx, userId, isAdmin, req.Msg.Id)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&template_4_your_project_namev1.TypeTemplate4ServiceNameDeleteResponse{}), nil
}

// Count returns the number of type template_4_your_project_names
func (s *TypeTemplate4ServiceNameConnectServer) Count(
	ctx context.Context,
	req *connect.Request[template_4_your_project_namev1.TypeTemplate4ServiceNameCountRequest],
) (*connect.Response[template_4_your_project_namev1.TypeTemplate4ServiceNameCountResponse], error) {
	s.Log.Info("Connect: TypeTemplate4ServiceName.Count called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("TypeTemplate4ServiceName.Count", "userId", userId)

	msg := req.Msg
	params := TypeTemplate4ServiceNameCountParams{}
	if msg.Keywords != "" {
		params.Keywords = &msg.Keywords
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}

	count, err := s.BusinessService.CountTypeTemplate4ServiceNames(ctx, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &template_4_your_project_namev1.TypeTemplate4ServiceNameCountResponse{
		Count: count,
	}
	return connect.NewResponse(response), nil
}
