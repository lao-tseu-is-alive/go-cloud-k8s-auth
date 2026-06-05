package template4gopackage

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

func (m *MockStorage) List(ctx context.Context, offset, limit int, params ListParams) ([]*Template4ServiceNameList, error) {
	args := m.Called(ctx, offset, limit, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Template4ServiceNameList), args.Error(1)
}

func (m *MockStorage) ListByExternalId(ctx context.Context, offset, limit int, externalId int) ([]*Template4ServiceNameList, error) {
	args := m.Called(ctx, offset, limit, externalId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Template4ServiceNameList), args.Error(1)
}

func (m *MockStorage) Search(ctx context.Context, offset, limit int, params SearchParams) ([]*Template4ServiceNameList, error) {
	args := m.Called(ctx, offset, limit, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Template4ServiceNameList), args.Error(1)
}

func (m *MockStorage) Get(ctx context.Context, id uuid.UUID) (*Template4ServiceName, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Template4ServiceName), args.Error(1)
}

func (m *MockStorage) Exist(ctx context.Context, id uuid.UUID) bool {
	args := m.Called(ctx, id)
	return args.Bool(0)
}

func (m *MockStorage) Count(ctx context.Context, params CountParams) (int32, error) {
	args := m.Called(ctx, params)
	return int32(args.Int(0)), args.Error(1)
}

func (m *MockStorage) Create(ctx context.Context, template_4_your_project_name Template4ServiceName) (*Template4ServiceName, error) {
	args := m.Called(ctx, template_4_your_project_name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Template4ServiceName), args.Error(1)
}

func (m *MockStorage) Update(ctx context.Context, id uuid.UUID, template_4_your_project_name Template4ServiceName) (*Template4ServiceName, error) {
	args := m.Called(ctx, id, template_4_your_project_name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Template4ServiceName), args.Error(1)
}

func (m *MockStorage) Delete(ctx context.Context, id uuid.UUID, userId int32) error {
	args := m.Called(ctx, id, userId)
	return args.Error(0)
}

func (m *MockStorage) IsTemplate4ServiceNameActive(ctx context.Context, id uuid.UUID) bool {
	args := m.Called(ctx, id)
	return args.Bool(0)
}

func (m *MockStorage) IsUserOwner(ctx context.Context, id uuid.UUID, userId int32) bool {
	args := m.Called(ctx, id, userId)
	return args.Bool(0)
}

func (m *MockStorage) CreateTypeTemplate4ServiceName(ctx context.Context, typeTemplate4ServiceName TypeTemplate4ServiceName) (*TypeTemplate4ServiceName, error) {
	args := m.Called(ctx, typeTemplate4ServiceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TypeTemplate4ServiceName), args.Error(1)
}

func (m *MockStorage) UpdateTypeTemplate4ServiceName(ctx context.Context, id int32, typeTemplate4ServiceName TypeTemplate4ServiceName) (*TypeTemplate4ServiceName, error) {
	args := m.Called(ctx, id, typeTemplate4ServiceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TypeTemplate4ServiceName), args.Error(1)
}

func (m *MockStorage) DeleteTypeTemplate4ServiceName(ctx context.Context, id int32, userId int32) error {
	args := m.Called(ctx, id, userId)
	return args.Error(0)
}

func (m *MockStorage) ListTypeTemplate4ServiceName(ctx context.Context, offset, limit int, params TypeTemplate4ServiceNameListParams) ([]*TypeTemplate4ServiceNameList, error) {
	args := m.Called(ctx, offset, limit, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*TypeTemplate4ServiceNameList), args.Error(1)
}

func (m *MockStorage) GetTypeTemplate4ServiceName(ctx context.Context, id int32) (*TypeTemplate4ServiceName, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TypeTemplate4ServiceName), args.Error(1)
}

func (m *MockStorage) CountTypeTemplate4ServiceName(ctx context.Context, params TypeTemplate4ServiceNameCountParams) (int32, error) {
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

		template_4_your_project_nameID := uuid.New()
		newTemplate4ServiceName := Template4ServiceName{
			Id:   template_4_your_project_nameID,
			Name: "Test Template4ServiceName",
		}

		expectedTemplate4ServiceName := newTemplate4ServiceName
		expectedTemplate4ServiceName.CreatedBy = 123

		// Mock TypeTemplate4ServiceName existence check
		mockDB.On("GetQueryInt", mock.Anytemplate_4_your_project_name, existTypeTemplate4ServiceName, []interface{}{newTemplate4ServiceName.TypeId}).Return(1, nil)
		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(false)
		mockStore.On("Create", mock.Anytemplate_4_your_project_name, mock.Anytemplate_4_your_project_nameOfType("Template4ServiceName")).Return(&expectedTemplate4ServiceName, nil)

		result, err := service.Create(ctx, 123, newTemplate4ServiceName)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(123), result.CreatedBy)
		mockStore.AssertExpectations(t)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTemplate4ServiceName := Template4ServiceName{
			Id:   uuid.New(),
			Name: "  ", // Empty/whitespace name
		}

		result, err := service.Create(ctx, 123, newTemplate4ServiceName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("validation error - name too short", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTemplate4ServiceName := Template4ServiceName{
			Id:   uuid.New(),
			Name: "ab", // Less than MinNameLength (5)
		}

		result, err := service.Create(ctx, 123, newTemplate4ServiceName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("validation error - invalid type id", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTemplate4ServiceName := Template4ServiceName{
			Id:     uuid.New(),
			Name:   "Test Template4ServiceName",
			TypeId: 999,
		}

		// Mock TypeTemplate4ServiceName existence check failure
		mockDB.On("GetQueryInt", mock.Anytemplate_4_your_project_name, existTypeTemplate4ServiceName, []interface{}{newTemplate4ServiceName.TypeId}).Return(0, nil)

		result, err := service.Create(ctx, 123, newTemplate4ServiceName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTypeTemplate4ServiceNameNotFound)
	})

	t.Run("already exists error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()
		newTemplate4ServiceName := Template4ServiceName{
			Id:   template_4_your_project_nameID,
			Name: "Test Template4ServiceName",
		}

		// Mock TypeTemplate4ServiceName existence check
		mockDB.On("GetQueryInt", mock.Anytemplate_4_your_project_name, existTypeTemplate4ServiceName, []interface{}{newTemplate4ServiceName.TypeId}).Return(1, nil)
		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)

		result, err := service.Create(ctx, 123, newTemplate4ServiceName)

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

		template_4_your_project_nameID := uuid.New()
		expectedTemplate4ServiceName := &Template4ServiceName{
			Id:   template_4_your_project_nameID,
			Name: "Test Template4ServiceName",
		}

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)
		mockStore.On("Get", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(expectedTemplate4ServiceName, nil)

		result, err := service.Get(ctx, template_4_your_project_nameID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Template4ServiceName", result.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("template_4_your_project_name not found", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()
		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(false)

		result, err := service.Get(ctx, template_4_your_project_nameID)

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

		template_4_your_project_nameID := uuid.New()
		userID := int32(123)
		updateTemplate4ServiceName := Template4ServiceName{
			Id:   template_4_your_project_nameID,
			Name: "Updated Template4ServiceName",
		}

		expectedTemplate4ServiceName := updateTemplate4ServiceName
		expectedTemplate4ServiceName.LastModifiedBy = &userID

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)
		mockStore.On("IsUserOwner", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, userID).Return(true)
		// Mock TypeTemplate4ServiceName existence check
		mockDB.On("GetQueryInt", mock.Anytemplate_4_your_project_name, existTypeTemplate4ServiceName, []interface{}{updateTemplate4ServiceName.TypeId}).Return(1, nil)
		mockStore.On("Update", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, mock.Anytemplate_4_your_project_nameOfType("Template4ServiceName")).Return(&expectedTemplate4ServiceName, nil)

		result, err := service.Update(ctx, userID, template_4_your_project_nameID, updateTemplate4ServiceName)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated Template4ServiceName", result.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("not owner error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()
		userID := int32(123)
		updateTemplate4ServiceName := Template4ServiceName{
			Id:   template_4_your_project_nameID,
			Name: "Updated Template4ServiceName",
		}

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)
		mockStore.On("IsUserOwner", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, userID).Return(false)

		result, err := service.Update(ctx, userID, template_4_your_project_nameID, updateTemplate4ServiceName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockStore.AssertExpectations(t)
	})
	t.Run("validation error - invalid type id", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()
		userID := int32(123)
		updateTemplate4ServiceName := Template4ServiceName{
			Id:     template_4_your_project_nameID,
			Name:   "Updated Template4ServiceName",
			TypeId: 999,
		}

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)
		mockStore.On("IsUserOwner", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, userID).Return(true)
		// Mock TypeTemplate4ServiceName existence check failure
		mockDB.On("GetQueryInt", mock.Anytemplate_4_your_project_name, existTypeTemplate4ServiceName, []interface{}{updateTemplate4ServiceName.TypeId}).Return(0, nil)

		result, err := service.Update(ctx, userID, template_4_your_project_nameID, updateTemplate4ServiceName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTypeTemplate4ServiceNameNotFound)
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

		template_4_your_project_nameID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)
		mockStore.On("IsUserOwner", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, userID).Return(true)
		mockStore.On("Delete", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, userID).Return(nil)

		err := service.Delete(ctx, userID, template_4_your_project_nameID)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("not owner error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)
		mockStore.On("IsUserOwner", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, userID).Return(false)

		err := service.Delete(ctx, userID, template_4_your_project_nameID)

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

		expectedList := []*Template4ServiceNameList{
			{Id: uuid.New(), Name: "Template4ServiceName 1"},
			{Id: uuid.New(), Name: "Template4ServiceName 2"},
		}
		params := ListParams{}

		mockStore.On("List", mock.Anytemplate_4_your_project_name, 0, 10, params).Return(expectedList, nil)

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
		mockStore.On("List", mock.Anytemplate_4_your_project_name, 0, 10, params).Return(nil, pgx.ErrNoRows)

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
		mockStore.On("List", mock.Anytemplate_4_your_project_name, 0, 10, params).Return(nil, dbError)

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
		mockStore.On("Count", mock.Anytemplate_4_your_project_name, params).Return(42, nil)

		result, err := service.Count(ctx, params)

		assert.NoError(t, err)
		assert.Equal(t, int32(42), result)
		mockStore.AssertExpectations(t)
	})
}

// Test CreateTypeTemplate4ServiceName operation
func TestBusinessService_CreateTypeTemplate4ServiceName(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation by admin", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTypeTemplate4ServiceName := TypeTemplate4ServiceName{
			Name: "Test Type",
		}

		expectedTypeTemplate4ServiceName := newTypeTemplate4ServiceName
		expectedTypeTemplate4ServiceName.Id = 1
		expectedTypeTemplate4ServiceName.CreatedBy = 123

		mockStore.On("CreateTypeTemplate4ServiceName", mock.Anytemplate_4_your_project_name, mock.Anytemplate_4_your_project_nameOfType("TypeTemplate4ServiceName")).Return(&expectedTypeTemplate4ServiceName, nil)

		result, err := service.CreateTypeTemplate4ServiceName(ctx, 123, true, newTypeTemplate4ServiceName)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(123), result.CreatedBy)
		mockStore.AssertExpectations(t)
	})

	t.Run("non-admin rejection", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTypeTemplate4ServiceName := TypeTemplate4ServiceName{
			Name: "Test Type",
		}

		result, err := service.CreateTypeTemplate4ServiceName(ctx, 123, false, newTypeTemplate4ServiceName)

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
