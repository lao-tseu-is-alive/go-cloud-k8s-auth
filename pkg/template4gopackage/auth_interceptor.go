// Package template4gopackage provides Connect RPC authentication interceptor.
package template4gopackage

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
)

// Context keys for storing user information
type authContextKey string

const (
	userIDKey   authContextKey = "auth_user_id"
	isAdminKey  authContextKey = "auth_is_admin"
	userInfoKey authContextKey = "auth_user_info"
)

// NewAuthInterceptor AuthInterceptor creates a Connect unary interceptor that validates JWT tokens
// and injects user information into the request context.
//
// This interceptor:
// 1. Extracts the JWT from the Authorization header (Bearer token)
// 2. Validates the token using the provided JwtChecker
// 3. Injects userId and isAdmin into the context
// 4. Rejects unauthenticated requests with CodeUnauthenticated
//
// Usage:
//
//	interceptors := connect.WithInterceptors(NewAuthInterceptor(jwtCheck, log))
//	handler := template4gopackagev1connect.NewTemplate4ServiceNameServiceHandler(server, interceptors)
func NewAuthInterceptor(jwtCheck goHttpEcho.JwtChecker, log *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Extract Authorization header
			auth := req.Header().Get("Authorization")
			if auth == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing authorization header"))
			}

			// Extract Bearer token
			token := strings.TrimPrefix(auth, "Bearer ")
			if token == auth {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid authorization format"))
			}

			// Validate token and extract claims
			claims, err := jwtCheck.ParseToken(token)
			if err != nil {
				log.Warn("invalid JWT token", "error", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
			}

			// Inject user information into context
			ctx = context.WithValue(ctx, userIDKey, int32(claims.User.UserId))
			ctx = context.WithValue(ctx, isAdminKey, claims.User.IsAdmin)
			ctx = context.WithValue(ctx, userInfoKey, claims.User)

			// Call the actual handler with enriched context
			return next(ctx, req)
		}
	}
}

// GetUserFromContext extracts user information from the context.
// This should be called from RPC handlers after the AuthInterceptor has run.
//
// Returns userId (0 if not found) and isAdmin (false if not found).
func GetUserFromContext(ctx context.Context) (userId int32, isAdmin bool) {
	if id, ok := ctx.Value(userIDKey).(int32); ok {
		userId = id
	}
	if admin, ok := ctx.Value(isAdminKey).(bool); ok {
		isAdmin = admin
	}
	return
}

// GetUserInfoFromContext extracts the full UserInfo from the context.
// Returns nil if user info is not present in the context.
func GetUserInfoFromContext(ctx context.Context) *goHttpEcho.UserInfo {
	if info, ok := ctx.Value(userInfoKey).(*goHttpEcho.UserInfo); ok {
		return info
	}
	return nil
}

// MustGetUserFromContext extracts user information from the context.
// Panics if user info is not present (should only be used when interceptor is guaranteed to run).
func MustGetUserFromContext(ctx context.Context) (userId int32, isAdmin bool) {
	userId, isAdmin = GetUserFromContext(ctx)
	if userId == 0 {
		panic("MustGetUserFromContext called but no user in context - is AuthInterceptor configured?")
	}
	return
}
