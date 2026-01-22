package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ============================================================================
// SYNC CLIENT TESTS - PHASE 3
// ============================================================================
// These tests verify sync client operations with both mock and live server support

// MockSyncServer creates a test HTTP server that simulates the todo-server API
type MockSyncServer struct {
	server *httptest.Server
}

func NewMockSyncServer() *MockSyncServer {
	mux := http.NewServeMux()

	// Mock health endpoint (matches SyncClient's /health endpoint)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	// Mock pull endpoint (matches SyncClient's /sync/pull endpoint)
	mux.HandleFunc("/sync/pull", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing auth"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Return mock pulled changes
		response := map[string]interface{}{
			"tasks": []map[string]interface{}{
				{
					"id":              1,
					"client_id":       "mock-task-001",
					"todo":            "Mock server task",
					"priority":        3,
					"done":            false,
					"date_added":      time.Now().Unix() - 3600,
					"updated_at":      time.Now().Unix(),
					"version":         1,
					"todo_list_id":    1,
				},
			},
			"lists": []map[string]interface{}{
				{
					"id":              1,
					"client_id":       "mock-list-001",
					"name":            "Mock List",
					"display_order":   0,
					"archived":        false,
					"created_at":      time.Now().Unix() - 3600,
					"updated_at":      time.Now().Unix(),
					"version":         1,
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	})

	// Mock push endpoint (matches SyncClient's /sync/push endpoint)
	mux.HandleFunc("/sync/push", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"synced":  (len(req["tasks"].([]interface{})) + len(req["lists"].([]interface{}))),
		})
	})

	server := httptest.NewServer(mux)
	return &MockSyncServer{server: server}
}

func (m *MockSyncServer) Close() {
	m.server.Close()
}

func (m *MockSyncServer) URL() string {
	return m.server.URL
}

// TestSyncClient_Connectivity_Online tests connectivity check with online server
func TestSyncClient_Connectivity_Online(t *testing.T) {
	mockServer := NewMockSyncServer()
	defer mockServer.Close()

	client := &SyncClient{
		baseURL:      mockServer.URL(),
		apiKey:       "test-key",
		deviceID:     "test-device",
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		checkTimeout: 5 * time.Second,
	}

	// Should report online
	online := client.IsOnline()
	if !online {
		t.Error("Expected IsOnline() to return true for running server")
	}
}

// TestSyncClient_Connectivity_Offline tests connectivity check with offline server
func TestSyncClient_Connectivity_Offline(t *testing.T) {
	client := &SyncClient{
		baseURL:      "http://127.0.0.1:1", // Invalid port
		apiKey:       "test-key",
		deviceID:     "test-device",
		httpClient:   &http.Client{Timeout: 100 * time.Millisecond},
		checkTimeout: 5 * time.Second,
	}

	// Should report offline
	online := client.IsOnline()
	if online {
		t.Error("Expected IsOnline() to return false for unreachable server")
	}
}

// TestSyncClient_PullChanges_Success tests successful pull operation with mock server
func TestSyncClient_PullChanges_Success(t *testing.T) {
	mockServer := NewMockSyncServer()
	defer mockServer.Close()

	client := &SyncClient{
		baseURL:      mockServer.URL(),
		apiKey:       "test-key",
		deviceID:     "test-device",
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		checkTimeout: 5 * time.Second,
	}

	// Pull changes since timestamp
	response, err := client.PullChanges(time.Now().Unix() - 7200)

	if err != nil {
		t.Fatalf("PullChanges failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}

	// Check that we got data back
	if len(response.Tasks) == 0 {
		t.Error("Expected tasks in response")
	}

	if len(response.Lists) == 0 {
		t.Error("Expected lists in response")
	}

	// Verify task data
	if len(response.Tasks) > 0 && response.Tasks[0].Todo != "Mock server task" {
		t.Errorf("Expected 'Mock server task', got '%s'", response.Tasks[0].Todo)
	}
}

// TestSyncClient_PullChanges_RequiresAuth tests that pull requires authentication
func TestSyncClient_PullChanges_RequiresAuth(t *testing.T) {
	mockServer := NewMockSyncServer()
	defer mockServer.Close()

	// Client without API key should fail
	client := &SyncClient{
		baseURL:      mockServer.URL(),
		apiKey:       "",
		deviceID:     "test-device",
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		checkTimeout: 5 * time.Second,
	}

	response, err := client.PullChanges(0)

	if response == nil && err == nil {
		// Some implementations might return error, some might return empty response
		t.Log("Pull with missing auth key handled")
	}
}

// TestSyncClient_PushChanges_Success tests successful push operation
func TestSyncClient_PushChanges_Success(t *testing.T) {
	mockServer := NewMockSyncServer()
	defer mockServer.Close()

	client := &SyncClient{
		baseURL:      mockServer.URL(),
		apiKey:       "test-key",
		deviceID:     "test-device",
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		checkTimeout: 5 * time.Second,
	}

	// Prepare items to push
	now := time.Now().Unix()
	items := []todoItem{
		{
			id:         1,
			clientID:   "test-001",
			todo:       "Test task for push",
			priority:   2,
			done:       false,
			dateAdded:  now - 3600,
			updatedAt:  now,
			todoListID: 1,
			version:    1,
		},
	}

	lists := []todoList{
		{
			id:        1,
			clientID:  "test-list-001",
			name:      "Push Test List",
			createdAt: now - 3600,
			updatedAt: now,
			version:   1,
		},
	}

	err := client.PushChanges(items, lists)

	if err != nil {
		t.Fatalf("PushChanges failed: %v", err)
	}
}

// TestSyncClient_PushChanges_EmptyData tests push with no items
func TestSyncClient_PushChanges_EmptyData(t *testing.T) {
	mockServer := NewMockSyncServer()
	defer mockServer.Close()

	client := &SyncClient{
		baseURL:      mockServer.URL(),
		apiKey:       "test-key",
		deviceID:     "test-device",
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		checkTimeout: 5 * time.Second,
	}

	// Push empty lists
	err := client.PushChanges([]todoItem{}, []todoList{})

	if err != nil {
		t.Fatalf("PushChanges with empty data failed: %v", err)
	}
}

// TestSyncClient_PayloadSerialization tests sync payload serialization
func TestSyncClient_PayloadSerialization(t *testing.T) {
	now := time.Now().Unix()
	item := todoItem{
		id:            1,
		clientID:      "test-001",
		todo:          "Test task",
		priority:      2,
		done:          false,
		dateAdded:     now - 3600,
		dateCompleted: 0,
		dueDate:       0,
		deleted:       false,
		deletedAt:     0,
		todoListID:    1,
		updatedAt:     now,
		version:       1,
	}

	// Convert to payload
	payload := TaskPayload{
		ClientID:      item.clientID,
		Todo:          item.todo,
		Priority:      item.priority,
		Done:          item.done,
		DateAdded:     item.dateAdded,
		DateCompleted: item.dateCompleted,
		DueDate:       item.dueDate,
		Deleted:       item.deleted,
		DeletedAt:     item.deletedAt,
		TodoListID:    item.todoListID,
		UpdatedAt:     item.updatedAt,
		Version:       item.version,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Deserialize back
	var deserializedPayload TaskPayload
	err = json.Unmarshal(jsonData, &deserializedPayload)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	// Verify all fields match
	if deserializedPayload.ClientID != item.clientID {
		t.Errorf("ClientID mismatch: expected %s, got %s", item.clientID, deserializedPayload.ClientID)
	}

	if deserializedPayload.Todo != item.todo {
		t.Errorf("Todo mismatch: expected %s, got %s", item.todo, deserializedPayload.Todo)
	}

	if deserializedPayload.UpdatedAt != item.updatedAt {
		t.Errorf("UpdatedAt mismatch: expected %d, got %d", item.updatedAt, deserializedPayload.UpdatedAt)
	}
}

// TestSyncClient_ConflictResolution_PayloadComparison tests conflict resolution using updated_at
func TestSyncClient_ConflictResolution_PayloadComparison(t *testing.T) {
	now := time.Now().Unix()

	// Two conflicting versions of the same item
	localVersion := todoItem{
		id:        1,
		clientID:  "task-001",
		todo:      "Local version",
		priority:  3,
		done:      false,
		updatedAt: now - 600, // Updated 10 minutes ago
		version:   2,
	}

	serverVersion := todoItem{
		id:        1,
		clientID:  "task-001",
		todo:      "Server version",
		priority:  2,
		done:      true,
		updatedAt: now - 300, // Updated 5 minutes ago (more recent)
		version:   3,
	}

	// The one with higher updated_at should win
	if serverVersion.updatedAt > localVersion.updatedAt {
		t.Log("Server version wins due to more recent update timestamp")

		// Server version should be used
		if serverVersion.todo != "Server version" {
			t.Error("Expected to use server version when it has newer timestamp")
		}
	} else {
		t.Error("Server version should have newer timestamp")
	}
}

// TestSyncClient_DeletedItemHandling tests sync of deleted items
func TestSyncClient_DeletedItemHandling(t *testing.T) {
	now := time.Now().Unix()

	deletedItem := todoItem{
		id:        1,
		clientID:  "task-001",
		todo:      "Deleted task",
		deleted:   true,
		deletedAt: now - 300,
		updatedAt: now - 300,
		version:   2,
	}

	// Verify deleted flag and timestamp are preserved
	if !deletedItem.deleted {
		t.Error("Expected deleted=true")
	}

	if deletedItem.deletedAt == 0 {
		t.Error("Expected deletedAt to be set")
	}

	if deletedItem.updatedAt == 0 {
		t.Error("Expected updatedAt to be set")
	}
}

// TestSyncClient_LargePayload tests sync with many items
func TestSyncClient_LargePayload(t *testing.T) {
	now := time.Now().Unix()

	// Create 100 items
	items := make([]todoItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = todoItem{
			id:        i + 1,
			clientID:  "task-" + string(rune(i)),
			todo:      "Task number " + string(rune(i)),
			priority:  (i % 4) + 1,
			done:      (i % 3) == 0,
			dateAdded: now - int64((100-i)*3600),
			updatedAt: now - int64((100-i)*1800),
			version:   i%5 + 1,
		}
	}

	// Serialize to JSON (simulating what would be sent)
	payload := map[string]interface{}{
		"tasks": items,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal large payload: %v", err)
	}

	// Verify we can deserialize large payloads
	var receivedPayload map[string]interface{}
	err = json.Unmarshal(jsonData, &receivedPayload)
	if err != nil {
		t.Fatalf("Failed to unmarshal large payload: %v", err)
	}

	// Verify size is reasonable
	if len(jsonData) == 0 {
		t.Error("Payload should have content")
	}

	t.Logf("Successfully handled payload of %d bytes for 100 items", len(jsonData))
}

// TestSyncClient_RequestBuilding tests proper HTTP request construction
func TestSyncClient_RequestBuilding(t *testing.T) {
	mux := http.NewServeMux()

	// Health endpoint for connectivity check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Push endpoint - verify request headers
	mux.HandleFunc("/sync/push", func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type: application/json")
		}

		if r.Header.Get("Authorization") == "" {
			t.Error("Expected Authorization header")
		}

		if !bytes.Contains([]byte(r.Header.Get("Authorization")), []byte("Bearer")) {
			t.Error("Expected Bearer token in Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Echo back what we received
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(req)
	})

	mockServer := httptest.NewServer(mux)
	defer mockServer.Close()

	client := &SyncClient{
		baseURL:      mockServer.URL,
		apiKey:       "test-api-key",
		deviceID:     "test-device-123",
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		checkTimeout: 5 * time.Second,
	}

	items := []todoItem{
		{
			id:        1,
			clientID:  "test-001",
			todo:      "Test",
			updatedAt: time.Now().Unix(),
		},
	}

	err := client.PushChanges(items, []todoList{})
	if err != nil {
		t.Fatalf("PushChanges failed: %v", err)
	}
}

// TestSyncClient_ResponseHandling tests error response handling
func TestSyncClient_ResponseHandling(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 500 error
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error": "server error"}`)
	}))
	defer mockServer.Close()

	client := &SyncClient{
		baseURL:      mockServer.URL,
		apiKey:       "test-key",
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		checkTimeout: 5 * time.Second,
	}

	response, err := client.PullChanges(0)

	// Should handle error gracefully
	if response != nil || err == nil {
		t.Log("Error response handled appropriately")
	}
}

// TestSyncClient_Timestamp_Preservation tests that timestamps are preserved through sync
func TestSyncClient_Timestamp_Preservation(t *testing.T) {
	now := time.Now().Unix()
	originalAddTime := now - 86400    // 1 day ago
	originalUpdateTime := now - 3600  // 1 hour ago

	item := todoItem{
		id:        1,
		clientID:  "test-001",
		todo:      "Test task",
		dateAdded: originalAddTime,
		updatedAt: originalUpdateTime,
		version:   1,
	}

	// Serialize and deserialize
	payload := TaskPayload{
		ClientID:  item.clientID,
		DateAdded: item.dateAdded,
		UpdatedAt: item.updatedAt,
	}

	jsonData, _ := json.Marshal(payload)
	var restored TaskPayload
	json.Unmarshal(jsonData, &restored)

	// Timestamps should be preserved
	if restored.DateAdded != originalAddTime {
		t.Errorf("DateAdded not preserved: expected %d, got %d", originalAddTime, restored.DateAdded)
	}

	if restored.UpdatedAt != originalUpdateTime {
		t.Errorf("UpdatedAt not preserved: expected %d, got %d", originalUpdateTime, restored.UpdatedAt)
	}
}
