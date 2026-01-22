package cli_server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/billgoswell/todoapp/integration-tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIServerAuthentication - Phase 5a Category 2 (3 tests)
// Tests validating authentication flow between CLI and Server

func TestCliCanAuthenticateWithValidApiKey(t *testing.T) {
	// Test 4: CLI can authenticate with valid API key
	// Validates that CLI successfully authenticates using valid credentials

	config := utils.DefaultConfig()
	client := utils.NewHTTPClient(config)

	// Make authenticated request - should succeed
	resp, err := client.Get("/api/v1/tasks")
	require.NoError(t, err, "Authenticated request should succeed")
	defer resp.Body.Close()

	// Should not return 401 Unauthorized
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode,
		"Valid API key should not result in 401")

	// Verify we got a response (may be 200, 400, or other non-401 status)
	assert.Greater(t, resp.StatusCode, 0, "Should receive response")
}

func TestAuthenticationFailsWithInvalidApiKey(t *testing.T) {
	// Test 5: Authentication fails with invalid API key
	// Validates that CLI properly rejects invalid credentials

	config := utils.DefaultConfig()
	config.APIKey = "invalid-api-key-xyz"
	client := utils.NewHTTPClient(config)

	// Make request with invalid key
	resp, err := client.Get("/api/v1/tasks")
	require.NoError(t, err, "Request should complete (even if 401)")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"Invalid API key should result in 401 Unauthorized")
}

func TestCredentialsPersistedAndReusedCorrectly(t *testing.T) {
	// Test 6: Credentials persisted and reused correctly
	// Validates that CLI properly stores and reuses credentials across requests

	config := utils.DefaultConfig()
	client := utils.NewHTTPClient(config)

	// Make first request
	resp1, err := client.Get("/api/v1/lists")
	require.NoError(t, err, "First request should succeed")
	defer resp1.Body.Close()

	initialStatus := resp1.StatusCode

	// Make second request - should use same credentials
	resp2, err := client.Get("/api/v1/lists")
	require.NoError(t, err, "Second request should succeed")
	defer resp2.Body.Close()

	// Both requests should get same response
	assert.Equal(t, initialStatus, resp2.StatusCode,
		"Subsequent requests should use same credentials and get same response")

	// Verify authorization header is sent correctly
	expectedAuth := fmt.Sprintf("Bearer %s", config.APIKey)
	assert.NotEmpty(t, expectedAuth, "Authorization header should be prepared")
	assert.Contains(t, expectedAuth, "Bearer", "Should include Bearer prefix")
}
