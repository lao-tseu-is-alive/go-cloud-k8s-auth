package auth

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the Storage interface for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GeoJson(ctx context.Context, offset, limit int, params GeoJsonParams) (string, error) {
	args := m.Called(ctx, offset, limit, params)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) List(ctx context.Context, offset, limit int, params ListParams) ([]*AuthList, error) {
	args := m.Called(ctx, offset, limit, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AuthList), args.Error(1)
}

func (m *MockStorage) ListByExternalId(ctx context.Context, offset, limit int, externalId int) ([]*AuthList, error) {
	args := m.Called(ctx, offset, limit, externalId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AuthList), args.Error(1)
}

func (m *MockStorage) Search(ctx context.Context, offset, limit int, params SearchParams) ([]*AuthList, error) {
	args := m.Called(ctx, offset, limit, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AuthList), args.Error(1)
}

func (m *MockStorage) Get(ctx context.Context, id uuid.UUID) (*Auth, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Auth), args.Error(1)
}

func (m *MockStorage) Exist(ctx context.Context, id uuid.UUID) bool {
	args := m.Called(ctx, id)
	return args.Bool(0)
}

func (m *MockStorage) Count(ctx context.Context, params CountParams) (int32, error) {
	args := m.Called(ctx, params)
	return int32(args.Int(0)), args.Error(1)
}

func (m *MockStorage) Create(ctx context.Context, go_cloud_auth Auth) (*Auth, error) {
	args := m.Called(ctx, go_cloud_auth)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Auth), args.Error(1)
}

func (m *MockStorage) Update(ctx context.Context, id uuid.UUID, go_cloud_auth Auth) (*Auth, error) {
	args := m.Called(ctx, id, go_cloud_auth)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Auth), args.Error(1)
}

func (m *MockStorage) Delete(ctx context.Context, id uuid.UUID, userId int32) error {
	args := m.Called(ctx, id, userId)
	return args.Error(0)
}

func (m *MockStorage) IsAuthActive(ctx context.Context, id uuid.UUID) bool {
	args := m.Called(ctx, id)
	return args.Bool(0)
}

func (m *MockStorage) IsUserOwner(ctx context.Context, id uuid.UUID, userId int32) bool {
	args := m.Called(ctx, id, userId)
	return args.Bool(0)
}

func (m *MockStorage) CreateTypeAuth(ctx context.Context, typeAuth TypeAuth) (*TypeAuth, error) {
	args := m.Called(ctx, typeAuth)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TypeAuth), args.Error(1)
}

func (m *MockStorage) UpdateTypeAuth(ctx context.Context, id int32, typeAuth TypeAuth) (*TypeAuth, error) {
	args := m.Called(ctx, id, typeAuth)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TypeAuth), args.Error(1)
}

func (m *MockStorage) DeleteTypeAuth(ctx context.Context, id int32, userId int32) error {
	args := m.Called(ctx, id, userId)
	return args.Error(0)
}

func (m *MockStorage) ListTypeAuth(ctx context.Context, offset, limit int, params TypeAuthListParams) ([]*TypeAuthList, error) {
	args := m.Called(ctx, offset, limit, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*TypeAuthList), args.Error(1)
}

func (m *MockStorage) GetTypeAuth(ctx context.Context, id int32) (*TypeAuth, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TypeAuth), args.Error(1)
}

func (m *MockStorage) CountTypeAuth(ctx context.Context, params TypeAuthCountParams) (int32, error) {
	args := m.Called(ctx, params)
	return int32(args.Int(0)), args.Error(1)
}

// MockDB is a minimal mock for database connection
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetQueryInt(ctx context.Context, query string, args ...interface{}) (int, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Int(0), callArgs.Error(1)
}

func (m *MockDB) GetVersion(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockDB) Close() {
	m.Called()
}

func (m *MockDB) HealthCheck(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockDB) GetQueryBool(ctx context.Context, query string, args ...interface{}) (bool, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Bool(0), callArgs.Error(1)
}

func (m *MockDB) ExecActionQuery(ctx context.Context, query string, args ...interface{}) (int, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Int(0), callArgs.Error(1)
}

func (m *MockDB) DoesTableExist(ctx context.Context, schema, table string) bool {
	args := m.Called(ctx, schema, table)
	return args.Bool(0)
}

func (m *MockDB) GetPGConn() (*pgxpool.Pool, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pgxpool.Pool), args.Error(1)
}

func (m *MockDB) GetQueryString(ctx context.Context, query string, args ...interface{}) (string, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.String(0), callArgs.Error(1)
}

func (m *MockDB) Insert(ctx context.Context, query string, args ...interface{}) (int, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Int(0), callArgs.Error(1)
}

// Helper function to create a test business service
func createTestBusinessService(mockStore *MockStorage, mockDB *MockDB) *BusinessService {
	logger := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	return NewBusinessService(mockStore, mockDB, logger, 50)
}

// Test Create operation
func TestBusinessService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		newAuth := Auth{
			Id:   go_cloud_authID,
			Name: "Test Auth",
		}

		expectedAuth := newAuth
		expectedAuth.CreatedBy = 123

		// Mock TypeAuth existence check
		mockDB.On("GetQueryInt", mock.Anygo_cloud_auth, existTypeAuth, []interface{}{newAuth.TypeId}).Return(1, nil)
		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(false)
		mockStore.On("Create", mock.Anygo_cloud_auth, mock.Anygo_cloud_authOfType("Auth")).Return(&expectedAuth, nil)

		result, err := service.Create(ctx, 123, newAuth)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(123), result.CreatedBy)
		mockStore.AssertExpectations(t)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newAuth := Auth{
			Id:   uuid.New(),
			Name: "  ", // Empty/whitespace name
		}

		result, err := service.Create(ctx, 123, newAuth)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("validation error - name too short", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newAuth := Auth{
			Id:   uuid.New(),
			Name: "ab", // Less than MinNameLength (5)
		}

		result, err := service.Create(ctx, 123, newAuth)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("validation error - invalid type id", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newAuth := Auth{
			Id:     uuid.New(),
			Name:   "Test Auth",
			TypeId: 999,
		}

		// Mock TypeAuth existence check failure
		mockDB.On("GetQueryInt", mock.Anygo_cloud_auth, existTypeAuth, []interface{}{newAuth.TypeId}).Return(0, nil)

		result, err := service.Create(ctx, 123, newAuth)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTypeAuthNotFound)
	})

	t.Run("already exists error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		newAuth := Auth{
			Id:   go_cloud_authID,
			Name: "Test Auth",
		}

		// Mock TypeAuth existence check
		mockDB.On("GetQueryInt", mock.Anygo_cloud_auth, existTypeAuth, []interface{}{newAuth.TypeId}).Return(1, nil)
		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)

		result, err := service.Create(ctx, 123, newAuth)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrAlreadyExists)
		mockStore.AssertExpectations(t)
	})
}

// Test Get operation
func TestBusinessService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		expectedAuth := &Auth{
			Id:   go_cloud_authID,
			Name: "Test Auth",
		}

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)
		mockStore.On("Get", mock.Anygo_cloud_auth, go_cloud_authID).Return(expectedAuth, nil)

		result, err := service.Get(ctx, go_cloud_authID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Auth", result.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("go_cloud_auth not found", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(false)

		result, err := service.Get(ctx, go_cloud_authID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrNotFound)
		mockStore.AssertExpectations(t)
	})
}

// Test Update operation
func TestBusinessService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		userID := int32(123)
		updateAuth := Auth{
			Id:   go_cloud_authID,
			Name: "Updated Auth",
		}

		expectedAuth := updateAuth
		expectedAuth.LastModifiedBy = &userID

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)
		mockStore.On("IsUserOwner", mock.Anygo_cloud_auth, go_cloud_authID, userID).Return(true)
		// Mock TypeAuth existence check
		mockDB.On("GetQueryInt", mock.Anygo_cloud_auth, existTypeAuth, []interface{}{updateAuth.TypeId}).Return(1, nil)
		mockStore.On("Update", mock.Anygo_cloud_auth, go_cloud_authID, mock.Anygo_cloud_authOfType("Auth")).Return(&expectedAuth, nil)

		result, err := service.Update(ctx, userID, go_cloud_authID, updateAuth)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated Auth", result.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("not owner error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		userID := int32(123)
		updateAuth := Auth{
			Id:   go_cloud_authID,
			Name: "Updated Auth",
		}

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)
		mockStore.On("IsUserOwner", mock.Anygo_cloud_auth, go_cloud_authID, userID).Return(false)

		result, err := service.Update(ctx, userID, go_cloud_authID, updateAuth)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockStore.AssertExpectations(t)
	})
	t.Run("validation error - invalid type id", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		userID := int32(123)
		updateAuth := Auth{
			Id:     go_cloud_authID,
			Name:   "Updated Auth",
			TypeId: 999,
		}

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)
		mockStore.On("IsUserOwner", mock.Anygo_cloud_auth, go_cloud_authID, userID).Return(true)
		// Mock TypeAuth existence check failure
		mockDB.On("GetQueryInt", mock.Anygo_cloud_auth, existTypeAuth, []interface{}{updateAuth.TypeId}).Return(0, nil)

		result, err := service.Update(ctx, userID, go_cloud_authID, updateAuth)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTypeAuthNotFound)
		mockStore.AssertExpectations(t)
	})
}

// Test Delete operation
func TestBusinessService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)
		mockStore.On("IsUserOwner", mock.Anygo_cloud_auth, go_cloud_authID, userID).Return(true)
		mockStore.On("Delete", mock.Anygo_cloud_auth, go_cloud_authID, userID).Return(nil)

		err := service.Delete(ctx, userID, go_cloud_authID)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("not owner error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)
		mockStore.On("IsUserOwner", mock.Anygo_cloud_auth, go_cloud_authID, userID).Return(false)

		err := service.Delete(ctx, userID, go_cloud_authID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockStore.AssertExpectations(t)
	})
}

// Test List operation
func TestBusinessService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		expectedList := []*AuthList{
			{Id: uuid.New(), Name: "Auth 1"},
			{Id: uuid.New(), Name: "Auth 2"},
		}
		params := ListParams{}

		mockStore.On("List", mock.Anygo_cloud_auth, 0, 10, params).Return(expectedList, nil)

		result, err := service.List(ctx, 0, 10, params)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockStore.AssertExpectations(t)
	})

	t.Run("empty list with pgx.ErrNoRows", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		params := ListParams{}
		mockStore.On("List", mock.Anygo_cloud_auth, 0, 10, params).Return(nil, pgx.ErrNoRows)

		result, err := service.List(ctx, 0, 10, params)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockStore.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		params := ListParams{}
		dbError := errors.New("database connection failed")
		mockStore.On("List", mock.Anygo_cloud_auth, 0, 10, params).Return(nil, dbError)

		result, err := service.List(ctx, 0, 10, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
	})
}

// Test Count operation
func TestBusinessService_Count(t *testing.T) {
	ctx := context.Background()

	t.Run("successful count", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		params := CountParams{}
		mockStore.On("Count", mock.Anygo_cloud_auth, params).Return(42, nil)

		result, err := service.Count(ctx, params)

		assert.NoError(t, err)
		assert.Equal(t, int32(42), result)
		mockStore.AssertExpectations(t)
	})
}

// Test CreateTypeAuth operation
func TestBusinessService_CreateTypeAuth(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation by admin", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTypeAuth := TypeAuth{
			Name: "Test Type",
		}

		expectedTypeAuth := newTypeAuth
		expectedTypeAuth.Id = 1
		expectedTypeAuth.CreatedBy = 123

		mockStore.On("CreateTypeAuth", mock.Anygo_cloud_auth, mock.Anygo_cloud_authOfType("TypeAuth")).Return(&expectedTypeAuth, nil)

		result, err := service.CreateTypeAuth(ctx, 123, true, newTypeAuth)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(123), result.CreatedBy)
		mockStore.AssertExpectations(t)
	})

	t.Run("non-admin rejection", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTypeAuth := TypeAuth{
			Name: "Test Type",
		}

		result, err := service.CreateTypeAuth(ctx, 123, false, newTypeAuth)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrAdminRequired)
	})
}

// Test validation function
func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid name", "Valid Name", false},
		{"empty string", "", true},
		{"only spaces", "   ", true},
		{"too short", "ab", true},
		{"exactly min length", "12345", false},
		{"longer than min", "Long Enough Name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
