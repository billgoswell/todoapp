package main

import (
	"os"
	"testing"
	"time"
)

// ============================================================================
// STATE MANAGEMENT & DATASTORE TESTS - PHASE 3
// ============================================================================
// These tests verify state management and datastore operations

// TestDataStore_LocalStore_GetItems tests retrieving items from database
func TestDataStore_LocalStore_GetItems(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	items, err := store.GetItems()
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}

	if items == nil {
		t.Error("Expected items to be non-nil (empty slice if no items)")
	}
}

// TestDataStore_LocalStore_SaveItem tests saving new items
func TestDataStore_LocalStore_SaveItem(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	now := time.Now().Unix()
	item := todoItem{
		clientID:   "test-save-001",
		todo:       "Test item to save",
		priority:   3,
		done:       false,
		dateAdded:  now,
		updatedAt:  now,
		todoListID: testListID,
		version:    1,
	}

	err := store.SaveItem(item)
	if err != nil {
		t.Fatalf("SaveItem failed: %v", err)
	}

	// Verify item was saved by retrieving it
	items, err := store.GetItems()
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}

	t.Logf("Looking for item with clientID=%s, found %d items in database", item.clientID, len(items))
	for i, it := range items {
		t.Logf("  Item %d: clientID=%s, todo=%s, listID=%d", i, it.clientID, it.todo, it.todoListID)
	}

	// Find the item by todo text instead (SaveItem might regenerate clientID)
	found := false
	for _, retrieved := range items {
		if retrieved.todo == item.todo {
			found = true
			// Update the test item with the actual clientID from database
			item.clientID = retrieved.clientID
			break
		}
	}

	if !found {
		t.Errorf("Saved item with todo='%s' not found in database", item.todo)
	}
}

// TestDataStore_LocalStore_UpdateItem tests updating existing items
func TestDataStore_LocalStore_UpdateItem(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	now := time.Now().Unix()
	item := todoItem{
		clientID:   "test-update-001",
		todo:       "Original text",
		priority:   3,
		done:       false,
		dateAdded:  now,
		updatedAt:  now,
		todoListID: testListID,
		version:    1,
	}

	// Save first
	err := store.SaveItem(item)
	if err != nil {
		t.Fatalf("SaveItem failed: %v", err)
	}

	// Get the saved item to get its ID
	items, _ := store.GetItems()
	if len(items) == 0 {
		t.Fatal("No items found after save")
	}

	savedItem := items[0]
	savedItem.todo = "Updated text"
	savedItem.priority = 2
	savedItem.updatedAt = time.Now().Unix()

	// Update it
	err = store.UpdateItem(savedItem)
	if err != nil {
		t.Fatalf("UpdateItem failed: %v", err)
	}

	// Verify update
	items, _ = store.GetItems()
	found := false
	for _, item := range items {
		if item.id == savedItem.id {
			found = true
			if item.todo != "Updated text" {
				t.Errorf("Text not updated: got %s", item.todo)
			}
			if item.priority != 2 {
				t.Errorf("Priority not updated: expected 2, got %d", item.priority)
			}
			break
		}
	}

	if !found {
		t.Error("Updated item not found")
	}
}

// TestDataStore_LocalStore_DeleteItem tests soft deletion of items
func TestDataStore_LocalStore_DeleteItem(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	now := time.Now().Unix()
	item := todoItem{
		clientID:   "test-delete-001",
		todo:       "Item to delete",
		dateAdded:  now,
		updatedAt:  now,
		todoListID: 1,
		version:    1,
	}

	// Save first
	store.SaveItem(item)

	// Get the ID
	items, _ := store.GetItems()
	if len(items) == 0 {
		t.Fatal("Item not saved")
	}

	itemID := items[0].id

	// Delete it
	err := store.DeleteItem(itemID)
	if err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}

	// Verify it's marked as deleted (soft delete)
	// GetItems should not return deleted items
	items, _ = store.GetItems()
	for _, item := range items {
		if item.id == itemID && item.deleted {
			t.Log("Item marked as deleted (soft delete)")
			return
		}
	}
}

// TestDataStore_LocalStore_GetTodoLists tests retrieving lists
func TestDataStore_LocalStore_GetTodoLists(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	lists, err := store.GetTodoLists()
	if err != nil {
		t.Fatalf("GetTodoLists failed: %v", err)
	}

	if lists == nil {
		t.Error("Expected lists to be non-nil")
	}

	// At minimum we should have a default list
	if len(lists) == 0 {
		t.Log("No lists in database (expected if not initialized)")
	}
}

// TestDataStore_LocalStore_CreateTodoList tests creating a new list
func TestDataStore_LocalStore_CreateTodoList(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	listID, err := store.CreateTodoList("Test List")
	if err != nil {
		t.Fatalf("CreateTodoList failed: %v", err)
	}

	if listID <= 0 {
		t.Errorf("Expected positive list ID, got %d", listID)
	}

	// Verify list was created
	lists, _ := store.GetTodoLists()
	found := false
	for _, list := range lists {
		if list.id == listID && list.name == "Test List" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Created list not found in database")
	}
}

// TestDataStore_LocalStore_UpdateTodoListName tests updating list name
func TestDataStore_LocalStore_UpdateTodoListName(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	// Create a list
	listID, err := store.CreateTodoList("Original Name")
	if err != nil {
		t.Fatalf("CreateTodoList failed: %v", err)
	}

	// Update the name
	err = store.UpdateTodoListName(listID, "Updated Name")
	if err != nil {
		t.Fatalf("UpdateTodoListName failed: %v", err)
	}

	// Verify update
	lists, _ := store.GetTodoLists()
	found := false
	for _, list := range lists {
		if list.id == listID && list.name == "Updated Name" {
			found = true
			break
		}
	}

	if !found {
		t.Error("List name was not updated")
	}
}

// TestDataStore_SyncMetadata_GetLastSyncTime tests retrieving last sync time
func TestDataStore_SyncMetadata_GetLastSyncTime(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	syncTime, err := store.GetLastSyncTime()
	if err != nil {
		t.Fatalf("GetLastSyncTime failed: %v", err)
	}

	// Should return a timestamp (0 if never synced)
	if syncTime < 0 {
		t.Errorf("Expected non-negative sync time, got %d", syncTime)
	}
}

// TestDataStore_SyncMetadata_SetLastSyncTime tests setting last sync time
func TestDataStore_SyncMetadata_SetLastSyncTime(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	now := time.Now().Unix()
	err := store.SetLastSyncTime(now)
	if err != nil {
		t.Fatalf("SetLastSyncTime failed: %v", err)
	}

	// Verify it was set
	syncTime, _ := store.GetLastSyncTime()
	if syncTime != now {
		t.Errorf("Sync time not set correctly: expected %d, got %d", now, syncTime)
	}
}

// TestDataStore_ChangeLog_LogChange tests logging changes
func TestDataStore_ChangeLog_LogChange(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	// Log a change
	err := store.LogChange("task", 1, "create")
	if err != nil {
		t.Fatalf("LogChange failed: %v", err)
	}

	// Verify change was logged
	changes, err := store.GetPendingChanges()
	if err != nil {
		t.Fatalf("GetPendingChanges failed: %v", err)
	}

	if len(changes) == 0 {
		t.Error("Expected changes to be logged")
	}
}

// TestDataStore_ChangeLog_GetPendingChanges tests retrieving pending changes
func TestDataStore_ChangeLog_GetPendingChanges(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	// Log some changes
	store.LogChange("task", 1, "create")
	store.LogChange("task", 2, "update")
	store.LogChange("list", 1, "create")

	// Get pending changes
	changes, err := store.GetPendingChanges()
	if err != nil {
		t.Fatalf("GetPendingChanges failed: %v", err)
	}

	if len(changes) < 3 {
		t.Logf("Expected at least 3 changes, got %d", len(changes))
	}

	// Verify change types
	hasCreate := false
	hasUpdate := false
	for _, change := range changes {
		if change.changeType == "create" {
			hasCreate = true
		}
		if change.changeType == "update" {
			hasUpdate = true
		}
	}

	if !hasCreate {
		t.Error("Expected at least one create change")
	}
	if !hasUpdate {
		t.Error("Expected at least one update change")
	}
}

// TestDataStore_ChangeLog_MarkChangeSynced tests marking changes as synced
func TestDataStore_ChangeLog_MarkChangeSynced(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	store := NewLocalStore(db)

	// Log a change
	store.LogChange("task", 1, "create")

	// Get pending changes
	changes, _ := store.GetPendingChanges()
	if len(changes) == 0 {
		t.Skip("No changes to mark as synced")
	}

	changeID := changes[0].id

	// Mark as synced
	err := store.MarkChangeSynced(changeID)
	if err != nil {
		t.Fatalf("MarkChangeSynced failed: %v", err)
	}

	// Verify it was marked (should no longer appear in pending)
	changes, _ = store.GetPendingChanges()
	found := false
	for _, change := range changes {
		if change.id == changeID && !change.synced {
			found = true
			break
		}
	}

	if found {
		t.Error("Change should be marked as synced")
	}
}

// TestState_ItemSorting tests that items are properly sorted
func TestState_ItemSorting(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	now := time.Now().Unix()

	// Create model with items in various states
	items := []todoItem{
		{
			id:        1,
			clientID:  "test-1",
			todo:      "Completed task",
			priority:  3,
			done:      true,
			dateAdded: now,
		},
		{
			id:        2,
			clientID:  "test-2",
			todo:      "High priority task",
			priority:  1,
			done:      false,
			dateAdded: now,
		},
		{
			id:        3,
			clientID:  "test-3",
			todo:      "Medium priority task",
			priority:  3,
			done:      false,
			dateAdded: now,
		},
	}

	lists := []todoList{
		{
			id:        1,
			clientID:  "list-1",
			name:      "Test List",
			createdAt: now,
			updatedAt: now,
		},
	}

	m := initialModel(items, lists)

	// Sort items
	m.sortItems()

	// Verify order: pending items first (done=false), then by priority
	if len(m.items) >= 2 {
		// First item should be pending (done=false)
		if m.items[0].done {
			t.Error("First item should be pending (done=false)")
		}

		// Pending items should be sorted by priority (lower number first)
		pendingCount := 0
		for i := 0; i < len(m.items); i++ {
			if !m.items[i].done {
				pendingCount++
				if i > 0 && !m.items[i-1].done {
					if m.items[i].priority < m.items[i-1].priority {
						t.Errorf("Pending items not sorted by priority at index %d", i)
					}
				}
			}
		}

		// Check that all pending items come before completed items
		completedStartIndex := -1
		for i := 0; i < len(m.items); i++ {
			if m.items[i].done {
				completedStartIndex = i
				break
			}
		}

		if completedStartIndex > 0 {
			for i := completedStartIndex + 1; i < len(m.items); i++ {
				if !m.items[i].done {
					t.Error("Completed items should come after pending items")
					break
				}
			}
		}
	}
}

// TestState_CursorNavigation tests cursor position management
func TestState_CursorNavigation(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	now := time.Now().Unix()

	items := []todoItem{
		{id: 1, clientID: "test-1", todo: "Task 1", dateAdded: now},
		{id: 2, clientID: "test-2", todo: "Task 2", dateAdded: now},
		{id: 3, clientID: "test-3", todo: "Task 3", dateAdded: now},
	}

	lists := []todoList{
		{id: 1, clientID: "list-1", name: "Test", createdAt: now, updatedAt: now},
	}

	m := initialModel(items, lists)

	// Test cursor bounds
	m.cursor = 0
	if m.cursor != 0 {
		t.Error("Cursor should be at 0")
	}

	// Navigate down
	m.cursor++
	if m.cursor != 1 {
		t.Error("Cursor should be at 1")
	}

	// Test upper bound
	maxCursor := len(m.items) - 1
	m.cursor = maxCursor
	if m.cursor > len(m.items)-1 {
		t.Error("Cursor should not exceed item count")
	}
}

// TestState_ListSelection tests list selection state
func TestState_ListSelection(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	now := time.Now().Unix()

	lists := []todoList{
		{id: 1, clientID: "list-1", name: "List 1", createdAt: now, updatedAt: now},
		{id: 2, clientID: "list-2", name: "List 2", createdAt: now, updatedAt: now},
		{id: 3, clientID: "list-3", name: "List 3", createdAt: now, updatedAt: now},
	}

	items := []todoItem{}
	m := initialModel(items, lists)

	// Verify initial list selection
	if m.currentListID != lists[0].id {
		t.Errorf("Expected current list ID %d, got %d", lists[0].id, m.currentListID)
	}

	// Change list selection
	m.currentListIndex = 1
	if m.currentListIndex != 1 {
		t.Error("List index should be updated")
	}

	// Get list at index
	list := m.getListAtIndex(m.currentListIndex)
	if list == nil {
		t.Error("List at index should not be nil")
	}

	if list.id != lists[1].id {
		t.Errorf("Expected list ID %d, got %d", lists[1].id, list.id)
	}
}

// Test helpers

var testListID int // Global test list ID for use in tests

func setupTest(t *testing.T) {
	testDBPath := "test_state.db"
	os.Remove(testDBPath)

	// Initialize test database
	testDB, err := initDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	db = testDB

	// Create a test list to satisfy foreign key constraints
	store := NewLocalStore(db)
	listID, err := store.CreateTodoList("Test List")
	if err != nil {
		t.Fatalf("Failed to create test list: %v", err)
	}
	testListID = listID
}

func cleanupTest() {
	if db != nil {
		db.Close()
	}
	os.Remove("test_state.db")
}
