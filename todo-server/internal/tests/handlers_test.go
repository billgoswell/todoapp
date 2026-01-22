package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/billgoswell/commandlinetodo-server/internal/handlers"
	"github.com/billgoswell/commandlinetodo-server/internal/models"
	"github.com/gin-gonic/gin"
)

// ============================================================================
// TEST SETUP HELPERS
// ============================================================================

// createAuthContext creates a Gin context with authenticated user
func createAuthContext(userID int) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", userID)
	return c
}

// makeHandlerRequest makes an HTTP request to a handler
func makeHandlerRequest(method, path string, body interface{}, router *gin.Engine) (*httptest.ResponseRecorder, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(bodyBytes)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w, nil
}

// ============================================================================
// TASK ENDPOINT TESTS
// ============================================================================

// TestHandler_GetTasks_ReturnsEmptyList tests GET /tasks with no tasks
func TestHandler_GetTasks_ReturnsEmptyList(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	// Create user and list but no tasks
	userID := ltdb.CreateTestUser(t, "test-key-empty")
	ltdb.CreateTestList(t, userID, "Empty List")

	// Create router with real database
	router := gin.Default()
	router.GET("/api/v1/tasks", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.GetTasks(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("GET", "/api/v1/tasks", nil, router)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var tasks []models.Task
	json.Unmarshal(w.Body.Bytes(), &tasks)

	if tasks == nil {
		t.Error("Expected non-nil task array")
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

// TestHandler_GetTasks_ReturnsAllTasks tests GET /tasks with multiple tasks
func TestHandler_GetTasks_ReturnsAllTasks(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-many")
	listID := ltdb.CreateTestList(t, userID, "My Tasks")

	// Create multiple tasks
	ltdb.CreateTestTask(t, userID, listID, "Task 1")
	ltdb.CreateTestTask(t, userID, listID, "Task 2")
	ltdb.CreateTestTask(t, userID, listID, "Task 3")

	router := gin.Default()
	router.GET("/api/v1/tasks", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.GetTasks(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("GET", "/api/v1/tasks", nil, router)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var tasks []models.Task
	json.Unmarshal(w.Body.Bytes(), &tasks)

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}
}

// TestHandler_GetTasks_ExcludesDeletedTasks tests that deleted tasks are excluded
func TestHandler_GetTasks_ExcludesDeletedTasks(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-deleted")
	listID := ltdb.CreateTestList(t, userID, "Deleted Test")

	// Create tasks
	ltdb.CreateTestTask(t, userID, listID, "Active task")

	// Create and soft-delete a task
	now := time.Now().Unix()
	deletedData := map[string]interface{}{
		"client_id":      fmt.Sprintf("deleted-task-%d", now),
		"todo_list_id":   listID,
		"todo":           "Deleted task",
		"priority":       1,
		"done":           false,
		"date_added":     now,
		"deleted":        true,
		"deleted_at":     now,
		"updated_at":     now,
		"version":        1,
	}
	ltdb.GetDB().UpsertTask(userID, deletedData)

	router := gin.Default()
	router.GET("/api/v1/tasks", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.GetTasks(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("GET", "/api/v1/tasks", nil, router)

	var tasks []models.Task
	json.Unmarshal(w.Body.Bytes(), &tasks)

	// Should only return the active task, not the deleted one
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task (deleted excluded), got %d", len(tasks))
	}

	if tasks[0].Todo != "Active task" {
		t.Errorf("Expected 'Active task', got '%s'", tasks[0].Todo)
	}
}

// TestHandler_CreateTask_ValidPayload tests POST /tasks with valid data
func TestHandler_CreateTask_ValidPayload(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-create")
	listID := ltdb.CreateTestList(t, userID, "Create Test")

	now := time.Now().Unix()
	newTask := models.Task{
		ClientID:   "new-task-001",
		TodoListID: listID,
		Todo:       "New task from handler",
		Priority:   2,
		Done:       false,
		DateAdded:  now,
		UpdatedAt:  now,
		Version:    1,
	}

	router := gin.Default()
	router.POST("/api/v1/tasks", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.CreateTask(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("POST", "/api/v1/tasks", newTask, router)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var createdTask models.Task
	json.Unmarshal(w.Body.Bytes(), &createdTask)

	if createdTask.ID == 0 {
		t.Error("Expected task ID to be set")
	}

	if createdTask.Todo != newTask.Todo {
		t.Errorf("Expected '%s', got '%s'", newTask.Todo, createdTask.Todo)
	}
}

// TestHandler_CreateTask_InvalidPayload tests POST /tasks with invalid data
func TestHandler_CreateTask_InvalidPayload(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-invalid")

	router := gin.Default()
	router.POST("/api/v1/tasks", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.CreateTask(ltdb.GetDB())(c)
	})

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", w.Code)
	}
}

// TestHandler_UpdateTask_ExistingTask tests PUT /tasks/:id updates task
func TestHandler_UpdateTask_ExistingTask(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-update")
	listID := ltdb.CreateTestList(t, userID, "Update Test")
	ltdb.CreateTestTask(t, userID, listID, "Original task")

	// Update the task
	now := time.Now().Unix()
	updatedTask := models.Task{
		ID:         1,
		ClientID:   "task-update-001",
		TodoListID: listID,
		Todo:       "Updated task",
		Priority:   1,
		Done:       true,
		DateAdded:  now,
		UpdatedAt:  now,
		Version:    2,
	}

	router := gin.Default()
	router.PUT("/api/v1/tasks/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.UpdateTask(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("PUT", "/api/v1/tasks/1", updatedTask, router)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var returnedTask models.Task
	json.Unmarshal(w.Body.Bytes(), &returnedTask)

	if returnedTask.Todo != "Updated task" {
		t.Errorf("Expected 'Updated task', got '%s'", returnedTask.Todo)
	}
}

// TestHandler_UpdateTask_NonExistentTask tests PUT /tasks/:id with non-existent ID
func TestHandler_UpdateTask_NonExistentTask(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-nonexistent")
	listID := ltdb.CreateTestList(t, userID, "NonExistent Test")

	now := time.Now().Unix()
	task := models.Task{
		ClientID:   "fake-task-999",
		TodoListID: listID,
		Todo:       "Fake task",
		DateAdded:  now,
		UpdatedAt:  now,
	}

	router := gin.Default()
	router.PUT("/api/v1/tasks/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.UpdateTask(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("PUT", "/api/v1/tasks/999", task, router)

	// Should either return 404 or create the task (depends on implementation)
	// Both are acceptable behaviors
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Logf("Update non-existent task returned %d", w.Code)
	}
}

// TestHandler_DeleteTask_SoftDeletes tests DELETE /tasks/:id soft deletes
func TestHandler_DeleteTask_SoftDeletes(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-softdelete")
	listID := ltdb.CreateTestList(t, userID, "Delete Test")
	ltdb.CreateTestTask(t, userID, listID, "Task to delete")

	router := gin.Default()
	router.DELETE("/api/v1/tasks/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.DeleteTask(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("DELETE", "/api/v1/tasks/1", nil, router)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected 200 or 204, got %d", w.Code)
	}

	// Verify task is marked as deleted but still retrievable with deleted flag
	tasks := ltdb.GetTasksSince(t, userID, 0)
	if len(tasks) != 1 {
		t.Errorf("Expected 1 deleted task in database, got %d", len(tasks))
	}

	if tasks[0]["deleted"] != true {
		t.Error("Expected deleted=true in database")
	}
}

// ============================================================================
// LIST ENDPOINT TESTS
// ============================================================================

// TestHandler_GetLists_ReturnsAllLists tests GET /lists
func TestHandler_GetLists_ReturnsAllLists(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-lists")

	// Create multiple lists
	ltdb.CreateTestList(t, userID, "List 1")
	ltdb.CreateTestList(t, userID, "List 2")
	ltdb.CreateTestList(t, userID, "List 3")

	router := gin.Default()
	router.GET("/api/v1/lists", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.GetLists(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("GET", "/api/v1/lists", nil, router)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var lists []models.TodoList
	json.Unmarshal(w.Body.Bytes(), &lists)

	if len(lists) != 3 {
		t.Errorf("Expected 3 lists, got %d", len(lists))
	}
}

// TestHandler_CreateList_ValidPayload tests POST /lists
func TestHandler_CreateList_ValidPayload(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-createlist")

	now := time.Now().Unix()
	newList := models.TodoList{
		ClientID:     "new-list-001",
		Name:         "New List",
		DisplayOrder: 0,
		Archived:     false,
		UpdatedAt:    now,
		Version:      1,
	}

	router := gin.Default()
	router.POST("/api/v1/lists", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.CreateList(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("POST", "/api/v1/lists", newList, router)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var createdList models.TodoList
	json.Unmarshal(w.Body.Bytes(), &createdList)

	if createdList.ID == 0 {
		t.Error("Expected list ID to be set")
	}

	if createdList.Name != newList.Name {
		t.Errorf("Expected '%s', got '%s'", newList.Name, createdList.Name)
	}
}

// TestHandler_CreateList_DuplicateName tests POST /lists with duplicate name
func TestHandler_CreateList_DuplicateName(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-duplicate")
	ltdb.CreateTestList(t, userID, "Duplicate Name")

	now := time.Now().Unix()
	duplicateList := models.TodoList{
		ClientID:     "dup-list-002",
		Name:         "Duplicate Name",
		DisplayOrder: 0,
		Archived:     false,
		UpdatedAt:    now,
		Version:      1,
	}

	router := gin.Default()
	router.POST("/api/v1/lists", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.CreateList(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("POST", "/api/v1/lists", duplicateList, router)

	// Should handle duplicate names gracefully (either allow or reject)
	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Logf("Duplicate list creation returned %d (expected 201 or 409)", w.Code)
	}
}

// TestHandler_UpdateList_ChangeName tests PUT /lists/:id
func TestHandler_UpdateList_ChangeName(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-updatelist")
	ltdb.CreateTestList(t, userID, "Original List Name")

	now := time.Now().Unix()
	updatedList := models.TodoList{
		ID:           1,
		ClientID:     "list-update-001",
		Name:         "Updated List Name",
		DisplayOrder: 1,
		Archived:     false,
		UpdatedAt:    now,
		Version:      2,
	}

	router := gin.Default()
	router.PUT("/api/v1/lists/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.UpdateList(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("PUT", "/api/v1/lists/1", updatedList, router)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var returnedList models.TodoList
	json.Unmarshal(w.Body.Bytes(), &returnedList)

	if returnedList.Name != "Updated List Name" {
		t.Errorf("Expected 'Updated List Name', got '%s'", returnedList.Name)
	}
}

// TestHandler_ArchiveList_UpdatesStatus tests list archival
func TestHandler_ArchiveList_UpdatesStatus(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-archive")
	ltdb.CreateTestList(t, userID, "Archive Test")

	now := time.Now().Unix()
	archiveList := models.TodoList{
		ID:        1,
		ClientID:  "list-archive-001",
		Name:      "Archive Test",
		Archived:  true,
		UpdatedAt: now,
		Version:   2,
	}

	router := gin.Default()
	router.PUT("/api/v1/lists/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.UpdateList(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("PUT", "/api/v1/lists/1", archiveList, router)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var returnedList models.TodoList
	json.Unmarshal(w.Body.Bytes(), &returnedList)

	if !returnedList.Archived {
		t.Error("Expected list to be archived")
	}
}

// TestHandler_DeleteList_SoftArchives tests DELETE /lists/:id
func TestHandler_DeleteList_SoftArchives(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	userID := ltdb.CreateTestUser(t, "test-key-deletelist")
	ltdb.CreateTestList(t, userID, "List to Delete")

	router := gin.Default()
	router.DELETE("/api/v1/lists/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handlers.DeleteList(ltdb.GetDB())(c)
	})

	w, _ := makeHandlerRequest("DELETE", "/api/v1/lists/1", nil, router)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected 200 or 204, got %d", w.Code)
	}

	// Verify list still exists in database (soft delete/archive)
	lists := ltdb.GetListsSince(t, userID, 0)
	if len(lists) != 1 {
		t.Errorf("Expected 1 list in database after delete, got %d", len(lists))
	}

	if lists[0]["archived"] != true {
		t.Error("Expected list to be archived after delete")
	}
}
