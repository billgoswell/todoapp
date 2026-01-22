package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/billgoswell/commandlinetodo-server/internal/handlers"
	"github.com/billgoswell/commandlinetodo-server/internal/models"
	"github.com/gin-gonic/gin"
)

// ============================================================================
// SYNC OPERATION TESTS
// ============================================================================

// TestSync_Pull_NoChanges tests pull with no changes since timestamp
func TestSync_Pull_NoChanges(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-sync")
	listID := ltdb.CreateTestList(t, userID, "Sync Test")

	// Create task
	ltdb.CreateTestTask(t, userID, listID, "Test task")

	// Query with future timestamp (no changes)
	futureTime := time.Now().Unix() + 10000
	syncReq := models.SyncRequest{Since: futureTime}

	router := gin.Default()
	router.POST("/api/v1/sync/pull", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.Pull(ltdb.GetDB())(c)
	})

	reqBody, _ := json.Marshal(syncReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/pull", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var syncResp models.SyncResponse
	json.Unmarshal(w.Body.Bytes(), &syncResp)

	if len(syncResp.Tasks) != 0 {
		t.Errorf("Expected 0 tasks with future timestamp, got %d", len(syncResp.Tasks))
	}

	if len(syncResp.Lists) != 0 {
		t.Errorf("Expected 0 lists with future timestamp, got %d", len(syncResp.Lists))
	}
}

// TestSync_Pull_ReturnsChangedItems tests pull returns changed items
func TestSync_Pull_ReturnsChangedItems(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-pullchanges")
	listID := ltdb.CreateTestList(t, userID, "Pull Changes")

	// Create multiple tasks
	ltdb.CreateTestTask(t, userID, listID, "Task 1")
	ltdb.CreateTestTask(t, userID, listID, "Task 2")
	ltdb.CreateTestTask(t, userID, listID, "Task 3")

	// Pull all changes since epoch
	syncReq := models.SyncRequest{Since: 0}

	router := gin.Default()
	router.POST("/api/v1/sync/pull", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.Pull(ltdb.GetDB())(c)
	})

	reqBody, _ := json.Marshal(syncReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/pull", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var syncResp models.SyncResponse
	json.Unmarshal(w.Body.Bytes(), &syncResp)

	if len(syncResp.Tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(syncResp.Tasks))
	}

	if len(syncResp.Lists) != 1 {
		t.Errorf("Expected 1 list, got %d", len(syncResp.Lists))
	}
}

// TestSync_Pull_FiltersByTimestamp tests that pull filters by timestamp
func TestSync_Pull_FiltersByTimestamp(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-timestamp-filter")
	listID := ltdb.CreateTestList(t, userID, "Timestamp Filter")

	now := time.Now().Unix()

	// Create old task
	oldTaskData := map[string]interface{}{
		"client_id":      "task-old-001",
		"todo_list_id":   listID,
		"todo":           "Old task",
		"priority":       1,
		"done":           false,
		"date_added":     now - 1000,
		"updated_at":     now - 1000,
		"version":        1,
	}
	ltdb.GetDB().UpsertTask(userID, oldTaskData)

	// Create new task
	newTaskData := map[string]interface{}{
		"client_id":      "task-new-001",
		"todo_list_id":   listID,
		"todo":           "New task",
		"priority":       1,
		"done":           false,
		"date_added":     now,
		"updated_at":     now,
		"version":        1,
	}
	ltdb.GetDB().UpsertTask(userID, newTaskData)

	// Pull with timestamp between old and new
	middleTime := now - 500
	syncReq := models.SyncRequest{Since: middleTime}

	router := gin.Default()
	router.POST("/api/v1/sync/pull", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.Pull(ltdb.GetDB())(c)
	})

	reqBody, _ := json.Marshal(syncReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/pull", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var syncResp models.SyncResponse
	json.Unmarshal(w.Body.Bytes(), &syncResp)

	// Should only get the new task
	if len(syncResp.Tasks) != 1 {
		t.Errorf("Expected 1 task after timestamp filter, got %d", len(syncResp.Tasks))
	}

	if syncResp.Tasks[0].Todo != "New task" {
		t.Errorf("Expected 'New task', got '%s'", syncResp.Tasks[0].Todo)
	}
}

// TestSync_Push_CreatesNewItems tests that push creates new items
func TestSync_Push_CreatesNewItems(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-pushcreate")
	listID := ltdb.CreateTestList(t, userID, "Push Create Test")

	now := time.Now().Unix()

	// Create push request with new items
	pushReq := models.PushRequest{
		Tasks: []models.Task{
			{
				ClientID:   "push-task-001",
				TodoListID: listID,
				Todo:       "Task from push",
				Priority:   2,
				Done:       false,
				DateAdded:  now,
				UpdatedAt:  now,
				Version:    1,
			},
		},
		Lists: []models.TodoList{
			{
				ClientID:     "push-list-001",
				Name:         "List from push",
				DisplayOrder: 1,
				Archived:     false,
				UpdatedAt:    now,
				Version:      1,
			},
		},
	}

	router := gin.Default()
	router.POST("/api/v1/sync/push", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.Push(ltdb.GetDB())(c)
	})

	reqBody, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify items were created
	tasks := ltdb.GetTasksSince(t, userID, 0)
	lists := ltdb.GetListsSince(t, userID, 0)

	if len(tasks) < 1 {
		t.Error("Task from push was not created")
	}

	if len(lists) < 2 {
		t.Error("List from push was not created")
	}
}

// TestSync_Push_UpdatesExistingItems tests that push updates items
func TestSync_Push_UpdatesExistingItems(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-pushupdate")
	listID := ltdb.CreateTestList(t, userID, "Push Update Test")

	// Create initial task
	ltdb.CreateTestTask(t, userID, listID, "Original task")

	now := time.Now().Unix()

	// Push update for the task
	pushReq := models.PushRequest{
		Tasks: []models.Task{
			{
				ClientID:   "task-update-via-push",
				TodoListID: listID,
				Todo:       "Updated via push",
				Priority:   1,
				Done:       true,
				DateAdded:  now,
				UpdatedAt:  now,
				Version:    2,
			},
		},
		Lists: []models.TodoList{},
	}

	router := gin.Default()
	router.POST("/api/v1/sync/push", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.Push(ltdb.GetDB())(c)
	})

	reqBody, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

// TestSync_Push_ConflictResolution tests last-write-wins conflict resolution
func TestSync_Push_ConflictResolution_LastWriteWins(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-conflict")
	listID := ltdb.CreateTestList(t, userID, "Conflict Test")

	now := time.Now().Unix()

	// Server already has newer version
	newerTaskData := map[string]interface{}{
		"client_id":      "conflict-task-001",
		"todo_list_id":   listID,
		"todo":           "Newer version on server",
		"priority":       2,
		"done":           false,
		"date_added":     now,
		"updated_at":     now + 100, // Newer timestamp
		"version":        2,
	}
	ltdb.GetDB().UpsertTask(userID, newerTaskData)

	// Client pushes older version
	pushReq := models.PushRequest{
		Tasks: []models.Task{
			{
				ClientID:   "conflict-task-001",
				TodoListID: listID,
				Todo:       "Older version from client",
				Priority:   1,
				Done:       true,
				DateAdded:  now,
				UpdatedAt:  now, // Older timestamp
				Version:    1,
			},
		},
		Lists: []models.TodoList{},
	}

	router := gin.Default()
	router.POST("/api/v1/sync/push", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.Push(ltdb.GetDB())(c)
	})

	reqBody, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	// Verify newer version is kept
	tasks := ltdb.GetTasksSince(t, userID, 0)
	if len(tasks) < 1 {
		t.Fatal("Task not found")
	}

	task := tasks[0]
	if task["todo"] != "Newer version on server" {
		t.Errorf("Expected 'Newer version on server', got '%v'", task["todo"])
	}
}

// TestSync_Push_HandlesDeletedItems tests that deleted items are handled
func TestSync_Push_HandlesDeletedItems(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-deleted-sync")
	listID := ltdb.CreateTestList(t, userID, "Deleted Sync Test")

	now := time.Now().Unix()

	// Push deleted task
	pushReq := models.PushRequest{
		Tasks: []models.Task{
			{
				ClientID:   "deleted-via-push-001",
				TodoListID: listID,
				Todo:       "Task to be deleted",
				Priority:   1,
				Done:       false,
				DateAdded:  now,
				Deleted:    true,
				DeletedAt:  &now,
				UpdatedAt:  now,
				Version:    1,
			},
		},
		Lists: []models.TodoList{},
	}

	router := gin.Default()
	router.POST("/api/v1/sync/push", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.Push(ltdb.GetDB())(c)
	})

	reqBody, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	// Verify deleted task is in database with deleted flag
	tasks := ltdb.GetTasksSince(t, userID, 0)
	if len(tasks) < 1 {
		t.Error("Deleted task should still exist in database")
	}

	task := tasks[0]
	if task["deleted"] != true {
		t.Error("Task should have deleted=true")
	}
}

// TestSync_SyncRoundTrip tests a full sync round trip
func TestSync_SyncRoundTrip(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-roundtrip")
	listID := ltdb.CreateTestList(t, userID, "RoundTrip Test")

	now := time.Now().Unix()

	// Step 1: Push changes
	pushReq := models.PushRequest{
		Tasks: []models.Task{
			{
				ClientID:   "roundtrip-task-001",
				TodoListID: listID,
				Todo:       "RoundTrip task",
				Priority:   2,
				Done:       false,
				DateAdded:  now,
				UpdatedAt:  now,
				Version:    1,
			},
		},
		Lists: []models.TodoList{},
	}

	router := gin.Default()
	router.POST("/api/v1/sync/push", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.Push(ltdb.GetDB())(c)
	})
	router.POST("/api/v1/sync/pull", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.Pull(ltdb.GetDB())(c)
	})

	// Push the changes
	reqBody, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Push failed with status %d", w.Code)
	}

	// Step 2: Pull changes (should see the pushed item)
	pullReq := models.SyncRequest{Since: 0}
	reqBody, _ = json.Marshal(pullReq)
	req = httptest.NewRequest("POST", "/api/v1/sync/pull", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var pullResp models.SyncResponse
	json.Unmarshal(w.Body.Bytes(), &pullResp)

	// Verify the pushed task is in the pull response
	found := false
	for _, task := range pullResp.Tasks {
		if task.Todo == "RoundTrip task" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Pushed task should appear in pull response")
	}
}
