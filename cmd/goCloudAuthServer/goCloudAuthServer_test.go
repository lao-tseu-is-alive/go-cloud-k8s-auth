package main

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-auth/pkg/version"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
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
}
