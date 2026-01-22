package tests

import (
	"testing"
	"time"
)

// TestLocalDB_Health verifies database connection is healthy
func TestLocalDB_Health(t *testing.T) {
	// This test runs with local testcontainers database
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local database test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	if err := ltdb.GetDB().Ping(); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

// TestLocalDB_CreateUser tests user creation
func TestLocalDB_CreateUser(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local database test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	apiKey := "test-api-key-user-001"
	userID := ltdb.CreateTestUser(t, apiKey)

	if userID <= 0 {
		t.Errorf("Expected positive user ID, got %d", userID)
	}

	// Verify user can be retrieved
	retrievedID := ltdb.GetUserByAPIKey(t, apiKey)
	if retrievedID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, retrievedID)
	}
}

// TestLocalDB_CreateList tests list creation
func TestLocalDB_CreateList(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local database test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-list")
	listID := ltdb.CreateTestList(t, userID, "My Test List")

	if listID <= 0 {
		t.Errorf("Expected positive list ID, got %d", listID)
	}

	// Verify list can be retrieved
	lists := ltdb.GetListsSince(t, userID, 0)
	if len(lists) != 1 {
		t.Errorf("Expected 1 list, got %d", len(lists))
	}
}

// TestLocalDB_CreateTask tests task creation
func TestLocalDB_CreateTask(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local database test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-task")
	listID := ltdb.CreateTestList(t, userID, "My List")
	taskID := ltdb.CreateTestTask(t, userID, listID, "Do something")

	if taskID <= 0 {
		t.Errorf("Expected positive task ID, got %d", taskID)
	}

	// Verify task can be retrieved
	tasks := ltdb.GetTasksSince(t, userID, 0)
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0]["todo"] != "Do something" {
		t.Errorf("Expected 'Do something', got '%v'", tasks[0]["todo"])
	}
}

// TestLocalDB_ConflictResolution_LastWriteWins tests last-write-wins conflict resolution
func TestLocalDB_ConflictResolution_LastWriteWins(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local database test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-conflict")
	listID := ltdb.CreateTestList(t, userID, "Conflict Test List")

	now := time.Now().Unix()

	// Create initial task with newer timestamp
	taskData1 := map[string]interface{}{
		"client_id":      "task-conflict-001",
		"todo_list_id":   listID,
		"todo":           "Newer version",
		"priority":       2,
		"done":           false,
		"date_added":     now,
		"updated_at":     now + 100,
		"version":        2,
	}
	ltdb.GetDB().UpsertTask(userID, taskData1)

	// Try to update with older timestamp
	taskData2 := map[string]interface{}{
		"client_id":      "task-conflict-001",
		"todo_list_id":   listID,
		"todo":           "Older version",
		"priority":       1,
		"done":           true,
		"date_added":     now,
		"updated_at":     now, // Older timestamp
		"version":        1,
	}
	ltdb.GetDB().UpsertTask(userID, taskData2)

	// Retrieve and verify newer version is kept
	tasks := ltdb.GetTasksSince(t, userID, 0)
	if len(tasks) == 0 {
		t.Fatal("No tasks found after conflict resolution")
	}

	task := tasks[0]
	if task["todo"] != "Newer version" {
		t.Errorf("Expected 'Newer version', got '%v'", task["todo"])
	}
	if task["priority"] != 2 {
		t.Errorf("Expected priority 2, got %v", task["priority"])
	}
}

// TestLocalDB_MultipleUsers tests isolation between users
func TestLocalDB_MultipleUsers(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local database test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	user1ID := ltdb.CreateTestUser(t, "user1-key")
	user2ID := ltdb.CreateTestUser(t, "user2-key")

	// Create lists for each user
	user1ListID := ltdb.CreateTestList(t, user1ID, "User 1 List")
	user2ListID := ltdb.CreateTestList(t, user2ID, "User 2 List")

	// Create tasks for each user
	ltdb.CreateTestTask(t, user1ID, user1ListID, "User 1 Task")
	ltdb.CreateTestTask(t, user2ID, user2ListID, "User 2 Task")

	// Verify isolation
	user1Tasks := ltdb.GetTasksSince(t, user1ID, 0)
	user2Tasks := ltdb.GetTasksSince(t, user2ID, 0)

	if len(user1Tasks) != 1 {
		t.Errorf("Expected 1 task for user 1, got %d", len(user1Tasks))
	}
	if len(user2Tasks) != 1 {
		t.Errorf("Expected 1 task for user 2, got %d", len(user2Tasks))
	}

	if user1Tasks[0]["todo"] != "User 1 Task" {
		t.Error("User 1 has wrong task")
	}
	if user2Tasks[0]["todo"] != "User 2 Task" {
		t.Error("User 2 has wrong task")
	}
}

// TestLocalDB_SyncTimestamp tests that GetTasksSince filters by timestamp
func TestLocalDB_SyncTimestamp(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local database test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-timestamp")
	listID := ltdb.CreateTestList(t, userID, "Timestamp Test List")

	now := time.Now().Unix()

	// Create task with old timestamp
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

	// Query with timestamp between old and new
	middleTime := now - 500
	tasks := ltdb.GetTasksSince(t, userID, middleTime)

	// Should not find the old task
	if len(tasks) > 0 {
		t.Errorf("Expected 0 tasks after timestamp %d, but got %d", middleTime, len(tasks))
	}

	// Create new task
	newTaskData := map[string]interface{}{
		"client_id":      "task-new-001",
		"todo_list_id":   listID,
		"todo":           "New task",
		"priority":       2,
		"done":           false,
		"date_added":     now,
		"updated_at":     now,
		"version":        1,
	}
	ltdb.GetDB().UpsertTask(userID, newTaskData)

	// Query with same middle timestamp
	tasks = ltdb.GetTasksSince(t, userID, middleTime)

	// Should find only the new task
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task after timestamp %d, got %d", middleTime, len(tasks))
	}
	if tasks[0]["todo"] != "New task" {
		t.Errorf("Expected 'New task', got '%v'", tasks[0]["todo"])
	}
}

// TestLocalDB_UpdateTask tests task updates
func TestLocalDB_UpdateTask(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local database test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-update")
	listID := ltdb.CreateTestList(t, userID, "Update Test List")
	ltdb.CreateTestTask(t, userID, listID, "Original task")

	now := time.Now().Unix()

	// Update task with new values
	updateData := map[string]interface{}{
		"client_id":      "task-update-001",
		"todo_list_id":   listID,
		"todo":           "Updated task",
		"priority":       1,
		"done":           true,
		"date_added":     now,
		"updated_at":     now,
		"version":        2,
	}
	ltdb.GetDB().UpsertTask(userID, updateData)

	// Retrieve and verify
	tasks := ltdb.GetTasksSince(t, userID, 0)
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task["todo"] != "Updated task" {
		t.Errorf("Expected 'Updated task', got '%v'", task["todo"])
	}
	if task["priority"] != 1 {
		t.Errorf("Expected priority 1, got %v", task["priority"])
	}
	if task["done"] != true {
		t.Errorf("Expected done=true, got %v", task["done"])
	}
}
