package main

import (
	"encoding/json"
	"testing"
)

// TestTaskPayloadWithListClientID verifies TaskPayload includes TodoListClientID
func TestTaskPayloadWithListClientID(t *testing.T) {
	task := TaskPayload{
		ClientID:         "task-001",
		Todo:             "Buy milk",
		Priority:         2,
		Done:             false,
		DateAdded:        1704067200,
		DateCompleted:    0,
		DueDate:          0,
		Deleted:          false,
		DeletedAt:        0,
		TodoListID:       1,
		TodoListClientID: "list-uuid-001",
		UpdatedAt:        1704067200,
		Version:          1,
	}

	// Verify fields are present
	if task.TodoListClientID != "list-uuid-001" {
		t.Fatalf("TodoListClientID not set correctly: expected 'list-uuid-001', got '%s'", task.TodoListClientID)
	}

	t.Log("✓ TaskPayload has TodoListClientID field")

	// Verify JSON marshaling includes the field
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal task: %v", err)
	}

	var taskMap map[string]interface{}
	if err := json.Unmarshal(data, &taskMap); err != nil {
		t.Fatalf("Failed to unmarshal task: %v", err)
	}

	if _, hasField := taskMap["todo_list_client_id"]; !hasField {
		t.Fatal("todo_list_client_id not in JSON output")
	}

	t.Logf("✓ JSON marshaling includes todo_list_client_id: %v", taskMap["todo_list_client_id"])

	// Verify unmarshaling works
	jsonData := []byte(`{
		"client_id": "task-002",
		"todo": "Pay bills",
		"priority": 3,
		"done": false,
		"date_added": 1704067200,
		"date_completed": 0,
		"due_date": 0,
		"deleted": false,
		"deleted_at": 0,
		"todo_list_id": 2,
		"todo_list_client_id": "list-uuid-002",
		"updated_at": 1704067200,
		"version": 1
	}`)

	var taskFromJSON TaskPayload
	if err := json.Unmarshal(jsonData, &taskFromJSON); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if taskFromJSON.TodoListClientID != "list-uuid-002" {
		t.Fatalf("TodoListClientID not deserialized correctly: expected 'list-uuid-002', got '%s'", taskFromJSON.TodoListClientID)
	}

	t.Logf("✓ JSON deserialization works: TodoListClientID='%s'", taskFromJSON.TodoListClientID)
	t.Log("✅ All TaskPayload sync tests passed!")
}
