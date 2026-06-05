package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

// BusinessService Business Service contains the transport-agnostic business logic for Auth operations
type BusinessService struct {
	Log              *slog.Logger
	DbConn           database.DB
	Store            Storage
	ListDefaultLimit int
}

// NewBusinessService creates a new instance of BusinessService
func NewBusinessService(store Storage, dbConn database.DB, log *slog.Logger, listDefaultLimit int) *BusinessService {
	return &BusinessService{
		Log:              log,
		DbConn:           dbConn,
		Store:            store,
		ListDefaultLimit: listDefaultLimit,
	}
}

// validateName validates the name field according to business rules
func validateName(name string) error {
	if len(strings.Trim(name, " ")) < 1 {
		return fmt.Errorf(FieldCannotBeEmpty, "name")
	}
	if len(name) < MinNameLength {
		return fmt.Errorf(FieldMinLengthIsN, "name", MinNameLength)
	}
	return nil
}

// GeoJson returns a geoJson representation of go_cloud_auths based on the given parameters
func (s *BusinessService) GeoJson(ctx context.Context, offset, limit int, params GeoJsonParams) (string, error) {
	jsonResult, err := s.Store.GeoJson(ctx, offset, limit, params)
	if err != nil {
		return "", fmt.Errorf("error retrieving geoJson: %w", err)
	}
	if jsonResult == "" {
		return "empty", nil
	}
	return jsonResult, nil
}

// List returns the list of go_cloud_auths based on the given parameters
func (s *BusinessService) List(ctx context.Context, offset, limit int, params ListParams) ([]*AuthList, error) {
	list, err := s.Store.List(ctx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows is not an error, return empty slice
			return make([]*AuthList, 0), nil
		}
		return nil, fmt.Errorf("error listing go_cloud_auths: %w", err)
	}
	if list == nil {
		return make([]*AuthList, 0), nil
	}
	return list, nil
}

// Create creates a new go_cloud_auth with the given data
func (s *BusinessService) Create(ctx context.Context, currentUserId int32, newAuth Auth) (*Auth, error) {
	// Validate name
	if err := validateName(newAuth.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Validate TypeId
	typeAuthCount, err := s.DbConn.GetQueryInt(ctx, existTypeAuth, newAuth.TypeId)
	if err != nil || typeAuthCount < 1 {
		return nil, fmt.Errorf("%w: typeId %v", ErrTypeAuthNotFound, newAuth.TypeId)
	}

	// Check if go_cloud_auth already exists
	if s.Store.Exist(ctx, newAuth.Id) {
		return nil, fmt.Errorf("%w: id %v", ErrAlreadyExists, newAuth.Id)
	}

	// Set creator
	newAuth.CreatedBy = currentUserId

	// Create in storage
	go_cloud_authCreated, err := s.Store.Create(ctx, newAuth)
	if err != nil {
		return nil, fmt.Errorf("error creating go_cloud_auth: %w", err)
	}

	s.Log.Info("Created go_cloud_auth", "id", go_cloud_authCreated.Id, "userId", currentUserId)
	return go_cloud_authCreated, nil
}

// Count returns the number of go_cloud_auths based on the given parameters
func (s *BusinessService) Count(ctx context.Context, params CountParams) (int32, error) {
	numAuths, err := s.Store.Count(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("error counting go_cloud_auths: %w", err)
	}
	return numAuths, nil
}

// Delete removes a go_cloud_auth with the given ID
func (s *BusinessService) Delete(ctx context.Context, currentUserId int32, go_cloud_authId uuid.UUID) error {
	// Check if go_cloud_auth exists
	if !s.Store.Exist(ctx, go_cloud_authId) {
		return fmt.Errorf("%w: id %v", ErrNotFound, go_cloud_authId)
	}

	// Check if user is owner
	if !s.Store.IsUserOwner(ctx, go_cloud_authId, currentUserId) {
		return fmt.Errorf("%w: user %d is not owner of go_cloud_auth %v", ErrUnauthorized, currentUserId, go_cloud_authId)
	}

	// Delete from storage
	err := s.Store.Delete(ctx, go_cloud_authId, currentUserId)
	if err != nil {
		return fmt.Errorf("error deleting go_cloud_auth: %w", err)
	}

	s.Log.Info("Deleted go_cloud_auth", "id", go_cloud_authId, "userId", currentUserId)
	return nil
}

// Get retrieves a go_cloud_auth by its ID
func (s *BusinessService) Get(ctx context.Context, go_cloud_authId uuid.UUID) (*Auth, error) {
	// Check if go_cloud_auth exists
	if !s.Store.Exist(ctx, go_cloud_authId) {
		return nil, fmt.Errorf("%w: id %v", ErrNotFound, go_cloud_authId)
	}

	// Get from storage
	go_cloud_auth, err := s.Store.Get(ctx, go_cloud_authId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving go_cloud_auth: %w", err)
	}

	return go_cloud_auth, nil
}

// Update updates a go_cloud_auth with the given ID
func (s *BusinessService) Update(ctx context.Context, currentUserId int32, go_cloud_authId uuid.UUID, updateAuth Auth) (*Auth, error) {
	// Check if go_cloud_auth exists
	if !s.Store.Exist(ctx, go_cloud_authId) {
		return nil, fmt.Errorf("%w: id %v", ErrNotFound, go_cloud_authId)
	}

	// Check if user is owner
	if !s.Store.IsUserOwner(ctx, go_cloud_authId, currentUserId) {
		return nil, fmt.Errorf("%w: user %d is not owner of go_cloud_auth %v", ErrUnauthorized, currentUserId, go_cloud_authId)
	}

	// Validate name
	if err := validateName(updateAuth.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Validate TypeId
	typeAuthCount, err := s.DbConn.GetQueryInt(ctx, existTypeAuth, updateAuth.TypeId)
	if err != nil || typeAuthCount < 1 {
		return nil, fmt.Errorf("%w: typeId %v", ErrTypeAuthNotFound, updateAuth.TypeId)
	}

	// Set last modifier
	updateAuth.LastModifiedBy = &currentUserId

	// Update in storage
	go_cloud_authUpdated, err := s.Store.Update(ctx, go_cloud_authId, updateAuth)
	if err != nil {
		return nil, fmt.Errorf("error updating go_cloud_auth: %w", err)
	}

	s.Log.Info("Updated go_cloud_auth", "id", go_cloud_authId, "userId", currentUserId)
	return go_cloud_authUpdated, nil
}

// ListByExternalId returns go_cloud_auths filtered by external ID
func (s *BusinessService) ListByExternalId(ctx context.Context, offset, limit, externalId int) ([]*AuthList, error) {
	list, err := s.Store.ListByExternalId(ctx, offset, limit, externalId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows is not an error, return empty slice
			return make([]*AuthList, 0), nil
		}
		return nil, fmt.Errorf("error listing go_cloud_auths by external id: %w", err)
	}
	if list == nil {
		return make([]*AuthList, 0), nil
	}
	return list, nil
}

// Search returns go_cloud_auths based on search criteria
func (s *BusinessService) Search(ctx context.Context, offset, limit int, params SearchParams) ([]*AuthList, error) {
	list, err := s.Store.Search(ctx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows is not an error, return empty slice
			return make([]*AuthList, 0), nil
		}
		return nil, fmt.Errorf("error searching go_cloud_auths: %w", err)
	}
	if list == nil {
		return make([]*AuthList, 0), nil
	}
	return list, nil
}

// ListTypeAuths returns a list of TypeAuth based on parameters
func (s *BusinessService) ListTypeAuths(ctx context.Context, offset, limit int, params TypeAuthListParams) ([]*TypeAuthList, error) {
	list, err := s.Store.ListTypeAuth(ctx, offset, limit, params)
	if err != nil {
		return nil, fmt.Errorf("error listing type go_cloud_auths: %w", err)
	}
	if list == nil {
		return make([]*TypeAuthList, 0), nil
	}
	return list, nil
}

// CreateTypeAuth creates a new TypeAuth
func (s *BusinessService) CreateTypeAuth(ctx context.Context, currentUserId int32, isAdmin bool, newTypeAuth TypeAuth) (*TypeAuth, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Validate name
	if err := validateName(newTypeAuth.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Set creator
	newTypeAuth.CreatedBy = currentUserId

	// Create in storage
	typeAuthCreated, err := s.Store.CreateTypeAuth(ctx, newTypeAuth)
	if err != nil {
		return nil, fmt.Errorf("error creating type go_cloud_auth: %w", err)
	}

	s.Log.Info("Created TypeAuth", "id", typeAuthCreated.Id, "userId", currentUserId)
	return typeAuthCreated, nil
}

// CountTypeAuths returns the count of TypeAuths based on parameters
func (s *BusinessService) CountTypeAuths(ctx context.Context, params TypeAuthCountParams) (int32, error) {
	numAuths, err := s.Store.CountTypeAuth(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("error counting type go_cloud_auths: %w", err)
	}
	return numAuths, nil
}

// DeleteTypeAuth deletes a TypeAuth by ID
func (s *BusinessService) DeleteTypeAuth(ctx context.Context, currentUserId int32, isAdmin bool, typeAuthId int32) error {
	// Check admin privileges
	if !isAdmin {
		return ErrAdminRequired
	}

	// Check if TypeAuth exists
	typeAuthCount, err := s.DbConn.GetQueryInt(ctx, existTypeAuth, typeAuthId)
	if err != nil || typeAuthCount < 1 {
		return fmt.Errorf("%w: id %d", ErrTypeAuthNotFound, typeAuthId)
	}

	// Delete from storage
	err = s.Store.DeleteTypeAuth(ctx, typeAuthId, currentUserId)
	if err != nil {
		return fmt.Errorf("error deleting type go_cloud_auth: %w", err)
	}

	s.Log.Info("Deleted TypeAuth", "id", typeAuthId, "userId", currentUserId)
	return nil
}

// GetTypeAuth retrieves a TypeAuth by ID
func (s *BusinessService) GetTypeAuth(ctx context.Context, isAdmin bool, typeAuthId int32) (*TypeAuth, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Check if TypeAuth exists
	typeAuthCount, err := s.DbConn.GetQueryInt(ctx, existTypeAuth, typeAuthId)
	if err != nil || typeAuthCount < 1 {
		return nil, fmt.Errorf("%w: id %d", ErrTypeAuthNotFound, typeAuthId)
	}

	// Get from storage
	typeAuth, err := s.Store.GetTypeAuth(ctx, typeAuthId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving type go_cloud_auth: %w", err)
	}

	return typeAuth, nil
}

// UpdateTypeAuth updates a TypeAuth
func (s *BusinessService) UpdateTypeAuth(ctx context.Context, currentUserId int32, isAdmin bool, typeAuthId int32, updateTypeAuth TypeAuth) (*TypeAuth, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Check if TypeAuth exists
	typeAuthCount, err := s.DbConn.GetQueryInt(ctx, existTypeAuth, typeAuthId)
	if err != nil || typeAuthCount < 1 {
		return nil, fmt.Errorf("%w: id %d", ErrTypeAuthNotFound, typeAuthId)
	}

	// Validate name
	if err := validateName(updateTypeAuth.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Set last modifier
	updateTypeAuth.LastModifiedBy = &currentUserId

	// Update in storage
	go_cloud_authUpdated, err := s.Store.UpdateTypeAuth(ctx, typeAuthId, updateTypeAuth)
	if err != nil {
		return nil, fmt.Errorf("error updating type go_cloud_auth: %w", err)
	}

	s.Log.Info("Updated TypeAuth", "id", typeAuthId, "userId", currentUserId)
	return go_cloud_authUpdated, nil
}
