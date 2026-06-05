//go:build !dev

// constants.go
package main

import (
	"github.com/labstack/echo/v4"
)

// Default for non-dev builds (production)
var IsDevBuild = false

func DevRoutes(e *echo.Echo, service *Service, jwtAuthUrl string) {
	// No-op for production
}
