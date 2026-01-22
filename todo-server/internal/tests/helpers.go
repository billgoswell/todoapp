package tests

// CONSOLIDATION NOTE (Phase 1.3 - Refactoring):
// Test helpers are conceptually consolidated across two locations:
// 1. Server: internal/tests/helpers.go (database/fixtures for unit tests)
// 2. Integration: integration-tests/utils/test_helpers.go (API/CLI testing)
//
// Since these are in different modules with different dependencies,
// they're maintained separately but conceptually as one system:
// - Both use shared-types for Task/TodoList/User models
// - Both provide builders (NewTestTask, NewTestList) using shared-types
// - Server tests import from internal/tests/
// - Integration tests import from integration-tests/utils/
//
// This approach maintains type safety while minimizing duplication.

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/billgoswell/commandlinetodo-server/internal/db"
	"github.com/billgoswell/commandlinetodo-server/internal/models"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// LocalTestDB represents a local PostgreSQL test database using testcontainers
type LocalTestDB struct {
	db        *db.DB
	container testcontainers.Container
	ctx       context.Context
}

// isDockerAvailable checks if Docker is available by checking for Docker socket
func isDockerAvailable() bool {
	// Check if Docker socket exists (most common setup)
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		return true
	}

	// Check alternative Docker socket location (some systems)
	if _, err := os.Stat("/run/docker.sock"); err == nil {
		return true
	}

	// Check for Docker via environment variable
	if os.Getenv("DOCKER_HOST") != "" {
		return true
	}

	return false
}

// SetupLocalTestDB creates a test database with PostgreSQL container
func SetupLocalTestDB(t *testing.T) *LocalTestDB {
	// Check if Docker is available
	if !isDockerAvailable() {
		t.Skip("Docker not available - skipping local database tests. Use TEST_USE_LOCAL_SERVER=false for live server tests.")
	}

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
	database, err := db.NewDB(host, port.Port(), "testuser", "testpass", "testdb")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &LocalTestDB{
		db:        database,
		container: container,
		ctx:       ctx,
	}
}

// GetDB returns the database connection
func (ltdb *LocalTestDB) GetDB() *db.DB {
	return ltdb.db
}

// Cleanup closes the database and stops the container
func (ltdb *LocalTestDB) Cleanup() error {
	if ltdb.db != nil {
		if err := ltdb.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}
	if ltdb.container != nil {
		if err := ltdb.container.Terminate(ltdb.ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
	}
	return nil
}

// ============================================================================
// TEST FIXTURE HELPERS
// ============================================================================

// CreateTestUser creates a test user in the database
func (ltdb *LocalTestDB) CreateTestUser(t *testing.T, apiKey string) int {
	userID, err := ltdb.db.CreateUser(apiKey)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// CreateTestList creates a test list in the database
func (ltdb *LocalTestDB) CreateTestList(t *testing.T, userID int, name string) int {
	listData := map[string]interface{}{
		"client_id":     fmt.Sprintf("test-list-%d", time.Now().Unix()),
		"name":          name,
		"display_order": 0,
		"archived":      false,
		"updated_at":    time.Now().Unix(),
		"version":       1,
	}
	id, err := ltdb.db.UpsertList(userID, listData)
	if err != nil {
		t.Fatalf("Failed to create test list: %v", err)
	}
	return id
}

// CreateTestTask creates a test task in the database
func (ltdb *LocalTestDB) CreateTestTask(t *testing.T, userID int, listID int, todo string) int {
	taskData := map[string]interface{}{
		"client_id":      fmt.Sprintf("test-task-%d", time.Now().Unix()),
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
	id, err := ltdb.db.UpsertTask(userID, taskData)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	return id
}

// GetUserByAPIKey retrieves a user by their API key
func (ltdb *LocalTestDB) GetUserByAPIKey(t *testing.T, apiKey string) int {
	userID, err := ltdb.db.GetUserByAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to get user by API key: %v", err)
	}
	return userID
}

// GetTasksSince retrieves tasks modified since a given timestamp
func (ltdb *LocalTestDB) GetTasksSince(t *testing.T, userID int, since int64) []map[string]interface{} {
	tasks, err := ltdb.db.GetTasksSince(userID, since)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	return tasks
}

// GetListsSince retrieves lists modified since a given timestamp
func (ltdb *LocalTestDB) GetListsSince(t *testing.T, userID int, since int64) []map[string]interface{} {
	lists, err := ltdb.db.GetListsSince(userID, since)
	if err != nil {
		t.Fatalf("Failed to get lists: %v", err)
	}
	return lists
}

// ============================================================================
// TEST DATA BUILDERS
// ============================================================================

// NewTestTask creates a task with sensible defaults
func NewTestTask(listID int) models.Task {
	now := time.Now().Unix()
	return models.Task{
		ClientID:   fmt.Sprintf("task-%d", now),
		TodoListID: listID,
		Todo:       "Test task",
		Priority:   3,
		Done:       false,
		DateAdded:  now,
		UpdatedAt:  now,
		Version:    1,
	}
}

// NewTestList creates a list with sensible defaults
func NewTestList() models.TodoList {
	now := time.Now().Unix()
	return models.TodoList{
		ClientID:     fmt.Sprintf("list-%d", now),
		Name:         "Test List",
		DisplayOrder: 0,
		Archived:     false,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}
}

// NewTestSyncRequest creates a sync request for testing
func NewTestSyncRequest(since int64) models.SyncRequest {
	return models.SyncRequest{Since: since}
}

// NewTestPushRequest creates a push request with test data
func NewTestPushRequest(tasks []models.Task, lists []models.TodoList) models.PushRequest {
	return models.PushRequest{
		Tasks: tasks,
		Lists: lists,
	}
}
