package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	cfg "github.com/billgoswell/commandlinetodo/internal/config"
	"github.com/billgoswell/commandlinetodo/internal/logger"
)

// ConnectivityStatus represents the current connection status
type ConnectivityStatus int

const (
	StatusOffline ConnectivityStatus = iota
	StatusOnline
	StatusSyncing
	StatusError
)

// SyncClient handles HTTP communication with the sync server
type SyncClient struct {
	baseURL      string
	httpClient   *http.Client
	apiKey       string
	deviceID     string
	status       ConnectivityStatus
	lastCheck    time.Time
	lastError    error
	mu           sync.RWMutex
	checkTimeout time.Duration
}

// TaskPayload represents a task for sync
type TaskPayload struct {
	ClientID         string `json:"client_id"`
	Todo             string `json:"todo"`
	Priority         int    `json:"priority"`
	Done             bool   `json:"done"`
	DateAdded        int64  `json:"date_added"`
	DateCompleted    int64  `json:"date_completed"`
	DueDate          int64  `json:"due_date"`
	Deleted          bool   `json:"deleted"`
	DeletedAt        int64  `json:"deleted_at"`
	TodoListID       int    `json:"todo_list_id"`
	TodoListClientID string `json:"todo_list_client_id"`
	UpdatedAt        int64  `json:"updated_at"`
	Version          int    `json:"version"`
}

// ListPayload represents a todo list for sync
type ListPayload struct {
	ClientID    string `json:"client_id"`
	Name        string `json:"name"`
	DisplayOrder int    `json:"display_order"`
	Archived    bool   `json:"archived"`
	UpdatedAt   int64  `json:"updated_at"`
	Version     int    `json:"version"`
}

// PullRequest is the request for pulling changes
type PullRequest struct {
	Since int64 `json:"since"`
}

// PullResponse contains the changes from the server
type PullResponse struct {
	Tasks []TaskPayload `json:"tasks"`
	Lists []ListPayload `json:"lists"`
}

// PushRequest contains changes to push to server
type PushRequest struct {
	Tasks []TaskPayload `json:"tasks"`
	Lists []ListPayload `json:"lists"`
}

// IDMapping represents a client-to-server ID mapping returned from push
type IDMapping struct {
	ClientID   string `json:"client_id"`
	ServerID   int    `json:"server_id"`
	EntityType string `json:"entity_type"`
}

// PushResponse contains the result of a push operation
type PushResponse struct {
	Status           string      `json:"status"`
	ListIDMappings   []IDMapping `json:"list_id_mappings"`
	TaskIDMappings   []IDMapping `json:"task_id_mappings"`
}

// HealthResponse is the response from the health endpoint
type HealthResponse struct {
	Status string `json:"status"`
}

// NewSyncClient creates a new sync client
func NewSyncClient(config cfg.SyncConfig) *SyncClient {
	return &SyncClient{
		baseURL:      config.ServerURL,
		apiKey:       config.APIKey,
		deviceID:     config.DeviceID,
		status:       StatusOffline,
		lastCheck:    time.Time{},
		checkTimeout: 5 * time.Second,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		},
	}
}

// CheckConnectivity checks if the server is reachable
func (c *SyncClient) CheckConnectivity() ConnectivityStatus {
	c.mu.RLock()
	// Return cached status if checked recently (within 5 seconds)
	if time.Since(c.lastCheck) < c.checkTimeout {
		defer c.mu.RUnlock()
		return c.status
	}
	c.mu.RUnlock()

	logger.LogDebug("Checking connectivity", "server", c.baseURL)

	// Try to reach the health endpoint
	req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
	if err != nil {
		logger.LogError("Failed to create health check request", "server", c.baseURL, "error", err)
		c.updateStatus(StatusError, err)
		return StatusError
	}

	c.addAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.LogWarn("Connectivity check failed", "server", c.baseURL, "error", err)
		c.updateStatus(StatusOffline, err)
		return StatusOffline
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logger.LogDebug("Connectivity check successful", "server", c.baseURL)
		c.updateStatus(StatusOnline, nil)
		return StatusOnline
	}

	logger.LogWarn("Connectivity check returned non-OK status",
		"server", c.baseURL,
		"status", resp.StatusCode)
	c.updateStatus(StatusError, fmt.Errorf("server returned status %d", resp.StatusCode))
	return StatusError
}

// IsOnline returns true if the sync client is online
func (c *SyncClient) IsOnline() bool {
	return c.CheckConnectivity() == StatusOnline
}

// PullChanges retrieves changes from the server since the given timestamp
func (c *SyncClient) PullChanges(since int64) (*PullResponse, error) {
	logger.LogDebug("Starting pull", "server", c.baseURL, "since", since)

	if !c.IsOnline() {
		logger.LogWarn("Pull failed: server offline", "server", c.baseURL)
		return nil, fmt.Errorf("not connected to sync server")
	}

	pullReq := PullRequest{Since: since}
	body, err := json.Marshal(pullReq)
	if err != nil {
		logger.LogError("Failed to marshal pull request", "error", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/sync/pull", bytes.NewReader(body))
	if err != nil {
		logger.LogError("Failed to create pull request", "error", err)
		return nil, err
	}

	c.addAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.LogError("Pull request failed", "server", c.baseURL, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.LogError("Pull failed with non-OK status",
			"server", c.baseURL,
			"status", resp.StatusCode,
			"response", string(bodyBytes))
		return nil, fmt.Errorf("pull changes failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var pullResp PullResponse
	if err := json.NewDecoder(resp.Body).Decode(&pullResp); err != nil {
		logger.LogError("Failed to decode pull response", "error", err)
		return nil, err
	}

	logger.LogDebug("Pull completed successfully", "tasks", len(pullResp.Tasks), "lists", len(pullResp.Lists))
	return &pullResp, nil
}

// PushChanges sends local changes to the server and returns ID mappings
func (c *SyncClient) PushChanges(items []todoItem, lists []todoList) (*PushResponse, error) {
	logger.LogDebug("Starting push", "server", c.baseURL, "items", len(items), "lists", len(lists))

	if !c.IsOnline() {
		logger.LogWarn("Push failed: server offline", "server", c.baseURL)
		return nil, fmt.Errorf("not connected to sync server")
	}

	// Convert items to payloads
	taskPayloads := make([]TaskPayload, len(items))
	for i, item := range items {
		taskPayloads[i] = TaskPayload{
			ClientID:         item.clientID,
			Todo:             item.todo,
			Priority:         item.priority,
			Done:             item.done,
			DateAdded:        item.dateAdded,
			DateCompleted:    item.dateCompleted,
			DueDate:          item.dueDate,
			Deleted:          item.deleted,
			DeletedAt:        item.deletedAt,
			TodoListID:       item.todoListID,
			TodoListClientID: item.listClientID,
			UpdatedAt:        item.updatedAt,
			Version:          item.version,
		}
	}

	// Convert lists to payloads
	listPayloads := make([]ListPayload, len(lists))
	for i, list := range lists {
		listPayloads[i] = ListPayload{
			ClientID:     list.clientID,
			Name:         list.name,
			DisplayOrder: list.displayOrder,
			Archived:     list.archived,
			UpdatedAt:    list.updatedAt,
			Version:      list.version,
		}
	}

	pushReq := PushRequest{
		Tasks: taskPayloads,
		Lists: listPayloads,
	}

	body, err := json.Marshal(pushReq)
	if err != nil {
		logger.LogError("Failed to marshal push request", "error", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/sync/push", bytes.NewReader(body))
	if err != nil {
		logger.LogError("Failed to create push request", "error", err)
		return nil, err
	}

	c.addAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.LogError("Push request failed", "server", c.baseURL, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.LogError("Push failed with non-OK status",
			"server", c.baseURL,
			"status", resp.StatusCode,
			"response", string(bodyBytes))
		return nil, fmt.Errorf("push changes failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the push response
	var pushResp PushResponse
	if err := json.NewDecoder(resp.Body).Decode(&pushResp); err != nil {
		logger.LogError("Failed to decode push response", "error", err)
		return nil, err
	}

	logger.LogDebug("Push completed successfully",
		"items", len(items),
		"lists", len(lists),
		"list_mappings", len(pushResp.ListIDMappings),
		"task_mappings", len(pushResp.TaskIDMappings))
	return &pushResp, nil
}

// addAuthHeaders adds authentication headers to requests
func (c *SyncClient) addAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Device-ID", c.deviceID)
}

// updateStatus updates the connection status
func (c *SyncClient) updateStatus(status ConnectivityStatus, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status = status
	c.lastError = err
	c.lastCheck = time.Now()
}

// GetLastError returns the last error that occurred
func (c *SyncClient) GetLastError() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastError
}
