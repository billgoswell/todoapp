package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB represents a test database instance
type TestDB struct {
	db        *DB
	container testcontainers.Container
	ctx       context.Context
}

// setupTestDB creates a test database connection with a PostgreSQL container
func setupTestDB(t *testing.T) *TestDB {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Get host and port
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Connect to database
	db, err := NewDB(host, port.Port(), "testuser", "testpass", "testdb")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := db.Migrate(); err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &TestDB{
		db:        db,
		container: container,
		ctx:       ctx,
	}
}

// Cleanup closes the database and stops the container
func (tdb *TestDB) Cleanup() error {
	if tdb.db != nil {
		if err := tdb.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}
	if tdb.container != nil {
		if err := tdb.container.Terminate(tdb.ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
	}
	return nil
}

// Helper: Create test user
func createTestUser(t *testing.T, db *DB, apiKey string) int {
	userID, err := db.CreateUser(apiKey)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// Helper: Create test list
func createTestList(t *testing.T, db *DB, userID int, name string) int {
	listData := map[string]interface{}{
		"client_id":     "test-list-" + time.Now().Format("20060102150405"),
		"name":          name,
		"display_order": 0,
		"archived":      false,
		"updated_at":    time.Now().Unix(),
		"version":       1,
	}
	id, err := db.UpsertList(userID, listData)
	if err != nil {
		t.Fatalf("Failed to create test list: %v", err)
	}
	return id
}

// Helper: Create test task
func createTestTask(t *testing.T, db *DB, userID int, listID int, todo string) int {
	taskData := map[string]interface{}{
		"client_id":      "test-task-" + time.Now().Format("20060102150405"),
		"todo_list_id":   listID,
		"todo":           todo,
		"priority":       3,
		"done":           false,
		"date_added":     time.Now().Unix(),
		"date_completed": nil,
		"due_date":       nil,
		"deleted":        false,
		"deleted_at":     nil,
		"updated_at":     time.Now().Unix(),
		"version":        1,
	}
	id, err := db.UpsertTask(userID, taskData)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	return id
}

// ============================================================================
// DATABASE LAYER TESTS
// ============================================================================

// Connection Tests
func TestDB_Connection(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	// Test: Database connection establishment
	// Expected: Connection succeeds and ping returns no error
	if err := tdb.db.Ping(); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestDB_MigrationCreatesSchema(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	// Verify tables exist
	tables := []string{"users", "todo_lists", "tasks"}
	for _, table := range tables {
		var exists bool
		err := tdb.db.conn.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_name = $1
			)
		`, table).Scan(&exists)
		if err != nil || !exists {
			t.Errorf("Table %s does not exist: %v", table, err)
		}
	}
}

// ============================================================================
// USER MANAGEMENT TESTS
// ============================================================================

func TestCreateUser_NewUser(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	apiKey := "test-api-key-12345"
	userID, err := tdb.db.CreateUser(apiKey)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if userID <= 0 {
		t.Errorf("Expected positive user ID, got %d", userID)
	}

	// Verify user can be retrieved
	retrievedID, err := tdb.db.GetUserByAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}

	if retrievedID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, retrievedID)
	}
}

func TestGetUserByAPIKey_ExistingUser(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	apiKey := "test-api-key-existing"
	expectedID, _ := tdb.db.CreateUser(apiKey)

	retrievedID, err := tdb.db.GetUserByAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}

	if retrievedID != expectedID {
		t.Errorf("Expected user ID %d, got %d", expectedID, retrievedID)
	}
}

func TestGetUserByAPIKey_NonExistentUser(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	_, err := tdb.db.GetUserByAPIKey("non-existent-key")
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
	if err != sql.ErrNoRows {
		// Some error expected, which is fine
	}
}

// ============================================================================
// TASK TESTS
// ============================================================================

func TestUpsertTask_NewTask(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	userID := createTestUser(t, tdb.db, "test-key-1")
	listID := createTestList(t, tdb.db, userID, "Test List")

	taskData := map[string]interface{}{
		"client_id":      "task-new-001",
		"todo_list_id":   listID,
		"todo":           "Test task",
		"priority":       2,
		"done":           false,
		"date_added":     time.Now().Unix(),
		"updated_at":     time.Now().Unix(),
		"version":        1,
	}

	taskID, err := tdb.db.UpsertTask(userID, taskData)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	if taskID <= 0 {
		t.Errorf("Expected positive task ID, got %d", taskID)
	}

	// Verify task can be retrieved
	tasks, err := tdb.db.GetTasksSince(userID, 0)
	if err != nil {
		t.Fatalf("Failed to retrieve tasks: %v", err)
	}

	if len(tasks) == 0 {
		t.Error("Expected at least one task in results")
	}
}

func TestUpsertTask_UpdateExistingTask(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	userID := createTestUser(t, tdb.db, "test-key-2")
	listID := createTestList(t, tdb.db, userID, "Test List 2")

	// Create initial task
	taskData := map[string]interface{}{
		"client_id":      "task-update-001",
		"todo_list_id":   listID,
		"todo":           "Original task",
		"priority":       1,
		"done":           false,
		"date_added":     time.Now().Unix(),
		"updated_at":     time.Now().Unix(),
		"version":        1,
	}
	taskID1, _ := tdb.db.UpsertTask(userID, taskData)

	// Update task with newer timestamp
	time.Sleep(100 * time.Millisecond)
	updatedData := map[string]interface{}{
		"client_id":      "task-update-001",
		"todo_list_id":   listID,
		"todo":           "Updated task",
		"priority":       3,
		"done":           true,
		"date_added":     taskData["date_added"],
		"updated_at":     time.Now().Unix(),
		"version":        2,
	}
	taskID2, err := tdb.db.UpsertTask(userID, updatedData)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	if taskID1 != taskID2 {
		t.Errorf("Expected same task ID, got %d and %d", taskID1, taskID2)
	}
}

func TestUpsertTask_ConflictResolution_LastWriteWins(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	userID := createTestUser(t, tdb.db, "test-key-conflict")
	listID := createTestList(t, tdb.db, userID, "Test List Conflict")

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
	tdb.db.UpsertTask(userID, taskData1)

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
	tdb.db.UpsertTask(userID, taskData2)

	// Retrieve and verify newer version is kept
	tasks, err := tdb.db.GetTasksSince(userID, 0)
	if err != nil {
		t.Fatalf("Failed to retrieve tasks: %v", err)
	}

	if len(tasks) == 0 {
		t.Fatal("No tasks found")
	}

	task := tasks[0]
	if task["todo"] != "Newer version" {
		t.Errorf("Expected 'Newer version', got '%s'", task["todo"])
	}
	if task["priority"] != 2 {
		t.Errorf("Expected priority 2, got %v", task["priority"])
	}
}

func TestGetTasksSince_NoChanges(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	userID := createTestUser(t, tdb.db, "test-key-since")
	listID := createTestList(t, tdb.db, userID, "Test List Since")

	// Create task
	createTestTask(t, tdb.db, userID, listID, "Test task")

	// Query with future timestamp
	futureTime := time.Now().Unix() + 10000
	tasks, err := tdb.db.GetTasksSince(userID, futureTime)
	if err != nil {
		t.Fatalf("Failed to retrieve tasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks with future timestamp, got %d", len(tasks))
	}
}

func TestGetTasksSince_AllTasks(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	userID := createTestUser(t, tdb.db, "test-key-all")
	listID := createTestList(t, tdb.db, userID, "Test List All")

	// Create multiple tasks
	createTestTask(t, tdb.db, userID, listID, "Task 1")
	createTestTask(t, tdb.db, userID, listID, "Task 2")
	createTestTask(t, tdb.db, userID, listID, "Task 3")

	// Query all tasks
	tasks, err := tdb.db.GetTasksSince(userID, 0)
	if err != nil {
		t.Fatalf("Failed to retrieve tasks: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}
}

// ============================================================================
// LIST TESTS
// ============================================================================

func TestUpsertList_NewList(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	userID := createTestUser(t, tdb.db, "test-key-list-new")

	listData := map[string]interface{}{
		"client_id":     "list-new-001",
		"name":          "New List",
		"display_order": 0,
		"archived":      false,
		"updated_at":    time.Now().Unix(),
		"version":       1,
	}

	listID, err := tdb.db.UpsertList(userID, listData)
	if err != nil {
		t.Fatalf("Failed to insert list: %v", err)
	}

	if listID <= 0 {
		t.Errorf("Expected positive list ID, got %d", listID)
	}
}

func TestUpsertList_UpdateExistingList(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	userID := createTestUser(t, tdb.db, "test-key-list-update")

	// Create initial list
	listData := map[string]interface{}{
		"client_id":     "list-update-001",
		"name":          "Original Name",
		"display_order": 0,
		"archived":      false,
		"updated_at":    time.Now().Unix(),
		"version":       1,
	}
	listID1, _ := tdb.db.UpsertList(userID, listData)

	// Update list
	time.Sleep(100 * time.Millisecond)
	updatedData := map[string]interface{}{
		"client_id":     "list-update-001",
		"name":          "Original Name",
		"display_order": 1,
		"archived":      true,
		"updated_at":    time.Now().Unix(),
		"version":       2,
	}
	listID2, err := tdb.db.UpsertList(userID, updatedData)
	if err != nil {
		t.Fatalf("Failed to update list: %v", err)
	}

	if listID1 != listID2 {
		t.Errorf("Expected same list ID, got %d and %d", listID1, listID2)
	}
}

func TestGetListsSince_AllLists(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.Cleanup()

	userID := createTestUser(t, tdb.db, "test-key-lists-all")

	// Create multiple lists
	createTestList(t, tdb.db, userID, "List 1")
	createTestList(t, tdb.db, userID, "List 2")
	createTestList(t, tdb.db, userID, "List 3")

	// Query all lists
	lists, err := tdb.db.GetListsSince(userID, 0)
	if err != nil {
		t.Fatalf("Failed to retrieve lists: %v", err)
	}

	if len(lists) != 3 {
		t.Errorf("Expected 3 lists, got %d", len(lists))
	}
}
