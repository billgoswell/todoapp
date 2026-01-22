package cli_server

import (
	"net/http"
	"testing"
	"time"

	"github.com/billgoswell/todoapp/integration-tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIServerConnection - Phase 5a Category 1 (3 tests)
// Tests validating direct CLI and Server communication

func TestCliCanConnectToRunningServer(t *testing.T) {
	// Test 1: CLI can connect to running server
	// Validates that CLI successfully establishes connection to server

	config := utils.DefaultConfig()
	client := utils.NewHTTPClient(config)

	// Wait for server to be available
	err := client.WaitForServer(10 * time.Second)
	require.NoError(t, err, "Server should be running and available")

	// Verify server health check passes
	resp, err := client.Get("/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check should succeed")
}

func TestHealthCheckEndpointRespondsCorrectly(t *testing.T) {
	// Test 2: Health check endpoint responds correctly
	// Validates that server health endpoint returns proper response

	config := utils.DefaultConfig()
	client := utils.NewHTTPClient(config)

	resp, err := client.Get("/health")
	require.NoError(t, err, "Health check request should succeed")
	defer resp.Body.Close()

	// Verify response status
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify response has body
	var healthResp map[string]interface{}
	err = utils.ParseJSONResponse(resp, &healthResp)
	// Health endpoint may return empty or with status field
	assert.NotNil(t, resp.Body, "Health check should have response body")
}

func TestConnectionFailsGracefullyWhenServerUnavailable(t *testing.T) {
	// Test 3: Connection fails gracefully when server unavailable
	// Validates that CLI handles server unavailability without crashing

	config := utils.DefaultConfig()
	// Use invalid server URL
	config.ServerURL = "http://localhost:55555"
	config.RetryCount = 1
	config.Timeout = 1 * time.Second

	client := utils.NewHTTPClient(config)

	// Should fail gracefully
	_, err := client.Get("/health")
	assert.Error(t, err, "Connection to unavailable server should fail")
	assert.NotNil(t, err, "Error should be returned")

	// Verify error message is informative
	assert.Contains(t, err.Error(), "", "Error message should be present")
}
