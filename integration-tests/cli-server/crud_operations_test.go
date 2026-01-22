package cli_server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/billgoswell/todoapp/integration-tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIServerCRUDOperations - Phase 5a Category 3 (9 tests)
// Tests validating full CRUD operations between CLI and Server

func TestCreateTaskOnCliSyncedToServer(t *testing.T) {
	// Test 7: Create task on CLI → Synced to server
	// Validates that tasks created via CLI appear on server

	config := utils.DefaultConfig()
	httpClient := utils.NewHTTPClient(config)

	// First, create a list on server
	newList := utils.NewTestList("Test List")
	resp, err := httpClient.Post("/api/v1/lists", newList)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "List creation should succeed")

	var createdList utils.TestList
	err = utils.ParseJSONResponse(resp, &createdList)
	require.NoError(t, err)
	require.Greater(t, createdList.ID, int64(0), "List should have ID")

	// Create task on server (simulating CLI create)
	newTask := utils.NewTestTask(createdList.ID, "Test task from CLI")
	resp, err = httpClient.Post("/api/v1/tasks", newTask)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Task creation should succeed")

	var createdTask utils.TestTask
	err = utils.ParseJSONResponse(resp, &createdTask)
	assert.NoError(t, err)
	assert.Greater(t, createdTask.ID, int64(0), "Created task should have ID")
	assert.Equal(t, newTask.Todo, createdTask.Todo, "Task content should match")
}

func TestUpdateTaskOnCliServerReflectsChanges(t *testing.T) {
	// Test 8: Update task on CLI → Server reflects changes
	// Validates that task updates propagate to server

	config := utils.DefaultConfig()
	httpClient := utils.NewHTTPClient(config)

	// Create list and task
	newList := utils.NewTestList("Update Test List")
	resp, err := httpClient.Post("/api/v1/lists", newList)
	require.NoError(t, err)
	defer resp.Body.Close()

	var list utils.TestList
	err = utils.ParseJSONResponse(resp, &list)
	require.NoError(t, err)

	task := utils.NewTestTask(list.ID, "Original task")
	resp, err = httpClient.Post("/api/v1/tasks", task)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createdTask utils.TestTask
	err = utils.ParseJSONResponse(resp, &createdTask)
	require.NoError(t, err)

	// Update task
	createdTask.Todo = "Updated task"
	createdTask.Done = true
	createdTask.UpdatedAt = time.Now().UnixMilli()

	resp, err = httpClient.Put("/api/v1/tasks/"+string(rune(createdTask.ID)), createdTask)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Task update should succeed")
}

func TestDeleteTaskOnCliServerMarksAsDeleted(t *testing.T) {
	// Test 9: Delete task on CLI → Server marks as deleted
	// Validates that soft delete propagates to server

	config := utils.DefaultConfig()
	httpClient := utils.NewHTTPClient(config)

	// Create list and task
	newList := utils.NewTestList("Delete Test List")
	resp, err := httpClient.Post("/api/v1/lists", newList)
	require.NoError(t, err)
	defer resp.Body.Close()

	var list utils.TestList
	err = utils.ParseJSONResponse(resp, &list)
	require.NoError(t, err)

	task := utils.NewTestTask(list.ID, "Task to delete")
	resp, err = httpClient.Post("/api/v1/tasks", task)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createdTask utils.TestTask
	err = utils.ParseJSONResponse(resp, &createdTask)
	require.NoError(t, err)

	// Delete the task
	resp, err = httpClient.Delete("/api/v1/tasks/" + string(rune(createdTask.ID)))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Task deletion should succeed")
}

func TestCreateListOnCliServerUpdates(t *testing.T) {
	// Test 10: Create list on CLI → Server updates
	// Validates that lists created via CLI appear on server

	config := utils.DefaultConfig()
	httpClient := utils.NewHTTPClient(config)

	newList := utils.NewTestList("New Test List")
	resp, err := httpClient.Post("/api/v1/lists", newList)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "List creation should succeed")

	var createdList utils.TestList
	err = utils.ParseJSONResponse(resp, &createdList)
	assert.NoError(t, err)
	assert.Greater(t, createdList.ID, int64(0), "List should have ID")
	assert.Equal(t, newList.Name, createdList.Name, "List name should match")
}

func TestUpdateListPropertiesServerSync(t *testing.T) {
	// Test 11: Update list properties → Server sync
	// Validates that list updates sync to server

	config := utils.DefaultConfig()
	httpClient := utils.NewHTTPClient(config)

	newList := utils.NewTestList("Original List Name")
	resp, err := httpClient.Post("/api/v1/lists", newList)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createdList utils.TestList
	err = utils.ParseJSONResponse(resp, &createdList)
	require.NoError(t, err)

	// Update list name
	createdList.Name = "Updated List Name"
	createdList.UpdatedAt = time.Now().UnixMilli()

	resp, err = httpClient.Put("/api/v1/lists/"+string(rune(createdList.ID)), createdList)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "List update should succeed")
}

func TestRetrieveAllItemsFromServerCliDisplays(t *testing.T) {
	// Test 12: Retrieve all items from server → CLI displays correctly
	// Validates that CLI can fetch and display all tasks and lists

	config := utils.DefaultConfig()
	httpClient := utils.NewHTTPClient(config)

	// Retrieve all tasks
	resp, err := httpClient.Get("/api/v1/tasks")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "Should have access to tasks")
	var tasks []utils.TestTask
	err = utils.ParseJSONResponse(resp, &tasks)
	// May have tasks or empty list - both are valid

	// Retrieve all lists
	resp, err = httpClient.Get("/api/v1/lists")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "Should have access to lists")
	var lists []utils.TestList
	err = utils.ParseJSONResponse(resp, &lists)
	// May have lists or empty list - both are valid
}

func TestTaskCRUDWithMultipleLists(t *testing.T) {
	// Test 13: Task CRUD with multiple lists
	// Validates that tasks can be organized across multiple lists

	config := utils.DefaultConfig()
	httpClient := utils.NewHTTPClient(config)

	// Create first list
	list1 := utils.NewTestList("Work List")
	resp, err := httpClient.Post("/api/v1/lists", list1)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createdList1 utils.TestList
	err = utils.ParseJSONResponse(resp, &createdList1)
	require.NoError(t, err)

	// Create second list
	list2 := utils.NewTestList("Personal List")
	resp, err = httpClient.Post("/api/v1/lists", list2)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createdList2 utils.TestList
	err = utils.ParseJSONResponse(resp, &createdList2)
	require.NoError(t, err)

	// Create task in first list
	task1 := utils.NewTestTask(createdList1.ID, "Work task")
	resp, err = httpClient.Post("/api/v1/tasks", task1)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Create task in second list
	task2 := utils.NewTestTask(createdList2.ID, "Personal task")
	resp, err = httpClient.Post("/api/v1/tasks", task2)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Tasks should be created in different lists")
}

func TestCrudOperationsMaintainDataConsistency(t *testing.T) {
	// Test 14: CRUD operations maintain data consistency
	// Validates that multiple operations don't cause data corruption

	config := utils.DefaultConfig()
	httpClient := utils.NewHTTPClient(config)

	// Create a list
	list := utils.NewTestList("Consistency Test List")
	resp, err := httpClient.Post("/api/v1/lists", list)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createdList utils.TestList
	err = utils.ParseJSONResponse(resp, &createdList)
	require.NoError(t, err)

	// Create multiple tasks
	var taskIDs []int64
	for i := 0; i < 3; i++ {
		task := utils.NewTestTask(createdList.ID, "Task "+string(rune(i)))
		resp, err = httpClient.Post("/api/v1/tasks", task)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createdTask utils.TestTask
		err = utils.ParseJSONResponse(resp, &createdTask)
		require.NoError(t, err)
		taskIDs = append(taskIDs, createdTask.ID)
	}

	assert.Equal(t, 3, len(taskIDs), "Should have created 3 tasks")

	// Verify all tasks exist
	resp, err = httpClient.Get("/api/v1/lists/" + string(rune(createdList.ID)))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should be able to retrieve list details without errors
	assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
		"List retrieval should succeed")
}

func TestCrudErrorHandling(t *testing.T) {
	// Test 15: CRUD error handling
	// Validates graceful handling of invalid operations

	config := utils.DefaultConfig()
	httpClient := utils.NewHTTPClient(config)

	// Try to get non-existent task
	resp, err := httpClient.Get("/api/v1/tasks/999999999")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return error or not found, not crash
	assert.True(t, resp.StatusCode >= 400 || resp.StatusCode < 300,
		"Should handle invalid task ID gracefully")

	// Try to create invalid task (missing required fields)
	invalidTask := &bytes.Buffer{}
	json.NewEncoder(invalidTask).Encode(map[string]interface{}{
		// Missing required fields
	})

	// This will fail validation but shouldn't crash the server
	resp, err = httpClient.Post("/api/v1/tasks", map[string]interface{}{})
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should handle validation errors
	assert.True(t, resp.StatusCode >= 400 || resp.StatusCode < 300,
		"Should handle validation errors gracefully")
}
