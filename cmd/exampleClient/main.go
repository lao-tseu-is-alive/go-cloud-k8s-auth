// Package main implements a simple command-line client that acts as a full example
// of how other services can integrate with and consume the go-cloud-k8s-auth service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	authv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1/authv1connect"
)

// NewAuthInterceptor creates a simple client-side Connect interceptor
// that injects the JWT token into the "Authorization: Bearer <token>" header.
func NewAuthInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if token != "" {
				req.Header().Set("Authorization", "Bearer "+token)
			}
			return next(ctx, req)
		}
	}
}

func main() {
	serverURL := flag.String("server", "http://localhost:9090", "The address of the go-cloud-k8s-auth server")
	cmd := flag.String("cmd", "start-oauth", "The command to run: start-oauth, validate, current-user, list-users")
	provider := flag.String("provider", "github", "The OAuth provider (used by start-oauth): google, github, microsoft")
	token := flag.String("token", "", "JWT token (required for validate, current-user, list-users)")
	limit := flag.Int("limit", 10, "Limit for list-users command")
	offset := flag.Int("offset", 0, "Offset for list-users command")
	flag.Parse()

	// 1. Setup the client interceptor (middleware) if a token is provided
	var interceptors []connect.ClientOption
	if *token != "" {
		interceptors = append(interceptors, connect.WithInterceptors(NewAuthInterceptor(*token)))
	}

	// 2. Instantiate the clients
	authClient := authv1connect.NewAuthServiceClient(
		http.DefaultClient,
		*serverURL,
		interceptors...,
	)

	userClient := authv1connect.NewUserServiceClient(
		http.DefaultClient,
		*serverURL,
		interceptors...,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Printf("Connecting to auth server at: %s\n", *serverURL)
	fmt.Printf("Executing command: %s\n\n", *cmd)

	switch *cmd {
	case "start-oauth":
		req := connect.NewRequest(&authv1.StartOAuthRequest{
			Provider: *provider,
		})
		res, err := authClient.StartOAuth(ctx, req)
		if err != nil {
			log.Fatalf("Error starting OAuth flow: %v", err)
		}
		fmt.Println("Successfully generated OAuth URL:")
		fmt.Printf("Auth URL:  %s\n", res.Msg.AuthUrl)
		fmt.Printf("CSRF State: %s\n", res.Msg.State)

	case "validate":
		if *token == "" {
			log.Fatal("Error: -token is required for 'validate' command")
		}
		req := connect.NewRequest(&authv1.ValidateTokenRequest{
			Token: *token,
		})
		res, err := authClient.ValidateToken(ctx, req)
		if err != nil {
			log.Fatalf("Error validating token: %v", err)
		}
		fmt.Printf("Token Valid: %t\n", res.Msg.Valid)
		if res.Msg.Valid && res.Msg.User != nil {
			fmt.Println("User details associated with this token:")
			printUser(res.Msg.User)
		}

	case "current-user":
		if *token == "" {
			log.Fatal("Error: -token is required for 'current-user' command")
		}
		// Notice that GetCurrentUserRequest doesn't take parameters; the interceptor
		// injects the JWT which identifies the user on the server.
		req := connect.NewRequest(&authv1.GetCurrentUserRequest{})
		res, err := authClient.GetCurrentUser(ctx, req)
		if err != nil {
			log.Fatalf("Error getting current user: %v", err)
		}
		fmt.Println("Current authenticated user profile:")
		printUser(res.Msg.User)

	case "list-users":
		if *token == "" {
			log.Fatal("Error: -token is required for 'list-users' command")
		}
		req := connect.NewRequest(&authv1.ListRequest{
			Limit:  int32(*limit),
			Offset: int32(*offset),
		})
		res, err := userClient.List(ctx, req)
		if err != nil {
			log.Fatalf("Error listing users: %v", err)
		}
		fmt.Printf("Retrieved %d users:\n", len(res.Msg.Users))
		for idx, u := range res.Msg.Users {
			fmt.Printf("[%d] ID: %s | Name: %s | Email: %s | Disabled: %t\n",
				idx+1, u.Id, u.Name, u.Email, u.Disabled)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q. Valid commands: start-oauth, validate, current-user, list-users\n", *cmd)
		os.Exit(1)
	}
}

func printUser(u *authv1.User) {
	fmt.Printf("  UUID:        %s\n", u.Id)
	fmt.Printf("  External ID: %d\n", u.ExternalId)
	fmt.Printf("  Name:        %s\n", u.Name)
	fmt.Printf("  Email:       %s\n", u.Email)
	fmt.Printf("  Provider:    %s\n", u.Provider)
	fmt.Printf("  Roles:       %v\n", u.Roles)
	fmt.Printf("  Disabled:    %t\n", u.Disabled)
}
