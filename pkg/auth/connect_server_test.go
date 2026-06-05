package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	go_cloud_authv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/go_cloud_auth/v1"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// Test Helpers
// =============================================================================

// Helper to create a test Connect server
func createTestAuthConnectServer(mockStore *MockStorage, mockDB *MockDB) *AuthConnectServer {
	logger := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	businessService := NewBusinessService(mockStore, mockDB, logger, 50)
	return NewAuthConnectServer(businessService, logger)
}

// Helper to create a test TypeAuth Connect server
func createTestTypeAuthConnectServer(mockStore *MockStorage, mockDB *MockDB) *TypeAuthConnectServer {
	logger := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	businessService := NewBusinessService(mockStore, mockDB, logger, 50)
	return NewTypeAuthConnectServer(businessService, logger)
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
// AuthConnectServer Tests
// =============================================================================

func TestAuthConnectServer_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestAuthConnectServer(mockStore, mockDB)

		// Setup mock storage
		now := time.Now()
		expectedList := []*AuthList{
			{Id: uuid.New(), Name: "Auth 1", CreatedAt: &now},
			{Id: uuid.New(), Name: "Auth 2", CreatedAt: &now},
		}
		mockStore.On("List", mock.Anygo_cloud_auth, 0, 50, ListParams{}).Return(expectedList, nil)

		// Create request and context with user
		req := createConnectRequest(&go_cloud_authv1.ListRequest{Limit: 0, Offset: 0})
		ctx := contextWithUser(123, false)

		// Call handler
		resp, err := server.List(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.Auths, 2)
		assert.Equal(t, "Auth 1", resp.Msg.Auths[0].Name)
		assert.Equal(t, "Auth 2", resp.Msg.Auths[1].Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("list with pagination", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestAuthConnectServer(mockStore, mockDB)

		now := time.Now()
		expectedList := []*AuthList{
			{Id: uuid.New(), Name: "Auth 3", CreatedAt: &now},
		}
		mockStore.On("List", mock.Anygo_cloud_auth, 10, 5, ListParams{}).Return(expectedList, nil)

		req := createConnectRequest(&go_cloud_authv1.ListRequest{Limit: 5, Offset: 10})
		ctx := contextWithUser(123, false)

		resp, err := server.List(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.Auths, 1)
		mockStore.AssertExpectations(t)
	})
}

func TestAuthConnectServer_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestAuthConnectServer(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		expectedAuth := &Auth{
			Id:   go_cloud_authID,
			Name: "Test Auth",
		}

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)
		mockStore.On("Get", mock.Anygo_cloud_auth, go_cloud_authID).Return(expectedAuth, nil)

		req := createConnectRequest(&go_cloud_authv1.GetRequest{Id: go_cloud_authID.String()})
		ctx := contextWithUser(123, false)

		resp, err := server.Get(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, go_cloud_authID.String(), resp.Msg.Auth.Id)
		assert.Equal(t, "Test Auth", resp.Msg.Auth.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestAuthConnectServer(mockStore, mockDB)

		go_cloud_authID := uuid.New()

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(false)

		req := createConnectRequest(&go_cloud_authv1.GetRequest{Id: go_cloud_authID.String()})
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
		server := createTestAuthConnectServer(mockStore, mockDB)

		req := createConnectRequest(&go_cloud_authv1.GetRequest{Id: "not-a-uuid"})
		ctx := contextWithUser(123, false)

		resp, err := server.Get(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	})
}

func TestAuthConnectServer_Create(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestAuthConnectServer(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		expectedAuth := &Auth{
			Id:        go_cloud_authID,
			Name:      "New Auth",
			CreatedBy: 123,
		}

		mockDB.On("GetQueryInt", mock.Anygo_cloud_auth, existTypeAuth, mock.Anygo_cloud_auth).Return(1, nil)
		mockStore.On("Exist", mock.Anygo_cloud_auth, mock.Anygo_cloud_authOfType("uuid.UUID")).Return(false)
		mockStore.On("Create", mock.Anygo_cloud_auth, mock.Anygo_cloud_authOfType("Auth")).Return(expectedAuth, nil)

		req := createConnectRequest(&go_cloud_authv1.CreateRequest{
			Auth: &go_cloud_authv1.Auth{
				Name: "New Auth",
			},
		})
		ctx := contextWithUser(123, false)

		resp, err := server.Create(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Auth", resp.Msg.Auth.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("validation error - missing go_cloud_auth", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestAuthConnectServer(mockStore, mockDB)

		req := createConnectRequest(&go_cloud_authv1.CreateRequest{Auth: nil})
		ctx := contextWithUser(123, false)

		resp, err := server.Create(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	})
}

func TestAuthConnectServer_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestAuthConnectServer(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)
		mockStore.On("IsUserOwner", mock.Anygo_cloud_auth, go_cloud_authID, userID).Return(true)
		mockStore.On("Delete", mock.Anygo_cloud_auth, go_cloud_authID, userID).Return(nil)

		req := createConnectRequest(&go_cloud_authv1.DeleteRequest{Id: go_cloud_authID.String()})
		ctx := contextWithUser(userID, false)

		resp, err := server.Delete(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockStore.AssertExpectations(t)
	})

	t.Run("permission denied - not owner", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestAuthConnectServer(mockStore, mockDB)

		go_cloud_authID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anygo_cloud_auth, go_cloud_authID).Return(true)
		mockStore.On("IsUserOwner", mock.Anygo_cloud_auth, go_cloud_authID, userID).Return(false)

		req := createConnectRequest(&go_cloud_authv1.DeleteRequest{Id: go_cloud_authID.String()})
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

func TestAuthConnectServer_Count(t *testing.T) {
	t.Run("successful count", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestAuthConnectServer(mockStore, mockDB)

		mockStore.On("Count", mock.Anygo_cloud_auth, CountParams{}).Return(42, nil)

		req := createConnectRequest(&go_cloud_authv1.CountRequest{})
		ctx := contextWithUser(123, false)

		resp, err := server.Count(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(42), resp.Msg.Count)
		mockStore.AssertExpectations(t)
	})
}

// =============================================================================
// TypeAuthConnectServer Tests
// =============================================================================

func TestTypeAuthConnectServer_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTypeAuthConnectServer(mockStore, mockDB)

		now := time.Now()
		expectedList := []*TypeAuthList{
			{Id: 1, Name: "Type 1", CreatedAt: now},
			{Id: 2, Name: "Type 2", CreatedAt: now},
		}

		mockStore.On("ListTypeAuth", mock.Anygo_cloud_auth, 0, 250, TypeAuthListParams{}).Return(expectedList, nil)

		req := createConnectRequest(&go_cloud_authv1.TypeAuthListRequest{})
		ctx := contextWithUser(123, false)

		resp, err := server.List(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.TypeAuths, 2)
		mockStore.AssertExpectations(t)
	})
}

func TestTypeAuthConnectServer_Create(t *testing.T) {
	t.Run("admin can create", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTypeAuthConnectServer(mockStore, mockDB)

		expectedTypeAuth := &TypeAuth{
			Id:        1,
			Name:      "New Type",
			CreatedBy: 123,
		}

		mockStore.On("CreateTypeAuth", mock.Anygo_cloud_auth, mock.Anygo_cloud_authOfType("TypeAuth")).Return(expectedTypeAuth, nil)

		req := createConnectRequest(&go_cloud_authv1.TypeAuthCreateRequest{
			TypeAuth: &go_cloud_authv1.TypeAuth{
				Name: "New Type",
			},
		})
		ctx := contextWithUser(123, true) // isAdmin = true

		resp, err := server.Create(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Type", resp.Msg.TypeAuth.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("non-admin rejected", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTypeAuthConnectServer(mockStore, mockDB)

		req := createConnectRequest(&go_cloud_authv1.TypeAuthCreateRequest{
			TypeAuth: &go_cloud_authv1.TypeAuth{
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
