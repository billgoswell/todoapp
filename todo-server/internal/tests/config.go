package tests

import (
	"os"
	"strconv"
)

// TestConfig holds configuration for test execution
type TestConfig struct {
	// UseLocalServer determines if tests should run against testcontainers-go PostgreSQL
	UseLocalServer bool

	// LiveServerHost is the host of the live server for integration tests
	LiveServerHost string

	// LiveServerPort is the port of the live server
	LiveServerPort string

	// TestAPIKey is the API key for testing against live server
	TestAPIKey string

	// Timeout in seconds for HTTP requests
	Timeout int
}

// GetTestConfig returns the current test configuration from environment variables
func GetTestConfig() TestConfig {
	config := TestConfig{
		UseLocalServer: true,  // Default: use local testcontainers
		LiveServerHost: "192.168.100.20",
		LiveServerPort: "8080",
		Timeout:        30,
	}

	// Check environment variables for overrides
	if useLocal := os.Getenv("TEST_USE_LOCAL_SERVER"); useLocal != "" {
		config.UseLocalServer, _ = strconv.ParseBool(useLocal)
	}

	if host := os.Getenv("TEST_LIVE_SERVER_HOST"); host != "" {
		config.LiveServerHost = host
	}

	if port := os.Getenv("TEST_LIVE_SERVER_PORT"); port != "" {
		config.LiveServerPort = port
	}

	if apiKey := os.Getenv("TEST_API_KEY"); apiKey != "" {
		config.TestAPIKey = apiKey
	}

	if timeout := os.Getenv("TEST_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			config.Timeout = t
		}
	}

	return config
}

// GetLiveServerURL returns the base URL for the live server
func (tc TestConfig) GetLiveServerURL() string {
	return "http://" + tc.LiveServerHost + ":" + tc.LiveServerPort
}
