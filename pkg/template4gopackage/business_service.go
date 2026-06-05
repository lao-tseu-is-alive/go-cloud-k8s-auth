package template4gopackage

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

// BusinessService Business Service contains the transport-agnostic business logic for Template4ServiceName operations
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

// GeoJson returns a geoJson representation of template_4_your_project_names based on the given parameters
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

// List returns the list of template_4_your_project_names based on the given parameters
func (s *BusinessService) List(ctx context.Context, offset, limit int, params ListParams) ([]*Template4ServiceNameList, error) {
	list, err := s.Store.List(ctx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows is not an error, return empty slice
			return make([]*Template4ServiceNameList, 0), nil
		}
		return nil, fmt.Errorf("error listing template_4_your_project_names: %w", err)
	}
	if list == nil {
		return make([]*Template4ServiceNameList, 0), nil
	}
	return list, nil
}

// Create creates a new template_4_your_project_name with the given data
func (s *BusinessService) Create(ctx context.Context, currentUserId int32, newTemplate4ServiceName Template4ServiceName) (*Template4ServiceName, error) {
	// Validate name
	if err := validateName(newTemplate4ServiceName.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Validate TypeId
	typeTemplate4ServiceNameCount, err := s.DbConn.GetQueryInt(ctx, existTypeTemplate4ServiceName, newTemplate4ServiceName.TypeId)
	if err != nil || typeTemplate4ServiceNameCount < 1 {
		return nil, fmt.Errorf("%w: typeId %v", ErrTypeTemplate4ServiceNameNotFound, newTemplate4ServiceName.TypeId)
	}

	// Check if template_4_your_project_name already exists
	if s.Store.Exist(ctx, newTemplate4ServiceName.Id) {
		return nil, fmt.Errorf("%w: id %v", ErrAlreadyExists, newTemplate4ServiceName.Id)
	}

	// Set creator
	newTemplate4ServiceName.CreatedBy = currentUserId

	// Create in storage
	template_4_your_project_nameCreated, err := s.Store.Create(ctx, newTemplate4ServiceName)
	if err != nil {
		return nil, fmt.Errorf("error creating template_4_your_project_name: %w", err)
	}

	s.Log.Info("Created template_4_your_project_name", "id", template_4_your_project_nameCreated.Id, "userId", currentUserId)
	return template_4_your_project_nameCreated, nil
}

// Count returns the number of template_4_your_project_names based on the given parameters
func (s *BusinessService) Count(ctx context.Context, params CountParams) (int32, error) {
	numTemplate4ServiceNames, err := s.Store.Count(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("error counting template_4_your_project_names: %w", err)
	}
	return numTemplate4ServiceNames, nil
}

// Delete removes a template_4_your_project_name with the given ID
func (s *BusinessService) Delete(ctx context.Context, currentUserId int32, template_4_your_project_nameId uuid.UUID) error {
	// Check if template_4_your_project_name exists
	if !s.Store.Exist(ctx, template_4_your_project_nameId) {
		return fmt.Errorf("%w: id %v", ErrNotFound, template_4_your_project_nameId)
	}

	// Check if user is owner
	if !s.Store.IsUserOwner(ctx, template_4_your_project_nameId, currentUserId) {
		return fmt.Errorf("%w: user %d is not owner of template_4_your_project_name %v", ErrUnauthorized, currentUserId, template_4_your_project_nameId)
	}

	// Delete from storage
	err := s.Store.Delete(ctx, template_4_your_project_nameId, currentUserId)
	if err != nil {
		return fmt.Errorf("error deleting template_4_your_project_name: %w", err)
	}

	s.Log.Info("Deleted template_4_your_project_name", "id", template_4_your_project_nameId, "userId", currentUserId)
	return nil
}

// Get retrieves a template_4_your_project_name by its ID
func (s *BusinessService) Get(ctx context.Context, template_4_your_project_nameId uuid.UUID) (*Template4ServiceName, error) {
	// Check if template_4_your_project_name exists
	if !s.Store.Exist(ctx, template_4_your_project_nameId) {
		return nil, fmt.Errorf("%w: id %v", ErrNotFound, template_4_your_project_nameId)
	}

	// Get from storage
	template_4_your_project_name, err := s.Store.Get(ctx, template_4_your_project_nameId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving template_4_your_project_name: %w", err)
	}

	return template_4_your_project_name, nil
}

// Update updates a template_4_your_project_name with the given ID
func (s *BusinessService) Update(ctx context.Context, currentUserId int32, template_4_your_project_nameId uuid.UUID, updateTemplate4ServiceName Template4ServiceName) (*Template4ServiceName, error) {
	// Check if template_4_your_project_name exists
	if !s.Store.Exist(ctx, template_4_your_project_nameId) {
		return nil, fmt.Errorf("%w: id %v", ErrNotFound, template_4_your_project_nameId)
	}

	// Check if user is owner
	if !s.Store.IsUserOwner(ctx, template_4_your_project_nameId, currentUserId) {
		return nil, fmt.Errorf("%w: user %d is not owner of template_4_your_project_name %v", ErrUnauthorized, currentUserId, template_4_your_project_nameId)
	}

	// Validate name
	if err := validateName(updateTemplate4ServiceName.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Validate TypeId
	typeTemplate4ServiceNameCount, err := s.DbConn.GetQueryInt(ctx, existTypeTemplate4ServiceName, updateTemplate4ServiceName.TypeId)
	if err != nil || typeTemplate4ServiceNameCount < 1 {
		return nil, fmt.Errorf("%w: typeId %v", ErrTypeTemplate4ServiceNameNotFound, updateTemplate4ServiceName.TypeId)
	}

	// Set last modifier
	updateTemplate4ServiceName.LastModifiedBy = &currentUserId

	// Update in storage
	template_4_your_project_nameUpdated, err := s.Store.Update(ctx, template_4_your_project_nameId, updateTemplate4ServiceName)
	if err != nil {
		return nil, fmt.Errorf("error updating template_4_your_project_name: %w", err)
	}

	s.Log.Info("Updated template_4_your_project_name", "id", template_4_your_project_nameId, "userId", currentUserId)
	return template_4_your_project_nameUpdated, nil
}

// ListByExternalId returns template_4_your_project_names filtered by external ID
func (s *BusinessService) ListByExternalId(ctx context.Context, offset, limit, externalId int) ([]*Template4ServiceNameList, error) {
	list, err := s.Store.ListByExternalId(ctx, offset, limit, externalId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows is not an error, return empty slice
			return make([]*Template4ServiceNameList, 0), nil
		}
		return nil, fmt.Errorf("error listing template_4_your_project_names by external id: %w", err)
	}
	if list == nil {
		return make([]*Template4ServiceNameList, 0), nil
	}
	return list, nil
}

// Search returns template_4_your_project_names based on search criteria
func (s *BusinessService) Search(ctx context.Context, offset, limit int, params SearchParams) ([]*Template4ServiceNameList, error) {
	list, err := s.Store.Search(ctx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows is not an error, return empty slice
			return make([]*Template4ServiceNameList, 0), nil
		}
		return nil, fmt.Errorf("error searching template_4_your_project_names: %w", err)
	}
	if list == nil {
		return make([]*Template4ServiceNameList, 0), nil
	}
	return list, nil
}

// ListTypeTemplate4ServiceNames returns a list of TypeTemplate4ServiceName based on parameters
func (s *BusinessService) ListTypeTemplate4ServiceNames(ctx context.Context, offset, limit int, params TypeTemplate4ServiceNameListParams) ([]*TypeTemplate4ServiceNameList, error) {
	list, err := s.Store.ListTypeTemplate4ServiceName(ctx, offset, limit, params)
	if err != nil {
		return nil, fmt.Errorf("error listing type template_4_your_project_names: %w", err)
	}
	if list == nil {
		return make([]*TypeTemplate4ServiceNameList, 0), nil
	}
	return list, nil
}

// CreateTypeTemplate4ServiceName creates a new TypeTemplate4ServiceName
func (s *BusinessService) CreateTypeTemplate4ServiceName(ctx context.Context, currentUserId int32, isAdmin bool, newTypeTemplate4ServiceName TypeTemplate4ServiceName) (*TypeTemplate4ServiceName, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Validate name
	if err := validateName(newTypeTemplate4ServiceName.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Set creator
	newTypeTemplate4ServiceName.CreatedBy = currentUserId

	// Create in storage
	typeTemplate4ServiceNameCreated, err := s.Store.CreateTypeTemplate4ServiceName(ctx, newTypeTemplate4ServiceName)
	if err != nil {
		return nil, fmt.Errorf("error creating type template_4_your_project_name: %w", err)
	}

	s.Log.Info("Created TypeTemplate4ServiceName", "id", typeTemplate4ServiceNameCreated.Id, "userId", currentUserId)
	return typeTemplate4ServiceNameCreated, nil
}

// CountTypeTemplate4ServiceNames returns the count of TypeTemplate4ServiceNames based on parameters
func (s *BusinessService) CountTypeTemplate4ServiceNames(ctx context.Context, params TypeTemplate4ServiceNameCountParams) (int32, error) {
	numTemplate4ServiceNames, err := s.Store.CountTypeTemplate4ServiceName(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("error counting type template_4_your_project_names: %w", err)
	}
	return numTemplate4ServiceNames, nil
}

// DeleteTypeTemplate4ServiceName deletes a TypeTemplate4ServiceName by ID
func (s *BusinessService) DeleteTypeTemplate4ServiceName(ctx context.Context, currentUserId int32, isAdmin bool, typeTemplate4ServiceNameId int32) error {
	// Check admin privileges
	if !isAdmin {
		return ErrAdminRequired
	}

	// Check if TypeTemplate4ServiceName exists
	typeTemplate4ServiceNameCount, err := s.DbConn.GetQueryInt(ctx, existTypeTemplate4ServiceName, typeTemplate4ServiceNameId)
	if err != nil || typeTemplate4ServiceNameCount < 1 {
		return fmt.Errorf("%w: id %d", ErrTypeTemplate4ServiceNameNotFound, typeTemplate4ServiceNameId)
	}

	// Delete from storage
	err = s.Store.DeleteTypeTemplate4ServiceName(ctx, typeTemplate4ServiceNameId, currentUserId)
	if err != nil {
		return fmt.Errorf("error deleting type template_4_your_project_name: %w", err)
	}

	s.Log.Info("Deleted TypeTemplate4ServiceName", "id", typeTemplate4ServiceNameId, "userId", currentUserId)
	return nil
}

// GetTypeTemplate4ServiceName retrieves a TypeTemplate4ServiceName by ID
func (s *BusinessService) GetTypeTemplate4ServiceName(ctx context.Context, isAdmin bool, typeTemplate4ServiceNameId int32) (*TypeTemplate4ServiceName, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Check if TypeTemplate4ServiceName exists
	typeTemplate4ServiceNameCount, err := s.DbConn.GetQueryInt(ctx, existTypeTemplate4ServiceName, typeTemplate4ServiceNameId)
	if err != nil || typeTemplate4ServiceNameCount < 1 {
		return nil, fmt.Errorf("%w: id %d", ErrTypeTemplate4ServiceNameNotFound, typeTemplate4ServiceNameId)
	}

	// Get from storage
	typeTemplate4ServiceName, err := s.Store.GetTypeTemplate4ServiceName(ctx, typeTemplate4ServiceNameId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving type template_4_your_project_name: %w", err)
	}

	return typeTemplate4ServiceName, nil
}

// UpdateTypeTemplate4ServiceName updates a TypeTemplate4ServiceName
func (s *BusinessService) UpdateTypeTemplate4ServiceName(ctx context.Context, currentUserId int32, isAdmin bool, typeTemplate4ServiceNameId int32, updateTypeTemplate4ServiceName TypeTemplate4ServiceName) (*TypeTemplate4ServiceName, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Check if TypeTemplate4ServiceName exists
	typeTemplate4ServiceNameCount, err := s.DbConn.GetQueryInt(ctx, existTypeTemplate4ServiceName, typeTemplate4ServiceNameId)
	if err != nil || typeTemplate4ServiceNameCount < 1 {
		return nil, fmt.Errorf("%w: id %d", ErrTypeTemplate4ServiceNameNotFound, typeTemplate4ServiceNameId)
	}

	// Validate name
	if err := validateName(updateTypeTemplate4ServiceName.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Set last modifier
	updateTypeTemplate4ServiceName.LastModifiedBy = &currentUserId

	// Update in storage
	template_4_your_project_nameUpdated, err := s.Store.UpdateTypeTemplate4ServiceName(ctx, typeTemplate4ServiceNameId, updateTypeTemplate4ServiceName)
	if err != nil {
		return nil, fmt.Errorf("error updating type template_4_your_project_name: %w", err)
	}

	s.Log.Info("Updated TypeTemplate4ServiceName", "id", typeTemplate4ServiceNameId, "userId", currentUserId)
	return template_4_your_project_nameUpdated, nil
}
