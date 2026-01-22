package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/billgoswell/commandlinetodo-server/internal/models"
)

// IntegrationTestClient provides methods for testing against the live server
type IntegrationTestClient struct {
	BaseURL   string
	APIKey    string
	HTTPClient *http.Client
}

// NewIntegrationTestClient creates a new client for testing the live server
func NewIntegrationTestClient(config TestConfig) *IntegrationTestClient {
	return &IntegrationTestClient{
		BaseURL: config.GetLiveServerURL(),
		APIKey:  config.TestAPIKey,
		HTTPClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// doRequest performs an HTTP request with proper headers
func (c *IntegrationTestClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	return c.HTTPClient.Do(req)
}

// GetHealth checks if the server is healthy
func (c *IntegrationTestClient) GetHealth() (bool, error) {
	resp, err := c.doRequest("GET", "/api/v1/health", nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, err
	}

	status, ok := response["status"].(string)
	return ok && status == "ok", nil
}

// GetTasks retrieves all tasks for the authenticated user
func (c *IntegrationTestClient) GetTasks() ([]models.Task, error) {
	resp, err := c.doRequest("GET", "/api/v1/tasks", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get tasks failed with status %d", resp.StatusCode)
	}

	var tasks []models.Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// CreateTask creates a new task
func (c *IntegrationTestClient) CreateTask(task *models.Task) (*models.Task, error) {
	resp, err := c.doRequest("POST", "/api/v1/tasks", task)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create task failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdTask models.Task
	if err := json.NewDecoder(resp.Body).Decode(&createdTask); err != nil {
		return nil, err
	}

	return &createdTask, nil
}

// GetLists retrieves all lists for the authenticated user
func (c *IntegrationTestClient) GetLists() ([]models.TodoList, error) {
	resp, err := c.doRequest("GET", "/api/v1/lists", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get lists failed with status %d", resp.StatusCode)
	}

	var lists []models.TodoList
	if err := json.NewDecoder(resp.Body).Decode(&lists); err != nil {
		return nil, err
	}

	return lists, nil
}

// CreateList creates a new list
func (c *IntegrationTestClient) CreateList(list *models.TodoList) (*models.TodoList, error) {
	resp, err := c.doRequest("POST", "/api/v1/lists", list)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create list failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdList models.TodoList
	if err := json.NewDecoder(resp.Body).Decode(&createdList); err != nil {
		return nil, err
	}

	return &createdList, nil
}

// Pull fetches changes from the server since a given timestamp
func (c *IntegrationTestClient) Pull(since int64) (*models.SyncResponse, error) {
	req := models.SyncRequest{Since: since}
	resp, err := c.doRequest("POST", "/api/v1/sync/pull", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pull failed with status %d", resp.StatusCode)
	}

	var syncResp models.SyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return nil, err
	}

	return &syncResp, nil
}

// Push sends changes to the server
func (c *IntegrationTestClient) Push(req *models.PushRequest) error {
	resp, err := c.doRequest("POST", "/api/v1/sync/push", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// ============================================================================
// INTEGRATION TESTS AGAINST LIVE SERVER
// ============================================================================

// TestLiveServer_Health checks if the live server is healthy
func TestLiveServer_Health(t *testing.T) {
	config := GetTestConfig()
	if config.UseLocalServer {
		t.Skip("Skipping live server test - using local database setup")
	}

	client := NewIntegrationTestClient(config)
	healthy, err := client.GetHealth()

	if err != nil {
		t.Fatalf("Failed to check health: %v", err)
	}

	if !healthy {
		t.Error("Server health check returned false")
	}
}

// TestLiveServer_GetTasks retrieves tasks from live server
func TestLiveServer_GetTasks(t *testing.T) {
	config := GetTestConfig()
	if config.UseLocalServer {
		t.Skip("Skipping live server test - using local database setup")
	}

	if config.TestAPIKey == "" {
		t.Skip("TEST_API_KEY not set - cannot test against live server")
	}

	client := NewIntegrationTestClient(config)
	tasks, err := client.GetTasks()

	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	if tasks == nil {
		t.Error("Expected tasks array, got nil")
	}
}

// TestLiveServer_GetLists retrieves lists from live server
func TestLiveServer_GetLists(t *testing.T) {
	config := GetTestConfig()
	if config.UseLocalServer {
		t.Skip("Skipping live server test - using local database setup")
	}

	if config.TestAPIKey == "" {
		t.Skip("TEST_API_KEY not set - cannot test against live server")
	}

	client := NewIntegrationTestClient(config)
	lists, err := client.GetLists()

	if err != nil {
		t.Fatalf("Failed to get lists: %v", err)
	}

	if lists == nil {
		t.Error("Expected lists array, got nil")
	}
}

// TestLiveServer_CreateTask creates a task on live server
func TestLiveServer_CreateTask(t *testing.T) {
	config := GetTestConfig()
	if config.UseLocalServer {
		t.Skip("Skipping live server test - using local database setup")
	}

	if config.TestAPIKey == "" {
		t.Skip("TEST_API_KEY not set - cannot test against live server")
	}

	client := NewIntegrationTestClient(config)

	// First get or create a list
	lists, err := client.GetLists()
	if err != nil {
		t.Fatalf("Failed to get lists: %v", err)
	}

	if len(lists) == 0 {
		t.Skip("No lists available on live server - create one first")
	}

	// Create task in first list
	task := &models.Task{
		ClientID:   fmt.Sprintf("test-task-%d", time.Now().Unix()),
		TodoListID: lists[0].ID,
		Todo:       "Test task from integration test",
		Priority:   3,
		Done:       false,
		DateAdded:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
		Version:    1,
	}

	createdTask, err := client.CreateTask(task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if createdTask.ID == 0 {
		t.Error("Expected task ID to be set after creation")
	}

	if createdTask.Todo != task.Todo {
		t.Errorf("Expected todo '%s', got '%s'", task.Todo, createdTask.Todo)
	}
}

// TestLiveServer_SyncPull tests pull operation against live server
func TestLiveServer_SyncPull(t *testing.T) {
	config := GetTestConfig()
	if config.UseLocalServer {
		t.Skip("Skipping live server test - using local database setup")
	}

	if config.TestAPIKey == "" {
		t.Skip("TEST_API_KEY not set - cannot test against live server")
	}

	client := NewIntegrationTestClient(config)

	// Pull all changes since epoch
	syncResp, err := client.Pull(0)
	if err != nil {
		t.Fatalf("Failed to pull: %v", err)
	}

	if syncResp == nil {
		t.Error("Expected sync response, got nil")
	}

	if syncResp.Tasks == nil {
		t.Error("Expected tasks array in sync response")
	}

	if syncResp.Lists == nil {
		t.Error("Expected lists array in sync response")
	}
}

// TestLiveServer_AuthRequired tests that endpoints require authentication
func TestLiveServer_AuthRequired(t *testing.T) {
	config := GetTestConfig()
	if config.UseLocalServer {
		t.Skip("Skipping live server test - using local database setup")
	}

	client := NewIntegrationTestClient(config)
	client.APIKey = "" // Remove auth

	resp, err := client.doRequest("GET", "/api/v1/tasks", nil)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}
