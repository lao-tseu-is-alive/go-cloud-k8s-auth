// Package auth provides Connect RPC handlers for the UserService.
package auth

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	authv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1/authv1connect"
)

// UserConnectServer implements the UserServiceHandler interface for ConnectRPC.
type UserConnectServer struct {
	UserService *UserBusinessService
	Log         *slog.Logger

	// Embed the unimplemented handler for forward compatibility
	authv1connect.UnimplementedUserServiceHandler
}

// NewUserConnectServer creates a new UserConnectServer.
func NewUserConnectServer(userService *UserBusinessService, log *slog.Logger) *UserConnectServer {
	return &UserConnectServer{
		UserService: userService,
		Log:         log,
	}
}

// mapErrorToConnect converts business errors to Connect errors.
func (s *UserConnectServer) mapErrorToConnect(err error) *connect.Error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, ErrUnauthorized):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, ErrAdminRequired):
		return connect.NewError(connect.CodePermissionDenied, errors.New(OnlyAdminCanManageUsers))
	case errors.Is(err, ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	default:
		s.Log.Error("internal error", "error", err)
		return connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}
}

// List returns a paginated list of users.
func (s *UserConnectServer) List(
	ctx context.Context,
	req *connect.Request[authv1.ListRequest],
) (*connect.Response[authv1.ListResponse], error) {
	s.Log.Info("Connect: List called")

	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("List", "callerUserId", userId)

	msg := req.Msg
	limit := s.UserService.ListDefaultLimit
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	var disabledFilter *bool
	if msg.Disabled {
		disabledFilter = &msg.Disabled
	}

	list, err := s.UserService.List(ctx, offset, limit, disabledFilter)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&authv1.ListResponse{
		Users: DomainUserListSliceToProto(list),
	}), nil
}

// Create creates a new user.
func (s *UserConnectServer) Create(
	ctx context.Context,
	req *connect.Request[authv1.CreateRequest],
) (*connect.Response[authv1.CreateResponse], error) {
	s.Log.Info("Connect: Create called")

	// Ensure caller is an admin
	_, isAdmin := GetUserFromContext(ctx)
	if !isAdmin {
		return nil, s.mapErrorToConnect(ErrAdminRequired)
	}

	protoUser := req.Msg.User
	if protoUser == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user is required"))
	}

	domainUser, err := ProtoUserToDomain(protoUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	created, err := s.UserService.Create(ctx, *domainUser)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&authv1.CreateResponse{
		User: DomainUserToProto(created),
	}), nil
}

// Get retrieves a user by ID.
func (s *UserConnectServer) Get(
	ctx context.Context,
	req *connect.Request[authv1.GetRequest],
) (*connect.Response[authv1.GetResponse], error) {
	s.Log.Info("Connect: Get called", "id", req.Msg.Id)

	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid user ID format"))
	}

	user, err := s.UserService.Get(ctx, id)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&authv1.GetResponse{
		User: DomainUserToProto(user),
	}), nil
}

// Update updates a user.
func (s *UserConnectServer) Update(
	ctx context.Context,
	req *connect.Request[authv1.UpdateRequest],
) (*connect.Response[authv1.UpdateResponse], error) {
	s.Log.Info("Connect: Update called", "id", req.Msg.Id)

	// Ensure caller is an admin
	_, isAdmin := GetUserFromContext(ctx)
	if !isAdmin {
		return nil, s.mapErrorToConnect(ErrAdminRequired)
	}

	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid user ID format"))
	}

	protoUser := req.Msg.User
	if protoUser == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user data is required"))
	}

	domainUser, err := ProtoUserToDomain(protoUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	updated, err := s.UserService.Update(ctx, id, *domainUser)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&authv1.UpdateResponse{
		User: DomainUserToProto(updated),
	}), nil
}

// Delete deletes a user.
func (s *UserConnectServer) Delete(
	ctx context.Context,
	req *connect.Request[authv1.DeleteRequest],
) (*connect.Response[authv1.DeleteResponse], error) {
	s.Log.Info("Connect: Delete called", "id", req.Msg.Id)

	// Ensure caller is an admin
	_, isAdmin := GetUserFromContext(ctx)
	if !isAdmin {
		return nil, s.mapErrorToConnect(ErrAdminRequired)
	}

	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid user ID format"))
	}

	if err := s.UserService.Delete(ctx, id); err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&authv1.DeleteResponse{}), nil
}

// Count returns the number of users.
func (s *UserConnectServer) Count(
	ctx context.Context,
	req *connect.Request[authv1.CountRequest],
) (*connect.Response[authv1.CountResponse], error) {
	s.Log.Info("Connect: Count called")

	var disabledFilter *bool
	if req.Msg.Inactivated {
		disabledFilter = &req.Msg.Inactivated
	}

	count, err := s.UserService.Count(ctx, disabledFilter)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&authv1.CountResponse{
		Count: count,
	}), nil
}

// GetByExternalId retrieves a user by their alternate app ID.
func (s *UserConnectServer) GetByExternalId(
	ctx context.Context,
	req *connect.Request[authv1.GetByExternalIdRequest],
) (*connect.Response[authv1.GetByExternalIdResponse], error) {
	s.Log.Info("Connect: GetByExternalId called", "externalId", req.Msg.ExternalId)

	user, err := s.UserService.GetByExternalId(ctx, int64(req.Msg.ExternalId))
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&authv1.GetByExternalIdResponse{
		User: DomainUserToProto(user),
	}), nil
}
