package handlers

import (
	"testing"
	"time"

	"github.com/billgoswell/commandlinetodo-server/internal/models"
)

// TestConvertTaskMaps verifies task conversion from maps to models
func TestConvertTaskMaps(t *testing.T) {
	tasks := []map[string]interface{}{
		{
			"id":              1,
			"client_id":       "uuid-123",
			"todo_list_id":    1,
			"todo":            "Test task",
			"priority":        3,
			"done":            false,
			"deleted":         false,
			"date_added":      int64(1000),
			"date_completed":  nil,
			"due_date":        nil,
			"deleted_at":      nil,
			"created_at":      int64(1000),
			"updated_at":      int64(1001),
			"version":         1,
		},
	}

	result := convertTaskMaps(tasks)

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}

	task := result[0]
	if task.ID != 1 {
		t.Errorf("expected task id 1, got %d", task.ID)
	}
	if task.ClientID != "uuid-123" {
		t.Errorf("expected client_id 'uuid-123', got %s", task.ClientID)
	}
	if task.Todo != "Test task" {
		t.Errorf("expected todo 'Test task', got %s", task.Todo)
	}
	if task.Done {
		t.Error("expected done to be false")
	}
	if task.UpdatedAt != 1001 {
		t.Errorf("expected updated_at 1001, got %d", task.UpdatedAt)
	}
}

// TestConvertListMaps verifies list conversion from maps to models
func TestConvertListMaps(t *testing.T) {
	lists := []map[string]interface{}{
		{
			"id":             1,
			"client_id":      "uuid-456",
			"name":           "Test List",
			"display_order":  0,
			"archived":       false,
			"created_at":     int64(1000),
			"updated_at":     int64(1001),
			"version":        1,
		},
	}

	result := convertListMaps(lists)

	if len(result) != 1 {
		t.Fatalf("expected 1 list, got %d", len(result))
	}

	list := result[0]
	if list.ID != 1 {
		t.Errorf("expected list id 1, got %d", list.ID)
	}
	if list.ClientID != "uuid-456" {
		t.Errorf("expected client_id 'uuid-456', got %s", list.ClientID)
	}
	if list.Name != "Test List" {
		t.Errorf("expected name 'Test List', got %s", list.Name)
	}
	if list.Archived {
		t.Error("expected archived to be false")
	}
	if list.Version != 1 {
		t.Errorf("expected version 1, got %d", list.Version)
	}
}

// TestConvertTaskToMap verifies task conversion from model to map
func TestConvertTaskToMap(t *testing.T) {
	task := models.Task{
		ID:         1,
		ClientID:   "uuid-123",
		TodoListID: 1,
		Todo:       "Test task",
		Priority:   3,
		Done:       false,
		DateAdded:  1000,
		Deleted:    false,
		CreatedAt:  1000,
		UpdatedAt:  1001,
		Version:    1,
	}

	result := convertTaskToMap(task)

	if result["client_id"] != "uuid-123" {
		t.Errorf("expected client_id 'uuid-123', got %v", result["client_id"])
	}
	if result["todo"] != "Test task" {
		t.Errorf("expected todo 'Test task', got %v", result["todo"])
	}
	if result["priority"] != 3 {
		t.Errorf("expected priority 3, got %v", result["priority"])
	}
	if result["updated_at"] != int64(1001) {
		t.Errorf("expected updated_at 1001, got %v", result["updated_at"])
	}
}

// TestConvertListToMap verifies list conversion from model to map
func TestConvertListToMap(t *testing.T) {
	list := models.TodoList{
		ID:           1,
		ClientID:     "uuid-456",
		Name:         "Test List",
		DisplayOrder: 0,
		Archived:     false,
		CreatedAt:    1000,
		UpdatedAt:    1001,
		Version:      1,
	}

	result := convertListToMap(list)

	if result["client_id"] != "uuid-456" {
		t.Errorf("expected client_id 'uuid-456', got %v", result["client_id"])
	}
	if result["name"] != "Test List" {
		t.Errorf("expected name 'Test List', got %v", result["name"])
	}
	if result["display_order"] != 0 {
		t.Errorf("expected display_order 0, got %v", result["display_order"])
	}
}

// TestConvertTaskMaps_OptionalTimestamps verifies that optional timestamps
// survive the pull conversion. GetTasksSince returns these as *time.Time,
// which convertTaskMaps must translate to *int64 unix timestamps.
// Regression test: these fields were previously dropped (asserted as *int64).
func TestConvertTaskMaps_OptionalTimestamps(t *testing.T) {
	completed := time.Unix(2000, 0)
	due := time.Unix(3000, 0)
	deleted := time.Unix(4000, 0)
	var unset *time.Time

	tasks := []map[string]interface{}{
		{
			"id":             1,
			"client_id":      "uuid-123",
			"todo_list_id":   1,
			"todo":           "Test task",
			"priority":       3,
			"done":           true,
			"deleted":        true,
			"date_added":     int64(1000),
			"date_completed": &completed,
			"due_date":       &due,
			"deleted_at":     &deleted,
			"created_at":     int64(1000),
			"updated_at":     int64(1001),
			"version":        1,
		},
		{
			"id":             2,
			"client_id":      "uuid-456",
			"todo_list_id":   1,
			"todo":           "No optional dates",
			"priority":       1,
			"done":           false,
			"deleted":        false,
			"date_added":     int64(1000),
			"date_completed": unset,
			"due_date":       unset,
			"deleted_at":     unset,
			"created_at":     int64(1000),
			"updated_at":     int64(1001),
			"version":        1,
		},
	}

	result := convertTaskMaps(tasks)
	if len(result) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result))
	}

	withDates := result[0]
	if withDates.DateCompleted == nil || *withDates.DateCompleted != 2000 {
		t.Errorf("expected date_completed 2000, got %v", withDates.DateCompleted)
	}
	if withDates.DueDate == nil || *withDates.DueDate != 3000 {
		t.Errorf("expected due_date 3000, got %v", withDates.DueDate)
	}
	if withDates.DeletedAt == nil || *withDates.DeletedAt != 4000 {
		t.Errorf("expected deleted_at 4000, got %v", withDates.DeletedAt)
	}

	withoutDates := result[1]
	if withoutDates.DateCompleted != nil {
		t.Errorf("expected nil date_completed, got %v", withoutDates.DateCompleted)
	}
	if withoutDates.DueDate != nil {
		t.Errorf("expected nil due_date, got %v", withoutDates.DueDate)
	}
	if withoutDates.DeletedAt != nil {
		t.Errorf("expected nil deleted_at, got %v", withoutDates.DeletedAt)
	}
}
