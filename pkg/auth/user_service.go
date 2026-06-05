package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// UserBusinessService contains transport-agnostic business logic for User CRUD operations.
type UserBusinessService struct {
	Store            UserStorage
	ListDefaultLimit int
	Log              *slog.Logger
}

// NewUserBusinessService creates a new UserBusinessService.
func NewUserBusinessService(store UserStorage, log *slog.Logger, listDefaultLimit int) *UserBusinessService {
	return &UserBusinessService{
		Store:            store,
		ListDefaultLimit: listDefaultLimit,
		Log:              log,
	}
}

// validateUserInput validates common user fields.
func validateUserInput(name, email string) error {
	if len(strings.TrimSpace(name)) < MinNameLength {
		return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(FieldMinLengthIsN, "name", MinNameLength))
	}
	if len(strings.TrimSpace(email)) < MinEmailLength {
		return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(FieldMinLengthIsN, "email", MinEmailLength))
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("%w: email must contain @", ErrInvalidInput)
	}
	return nil
}

// List returns a paginated list of users.
func (s *UserBusinessService) List(ctx context.Context, offset, limit int, disabledFilter *bool) ([]*UserList, error) {
	list, err := s.Store.List(ctx, offset, limit, disabledFilter)
	if err != nil {
		if err == pgx.ErrNoRows {
			return make([]*UserList, 0), nil
		}
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	return list, nil
}

// Create creates a new user with the given data.
func (s *UserBusinessService) Create(ctx context.Context, user User) (*User, error) {
	if err := validateUserInput(user.Name, user.Email); err != nil {
		return nil, err
	}

	if user.Provider == "" {
		user.Provider = "local"
	}
	if len(user.Roles) == 0 {
		user.Roles = []string{"user"}
	}
	user.IsActive = true

	created, err := s.Store.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	s.Log.Info("Created user", "id", created.ID, "email", created.Email)
	return created, nil
}

// Get retrieves a user by UUID.
func (s *UserBusinessService) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.Store.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("%w: id %v", ErrUserNotFound, id)
		}
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	return user, nil
}

// Update updates a user with the given UUID.
func (s *UserBusinessService) Update(ctx context.Context, id uuid.UUID, user User) (*User, error) {
	if err := validateUserInput(user.Name, user.Email); err != nil {
		return nil, err
	}

	if !s.Store.Exist(ctx, id) {
		return nil, fmt.Errorf("%w: id %v", ErrUserNotFound, id)
	}

	updated, err := s.Store.Update(ctx, id, user)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}
	s.Log.Info("Updated user", "id", id)
	return updated, nil
}

// Delete removes a user by UUID.
func (s *UserBusinessService) Delete(ctx context.Context, id uuid.UUID) error {
	if !s.Store.Exist(ctx, id) {
		return fmt.Errorf("%w: id %v", ErrUserNotFound, id)
	}

	if err := s.Store.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	s.Log.Info("Deleted user", "id", id)
	return nil
}

// Count returns the number of users.
func (s *UserBusinessService) Count(ctx context.Context, disabledFilter *bool) (int32, error) {
	count, err := s.Store.Count(ctx, disabledFilter)
	if err != nil {
		return 0, fmt.Errorf("error counting users: %w", err)
	}
	return count, nil
}

// GetByExternalId retrieves a user by their alternate app ID (legacy int ID).
func (s *UserBusinessService) GetByExternalId(ctx context.Context, externalID int64) (*User, error) {
	user, err := s.Store.GetByAlternateAppID(ctx, externalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("%w: external_id %d", ErrUserNotFound, externalID)
		}
		return nil, fmt.Errorf("error retrieving user by external id: %w", err)
	}
	return user, nil
}
