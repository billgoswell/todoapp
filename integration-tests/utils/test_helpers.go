package utils

// CONSOLIDATION NOTE (Phase 1.3 - Refactoring):
// Test helpers are conceptually consolidated across two locations:
// 1. Server: internal/tests/helpers.go (database/fixtures for unit tests)
// 2. Integration: utils/test_helpers.go (API/CLI testing)
//
// These files are maintained in parallel with complementary functionality:
// - Server: Database setup, testcontainers, fixture builders
// - Integration: HTTP client, CLI runner, API/CLI integration testing
//
// Both use shared-types for Task/TodoList/User models to ensure consistency.
// See: /shared-types/README.md for details on shared type definitions.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestConfig holds integration test configuration
type TestConfig struct {
	ServerURL   string
	APIKey      string
	CliPath     string
	Timeout     time.Duration
	RetryCount  int
	RetryDelay  time.Duration
}

// DefaultConfig returns default test configuration
func DefaultConfig() *TestConfig {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key-12345"
	}

	cliPath := os.Getenv("CLI_PATH")
	if cliPath == "" {
		cliPath = "../todo-cmdline/todo"
	}

	return &TestConfig{
		ServerURL:  serverURL,
		APIKey:     apiKey,
		CliPath:    cliPath,
		Timeout:    30 * time.Second,
		RetryCount: 3,
		RetryDelay: 100 * time.Millisecond,
	}
}

// HTTPClient makes HTTP requests with retry logic
type HTTPClient struct {
	BaseURL string
	APIKey  string
	Config  *TestConfig
	client  *http.Client
}

// NewHTTPClient creates a new HTTP client for testing
func NewHTTPClient(config *TestConfig) *HTTPClient {
	return &HTTPClient{
		BaseURL: config.ServerURL,
		APIKey:  config.APIKey,
		Config:  config,
		client:  &http.Client{Timeout: config.Timeout},
	}
}

// Request makes an HTTP request with retry logic
func (h *HTTPClient) Request(method, path string, body interface{}) (*http.Response, error) {
	url := h.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	var lastErr error
	for attempt := 0; attempt < h.Config.RetryCount; attempt++ {
		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.APIKey))
		req.Header.Set("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < h.Config.RetryCount-1 {
				time.Sleep(h.Config.RetryDelay)
			}
			continue
		}

		// Success
		if resp.StatusCode < 500 {
			return resp, nil
		}

		// Server error - retry
		lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
		resp.Body.Close()
		if attempt < h.Config.RetryCount-1 {
			time.Sleep(h.Config.RetryDelay)
		}
	}

	return nil, lastErr
}

// Get makes a GET request
func (h *HTTPClient) Get(path string) (*http.Response, error) {
	return h.Request("GET", path, nil)
}

// Post makes a POST request
func (h *HTTPClient) Post(path string, body interface{}) (*http.Response, error) {
	return h.Request("POST", path, body)
}

// Put makes a PUT request
func (h *HTTPClient) Put(path string, body interface{}) (*http.Response, error) {
	return h.Request("PUT", path, body)
}

// Delete makes a DELETE request
func (h *HTTPClient) Delete(path string) (*http.Response, error) {
	return h.Request("DELETE", path, nil)
}

// CLIRunner executes CLI commands
type CLIRunner struct {
	CliPath string
	Config  *TestConfig
}

// NewCLIRunner creates a new CLI runner
func NewCLIRunner(config *TestConfig) *CLIRunner {
	return &CLIRunner{
		CliPath: config.CliPath,
		Config:  config,
	}
}

// Run executes a CLI command and returns output
func (c *CLIRunner) Run(args ...string) (string, error) {
	cmd := exec.Command(c.CliPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SERVER_URL=%s", c.Config.ServerURL),
		fmt.Sprintf("API_KEY=%s", c.Config.APIKey),
	)

	err := cmd.Run()
	output := stdout.String()
	if err != nil {
		return output, fmt.Errorf("CLI error: %w (stderr: %s)", err, stderr.String())
	}
	return output, nil
}

// HealthCheck verifies server is responding
func (h *HTTPClient) HealthCheck() error {
	resp, err := h.Get("/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	return nil
}

// WaitForServer waits for server to be available
func (h *HTTPClient) WaitForServer(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := h.HealthCheck(); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server not available after %v", timeout)
}

// AssertStatusCode verifies HTTP response status code
func AssertStatusCode(t *testing.T, resp *http.Response, expected int) {
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d (body: %s)", expected, resp.StatusCode, string(body))
	}
}

// ParseJSONResponse parses JSON response body
func ParseJSONResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// TestTask represents a task for testing
type TestTask struct {
	ID           int64  `json:"id,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	Todo         string `json:"todo"`
	Priority     int    `json:"priority"`
	Done         bool   `json:"done"`
	TodoListID   int64  `json:"todo_list_id"`
	UpdatedAt    int64  `json:"updated_at,omitempty"`
	Deleted      bool   `json:"deleted"`
	DateAdded    int64  `json:"date_added,omitempty"`
	DateCompleted int64 `json:"date_completed,omitempty"`
	DueDate      int64  `json:"due_date,omitempty"`
}

// TestList represents a list for testing
type TestList struct {
	ID           int64  `json:"id,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
	Archived     bool   `json:"archived"`
	CreatedAt    int64  `json:"created_at,omitempty"`
	UpdatedAt    int64  `json:"updated_at,omitempty"`
}

// NewTestTask creates a test task
func NewTestTask(listID int64, todo string) *TestTask {
	return &TestTask{
		ClientID:   fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Todo:       todo,
		Priority:   3,
		Done:       false,
		TodoListID: listID,
		Deleted:    false,
	}
}

// NewTestList creates a test list
func NewTestList(name string) *TestList {
	return &TestList{
		ClientID:     fmt.Sprintf("list-%d", time.Now().UnixNano()),
		Name:         name,
		DisplayOrder: 0,
		Archived:     false,
	}
}

// IsServerRunning checks if server is running
func IsServerRunning(serverURL string) bool {
	client := NewHTTPClient(&TestConfig{ServerURL: serverURL})
	return client.HealthCheck() == nil
}

// IsCliAvailable checks if CLI binary exists
func IsCliAvailable(cliPath string) bool {
	_, err := os.Stat(cliPath)
	return err == nil
}

// TrimOutput removes whitespace and newlines
func TrimOutput(output string) string {
	return strings.TrimSpace(output)
}
