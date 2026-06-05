package auth

import "errors"

// Domain-specific errors for the Auth service
var (
	ErrNotFound                         = errors.New("go_cloud_auth not found")
	ErrAlreadyExists                    = errors.New("go_cloud_auth already exists")
	ErrTypeAuthNotFound = errors.New("type go_cloud_auth not found")
	ErrUnauthorized                     = errors.New("unauthorized")
	ErrInvalidInput                     = errors.New("invalid input")
	ErrNotOwner                         = errors.New("user is not the owner")
	ErrAdminRequired                    = errors.New("admin privileges required")
)
