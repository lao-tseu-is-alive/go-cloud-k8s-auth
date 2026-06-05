package template4gopackage

import "errors"

// Domain-specific errors for the Template4ServiceName service
var (
	ErrNotFound                         = errors.New("template_4_your_project_name not found")
	ErrAlreadyExists                    = errors.New("template_4_your_project_name already exists")
	ErrTypeTemplate4ServiceNameNotFound = errors.New("type template_4_your_project_name not found")
	ErrUnauthorized                     = errors.New("unauthorized")
	ErrInvalidInput                     = errors.New("invalid input")
	ErrNotOwner                         = errors.New("user is not the owner")
	ErrAdminRequired                    = errors.New("admin privileges required")
)
