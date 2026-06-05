package auth

import "errors"

// Domain-specific errors for the Auth service
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrAlreadyExists = errors.New("user already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrNotOwner      = errors.New("user is not the owner")
	ErrAdminRequired = errors.New("admin privileges required")
	ErrProviderError = errors.New("oauth provider error")
	ErrInvalidState  = errors.New("invalid or expired oauth state")
	ErrUserDisabled  = errors.New("user account is disabled")
)
