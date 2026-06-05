package template4gopackage

import (
	"context"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	template_4_your_project_namev1 "github.com/your-github-account/template-4-your-project-name/gen/template_4_your_project_name/v1"
)

// =============================================================================
// Test Helpers
// =============================================================================

// Helper to create a test Connect server
func createTestTemplate4ServiceNameConnectServer(mockStore *MockStorage, mockDB *MockDB) *Template4ServiceNameConnectServer {
	logger := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	businessService := NewBusinessService(mockStore, mockDB, logger, 50)
	return NewTemplate4ServiceNameConnectServer(businessService, logger)
}

// Helper to create a test TypeTemplate4ServiceName Connect server
func createTestTypeTemplate4ServiceNameConnectServer(mockStore *MockStorage, mockDB *MockDB) *TypeTemplate4ServiceNameConnectServer {
	logger := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	businessService := NewBusinessService(mockStore, mockDB, logger, 50)
	return NewTypeTemplate4ServiceNameConnectServer(businessService, logger)
}

// Helper to create a context with user info (simulating what AuthInterceptor does)
func contextWithUser(userId int32, isAdmin bool) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, userId)
	ctx = context.WithValue(ctx, isAdminKey, isAdmin)
	return ctx
}

// Helper to create a Connect request (no auth header needed since we inject via context)
func createConnectRequest[T any](msg *T) *connect.Request[T] {
	return connect.NewRequest(msg)
}

// =============================================================================
// Template4ServiceNameConnectServer Tests
// =============================================================================

func TestTemplate4ServiceNameConnectServer_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		// Setup mock storage
		now := time.Now()
		expectedList := []*Template4ServiceNameList{
			{Id: uuid.New(), Name: "Template4ServiceName 1", CreatedAt: &now},
			{Id: uuid.New(), Name: "Template4ServiceName 2", CreatedAt: &now},
		}
		mockStore.On("List", mock.Anytemplate_4_your_project_name, 0, 50, ListParams{}).Return(expectedList, nil)

		// Create request and context with user
		req := createConnectRequest(&template_4_your_project_namev1.ListRequest{Limit: 0, Offset: 0})
		ctx := contextWithUser(123, false)

		// Call handler
		resp, err := server.List(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.Template4ServiceNames, 2)
		assert.Equal(t, "Template4ServiceName 1", resp.Msg.Template4ServiceNames[0].Name)
		assert.Equal(t, "Template4ServiceName 2", resp.Msg.Template4ServiceNames[1].Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("list with pagination", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		now := time.Now()
		expectedList := []*Template4ServiceNameList{
			{Id: uuid.New(), Name: "Template4ServiceName 3", CreatedAt: &now},
		}
		mockStore.On("List", mock.Anytemplate_4_your_project_name, 10, 5, ListParams{}).Return(expectedList, nil)

		req := createConnectRequest(&template_4_your_project_namev1.ListRequest{Limit: 5, Offset: 10})
		ctx := contextWithUser(123, false)

		resp, err := server.List(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.Template4ServiceNames, 1)
		mockStore.AssertExpectations(t)
	})
}

func TestTemplate4ServiceNameConnectServer_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()
		expectedTemplate4ServiceName := &Template4ServiceName{
			Id:   template_4_your_project_nameID,
			Name: "Test Template4ServiceName",
		}

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)
		mockStore.On("Get", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(expectedTemplate4ServiceName, nil)

		req := createConnectRequest(&template_4_your_project_namev1.GetRequest{Id: template_4_your_project_nameID.String()})
		ctx := contextWithUser(123, false)

		resp, err := server.Get(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, template_4_your_project_nameID.String(), resp.Msg.Template4ServiceName.Id)
		assert.Equal(t, "Test Template4ServiceName", resp.Msg.Template4ServiceName.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(false)

		req := createConnectRequest(&template_4_your_project_namev1.GetRequest{Id: template_4_your_project_nameID.String()})
		ctx := contextWithUser(123, false)

		resp, err := server.Get(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
		mockStore.AssertExpectations(t)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		req := createConnectRequest(&template_4_your_project_namev1.GetRequest{Id: "not-a-uuid"})
		ctx := contextWithUser(123, false)

		resp, err := server.Get(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	})
}

func TestTemplate4ServiceNameConnectServer_Create(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()
		expectedTemplate4ServiceName := &Template4ServiceName{
			Id:        template_4_your_project_nameID,
			Name:      "New Template4ServiceName",
			CreatedBy: 123,
		}

		mockDB.On("GetQueryInt", mock.Anytemplate_4_your_project_name, existTypeTemplate4ServiceName, mock.Anytemplate_4_your_project_name).Return(1, nil)
		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, mock.Anytemplate_4_your_project_nameOfType("uuid.UUID")).Return(false)
		mockStore.On("Create", mock.Anytemplate_4_your_project_name, mock.Anytemplate_4_your_project_nameOfType("Template4ServiceName")).Return(expectedTemplate4ServiceName, nil)

		req := createConnectRequest(&template_4_your_project_namev1.CreateRequest{
			Template4ServiceName: &template_4_your_project_namev1.Template4ServiceName{
				Name: "New Template4ServiceName",
			},
		})
		ctx := contextWithUser(123, false)

		resp, err := server.Create(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Template4ServiceName", resp.Msg.Template4ServiceName.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("validation error - missing template_4_your_project_name", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		req := createConnectRequest(&template_4_your_project_namev1.CreateRequest{Template4ServiceName: nil})
		ctx := contextWithUser(123, false)

		resp, err := server.Create(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	})
}

func TestTemplate4ServiceNameConnectServer_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)
		mockStore.On("IsUserOwner", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, userID).Return(true)
		mockStore.On("Delete", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, userID).Return(nil)

		req := createConnectRequest(&template_4_your_project_namev1.DeleteRequest{Id: template_4_your_project_nameID.String()})
		ctx := contextWithUser(userID, false)

		resp, err := server.Delete(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockStore.AssertExpectations(t)
	})

	t.Run("permission denied - not owner", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		template_4_your_project_nameID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID).Return(true)
		mockStore.On("IsUserOwner", mock.Anytemplate_4_your_project_name, template_4_your_project_nameID, userID).Return(false)

		req := createConnectRequest(&template_4_your_project_namev1.DeleteRequest{Id: template_4_your_project_nameID.String()})
		ctx := contextWithUser(userID, false)

		resp, err := server.Delete(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
		mockStore.AssertExpectations(t)
	})
}

func TestTemplate4ServiceNameConnectServer_Count(t *testing.T) {
	t.Run("successful count", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTemplate4ServiceNameConnectServer(mockStore, mockDB)

		mockStore.On("Count", mock.Anytemplate_4_your_project_name, CountParams{}).Return(42, nil)

		req := createConnectRequest(&template_4_your_project_namev1.CountRequest{})
		ctx := contextWithUser(123, false)

		resp, err := server.Count(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(42), resp.Msg.Count)
		mockStore.AssertExpectations(t)
	})
}

// =============================================================================
// TypeTemplate4ServiceNameConnectServer Tests
// =============================================================================

func TestTypeTemplate4ServiceNameConnectServer_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTypeTemplate4ServiceNameConnectServer(mockStore, mockDB)

		now := time.Now()
		expectedList := []*TypeTemplate4ServiceNameList{
			{Id: 1, Name: "Type 1", CreatedAt: now},
			{Id: 2, Name: "Type 2", CreatedAt: now},
		}

		mockStore.On("ListTypeTemplate4ServiceName", mock.Anytemplate_4_your_project_name, 0, 250, TypeTemplate4ServiceNameListParams{}).Return(expectedList, nil)

		req := createConnectRequest(&template_4_your_project_namev1.TypeTemplate4ServiceNameListRequest{})
		ctx := contextWithUser(123, false)

		resp, err := server.List(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.TypeTemplate4ServiceNames, 2)
		mockStore.AssertExpectations(t)
	})
}

func TestTypeTemplate4ServiceNameConnectServer_Create(t *testing.T) {
	t.Run("admin can create", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTypeTemplate4ServiceNameConnectServer(mockStore, mockDB)

		expectedTypeTemplate4ServiceName := &TypeTemplate4ServiceName{
			Id:        1,
			Name:      "New Type",
			CreatedBy: 123,
		}

		mockStore.On("CreateTypeTemplate4ServiceName", mock.Anytemplate_4_your_project_name, mock.Anytemplate_4_your_project_nameOfType("TypeTemplate4ServiceName")).Return(expectedTypeTemplate4ServiceName, nil)

		req := createConnectRequest(&template_4_your_project_namev1.TypeTemplate4ServiceNameCreateRequest{
			TypeTemplate4ServiceName: &template_4_your_project_namev1.TypeTemplate4ServiceName{
				Name: "New Type",
			},
		})
		ctx := contextWithUser(123, true) // isAdmin = true

		resp, err := server.Create(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Type", resp.Msg.TypeTemplate4ServiceName.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("non-admin rejected", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTypeTemplate4ServiceNameConnectServer(mockStore, mockDB)

		req := createConnectRequest(&template_4_your_project_namev1.TypeTemplate4ServiceNameCreateRequest{
			TypeTemplate4ServiceName: &template_4_your_project_namev1.TypeTemplate4ServiceName{
				Name: "New Type",
			},
		})
		ctx := contextWithUser(123, false) // isAdmin = false

		resp, err := server.Create(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})
}

// =============================================================================
// AuthInterceptor Tests
// =============================================================================

func TestGetUserFromContext(t *testing.T) {
	t.Run("user present in context", func(t *testing.T) {
		ctx := contextWithUser(456, true)

		userId, isAdmin := GetUserFromContext(ctx)

		assert.Equal(t, int32(456), userId)
		assert.True(t, isAdmin)
	})

	t.Run("user not present in context", func(t *testing.T) {
		ctx := context.Background()

		userId, isAdmin := GetUserFromContext(ctx)

		assert.Equal(t, int32(0), userId)
		assert.False(t, isAdmin)
	})
}
