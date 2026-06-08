package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	authv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1/authv1connect"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-auth/pkg/version"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/gohttpclient"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/stretchr/testify/assert"
)

func TestHealthAndReadiness(t *testing.T) {
	// If DB is not configured, skip integration tests requiring the live server
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		t.Skip("Skipping integration test: DB_HOST not set")
	}

	l := golog.NewLogger("simple", os.Stdout, golog.DebugLevel, version.AppName)
	listenPort, _ := config.GetPort(defaultPort)
	listenAddr := fmt.Sprintf("http://localhost:%d", listenPort)

	// Start server in background
	go func() {
		main()
	}()

	// Wait for server to start
	gohttpclient.WaitForHttpServer(listenAddr, 100*time.Millisecond, 30, l)

	t.Run("GET /health", func(t *testing.T) {
		resp, err := http.Get(listenAddr + "/health")
		if err != nil {
			t.Fatalf("Failed to request health: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /readiness", func(t *testing.T) {
		resp, err := http.Get(listenAddr + "/readiness")
		if err != nil {
			t.Fatalf("Failed to request readiness: %v", err)
		}
		defer resp.Body.Close()
		// Might return 503 if database connection isn't fully ready, but should respond
		assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, resp.StatusCode)
	})

	t.Run("GET /goAppInfo", func(t *testing.T) {
		resp, err := http.Get(listenAddr + defaultAppInfoUrl)
		if err != nil {
			t.Fatalf("Failed to request app info: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ConnectRPC end-to-end tests", func(t *testing.T) {
		// 1. Create a JWT checker to sign test tokens
		jwtCheck, err := goHttpEcho.GetNewJwtCheckerFromConfig(version.AppName, 60, l)
		if err != nil {
			t.Fatalf("Failed to create JWT checker: %v", err)
		}

		// 2. Generate a valid admin JWT token
		adminUserInfo := &goHttpEcho.UserInfo{
			UserId:     999,
			ExternalId: 999,
			Name:       "Test Admin",
			Email:      "admin@test.org",
			Login:      "admin@test.org",
			IsAdmin:    true,
		}
		adminToken, err := jwtCheck.GetTokenFromUserInfo(adminUserInfo)
		if err != nil {
			t.Fatalf("Failed to generate admin token: %v", err)
		}

		// 3. Generate a regular user JWT token
		userUserInfo := &goHttpEcho.UserInfo{
			UserId:     100,
			ExternalId: 100,
			Name:       "Test User",
			Email:      "user@test.org",
			Login:      "user@test.org",
			IsAdmin:    false,
		}
		userToken, err := jwtCheck.GetTokenFromUserInfo(userUserInfo)
		if err != nil {
			t.Fatalf("Failed to generate user token: %v", err)
		}

		// 4. ConnectRPC Clients with appropriate authorization interceptors
		adminInterceptor := connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				req.Header().Set("Authorization", "Bearer "+adminToken.String())
				return next(ctx, req)
			}
		}))

		userInterceptor := connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				req.Header().Set("Authorization", "Bearer "+userToken.String())
				return next(ctx, req)
			}
		}))

		authClient := authv1connect.NewAuthServiceClient(http.DefaultClient, listenAddr)
		adminUserClient := authv1connect.NewUserServiceClient(http.DefaultClient, listenAddr, adminInterceptor)
		regularUserClient := authv1connect.NewUserServiceClient(http.DefaultClient, listenAddr, userInterceptor)

		// Test Token Validation with invalid token
		t.Run("ValidateToken - Invalid", func(t *testing.T) {
			resp, err := authClient.ValidateToken(context.Background(), connect.NewRequest(&authv1.ValidateTokenRequest{
				Token: "invalid-token-value",
			}))
			assert.NoError(t, err)
			assert.False(t, resp.Msg.Valid)
		})

		var createdUser *authv1.User
		t.Run("Create User - Admin allowed", func(t *testing.T) {
			newUser := &authv1.User{
				Email:    "new-user-e2e@test.org",
				Name:     "E2E New User",
				Roles:    []string{"user"},
				Disabled: false,
			}
			resp, err := adminUserClient.Create(context.Background(), connect.NewRequest(&authv1.CreateRequest{
				User: newUser,
			}))
			assert.NoError(t, err)
			assert.NotNil(t, resp.Msg.User)
			assert.NotEmpty(t, resp.Msg.User.Id)
			assert.Equal(t, "new-user-e2e@test.org", resp.Msg.User.Email)
			createdUser = resp.Msg.User
		})

		t.Run("Create User - Regular user forbidden", func(t *testing.T) {
			newUser := &authv1.User{
				Email:    "new-user-e2e-2@test.org",
				Name:     "E2E New User 2",
				Roles:    []string{"user"},
				Disabled: false,
			}
			_, err := regularUserClient.Create(context.Background(), connect.NewRequest(&authv1.CreateRequest{
				User: newUser,
			}))
			assert.Error(t, err)
			connectErr := &connect.Error{}
			if assert.ErrorAs(t, err, &connectErr) {
				assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
			}
		})

		t.Run("List Users", func(t *testing.T) {
			resp, err := regularUserClient.List(context.Background(), connect.NewRequest(&authv1.ListRequest{
				Limit: 10,
			}))
			assert.NoError(t, err)
			assert.NotEmpty(t, resp.Msg.Users)
		})

		t.Run("Delete User - Regular user forbidden", func(t *testing.T) {
			_, err := regularUserClient.Delete(context.Background(), connect.NewRequest(&authv1.DeleteRequest{
				Id: createdUser.Id,
			}))
			assert.Error(t, err)
			connectErr := &connect.Error{}
			if assert.ErrorAs(t, err, &connectErr) {
				assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
			}
		})

		t.Run("Delete User - Admin allowed", func(t *testing.T) {
			_, err := adminUserClient.Delete(context.Background(), connect.NewRequest(&authv1.DeleteRequest{
				Id: createdUser.Id,
			}))
			assert.NoError(t, err)
		})
	})
}
