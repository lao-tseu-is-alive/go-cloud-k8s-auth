//go:build dev

package main // IMPORTANT: This must match your main package name

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
)

// IsDevBuild is overridden to true when the 'dev' build tag is used.
var IsDevBuild = true // Overridden value for dev

// DevRoutes sets up routes that are only available in development
func DevRoutes(e *echo.Echo, yourService *Service, jwtAuthUrl string) {
	log.Println("Dev mode enabled: Adding dev-only routes.")
	// This is the dev-only route
	e.GET(jwtAuthUrl, yourService.fakeDevF5)
}

// fakeDevF5 is a handler that returns a fake JWT token for development purposes.
// This function will be part of your `Service` struct, as defined in main.go
func (s *Service) fakeDevF5(ctx echo.Context) error {
	s.Logger.TraceHttpRequest("fakeDevF5", ctx.Request())
	login := "USER_TEST"

	userInfo := &goHttpEcho.UserInfo{
		UserId:     config.GetAdminIdFromEnvOrPanic(defaultAdminId),
		ExternalId: config.GetAdminExternalIdFromEnvOrPanic(9999999),
		Name:       "NewSimpleAdminAuthenticator_Admin",
		Email:      config.GetAdminEmailFromEnvOrPanic(defaultAdminEmail),
		Login:      config.GetAdminUserFromEnvOrPanic(defaultAdminUser),
		IsAdmin:    false,
		Groups:     []int{}, // this is the group id of the global_admin group
	}
	token, err := s.server.JwtCheck.GetTokenFromUserInfo(userInfo)
	if err != nil {
		myErrMsg := fmt.Sprintf("error in fakeDevF5 failed to get jwt token from user info: %v", err)
		s.Logger.Error(myErrMsg)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"jwtStatus": myErrMsg, "token": ""})
	}
	// Prepare the http only cookie for jwt token
	cookie := new(http.Cookie)
	cookie.Name = s.jwtCookieName
	cookie.Path = "/"
	cookie.Value = token.String()
	cookie.Expires = time.Now().Add(4 * time.Hour) // Set expiration
	cookie.HttpOnly = true                         // ⭐ Most important part: prevents JS access
	cookie.Secure = true                           // to allow working in dev with vite
	cookie.SameSite = http.SameSiteNoneMode        // CSRF protection
	ctx.SetCookie(cookie)
	// Prepare the response
	response := map[string]string{
		"jwtStatus": "success",
		"token":     token.String(),
	}
	s.Logger.Info(fmt.Sprintf("getJwtCookieFromF5(%s) successful, token set in HTTP-Only cookie.", login))
	return ctx.JSON(http.StatusOK, response)

}
