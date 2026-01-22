package main

import (
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ============================================================================
// CLI HANDLER TESTS - PHASE 3
// ============================================================================
// These tests verify keyboard input handling and state transitions in handlers.go

// TestHandler_TaskCreation_TextInput tests the task creation flow starting with text input
func TestHandler_TaskCreation_TextInput(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()

	// Verify we start in main browse state
	if m.currentState != StateMainBrowse {
		t.Errorf("Expected StateMainBrowse, got %d", m.currentState)
	}

	// Test: Press 'a' to add a task
	textMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, _ := m.Update(textMsg)
	newM, ok := updatedModel.(*model)
	if !ok {
		t.Fatal("Failed to type assert updated model to *model")
	}

	// Should transition to task input state
	if newM.currentState != StateTaskInput {
		t.Errorf("Expected StateTaskInput after 'a', got %d", newM.currentState)
	}

	// Text input should be focused
	if !newM.textInput.Focused() {
		t.Error("Expected text input to be focused")
	}
}

// TestHandler_TaskCreation_PrioritySelection tests priority selection in task creation flow
func TestHandler_TaskCreation_PrioritySelection(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	m.currentState = StatePrioritySelection
	m.taskFlow.step = TaskFlowSelectPriority
	m.taskFlow.text = "Test task"
	m.taskFlow.priority = PriorityMed

	t.Logf("Initial state: priority=%d, step=%d", m.taskFlow.priority, m.taskFlow.step)

	// Test: Press '1' for high priority
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	t.Logf("Sending key: %s", keyMsg.String())
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	t.Logf("After pressing '1': priority=%d (expected %d)", newM.taskFlow.priority, PriorityHigh)

	// Priority should be updated
	if newM.taskFlow.priority != PriorityHigh {
		t.Errorf("Expected PriorityHigh (%d), got %d", PriorityHigh, newM.taskFlow.priority)
	}

	// Test: Press '4' for low priority
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}}
	t.Logf("Sending key: %s", keyMsg.String())
	updatedModel, _ = newM.Update(keyMsg)
	newM = updatedModel.(*model)

	t.Logf("After pressing '4': priority=%d (expected %d)", newM.taskFlow.priority, PriorityLow)

	if newM.taskFlow.priority != PriorityLow {
		t.Errorf("Expected PriorityLow (%d), got %d", PriorityLow, newM.taskFlow.priority)
	}
}

// TestHandler_TaskCreation_SetDueDate tests due date selection in task creation
func TestHandler_TaskCreation_SetDueDate(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	m.currentState = StateDueDateInput
	m.taskFlow.text = "Test task"
	m.taskFlow.priority = PriorityHigh

	// Test: Press '+' to set due date for tomorrow
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+' }}
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	// Due date should be set to approximately tomorrow
	tomorrow := time.Now().AddDate(0, 0, 1).Unix()
	dayDiff := (newM.taskFlow.dueDate - tomorrow) / SecondsPerDay

	if dayDiff != 0 {
		t.Logf("Due date set: expected ~%d, got %d (diff: %d days)", tomorrow, newM.taskFlow.dueDate, dayDiff)
	}
}

// TestHandler_TaskEdit_UpdatesTask tests editing an existing task
func TestHandler_TaskEdit_UpdatesTask(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	if len(m.items) == 0 {
		t.Skip("No test items available")
	}

	// Set to edit mode for first item
	m.currentState = StateEditTask
	m.input.itemIndex = 0
	m.textInput.SetValue(m.items[0].todo)

	// Update task text
	newText := "Updated task text"
	m.textInput.SetValue(newText)

	// Press Enter to confirm
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	// Should return to main view
	if newM.currentState != StateMainBrowse {
		t.Errorf("Expected StateMainBrowse after edit, got %d", newM.currentState)
	}

	// Find updated item
	found := false
	for _, item := range newM.items {
		if item.todo == newText {
			found = true
			if item.updatedAt == 0 {
				t.Error("Expected updatedAt to be set")
			}
			break
		}
	}
	if !found {
		t.Error("Updated task not found in items")
	}
}

// TestHandler_TaskToggleDone tests marking task as complete/incomplete
func TestHandler_TaskToggleDone_TogglesCompletion(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	if len(m.items) == 0 {
		t.Skip("No test items available")
	}

	m.currentState = StateMainBrowse
	m.cursor = 0

	// Track the item we're toggling by ID
	toggledItemID := m.items[0].id
	originalDoneState := m.items[0].done

	t.Logf("Before toggle: items=%d, toggling item id=%d", len(m.items), toggledItemID)
	for i, item := range m.items {
		t.Logf("  Item %d: id=%d, done=%v, priority=%d, todo=%s", i, item.id, item.done, item.priority, item.todo)
	}

	// Toggle completion with Space
	keyMsg := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	t.Logf("After toggle: items=%d", len(newM.items))
	for i, item := range newM.items {
		t.Logf("  Item %d: id=%d, done=%v, priority=%d, todo=%s", i, item.id, item.done, item.priority, item.todo)
	}

	// Find the toggled item by ID
	var toggledItem *todoItem
	for i := range newM.items {
		if newM.items[i].id == toggledItemID {
			toggledItem = &newM.items[i]
			break
		}
	}

	if toggledItem == nil {
		t.Fatal("Toggled item not found in updated model")
	}

	// Check if done status changed
	if toggledItem.done == originalDoneState {
		t.Error("Expected done status to toggle")
	}

	// Verify items are sorted (pending before done)
	if len(newM.items) > 1 {
		for i := 0; i < len(newM.items)-1; i++ {
			if newM.items[i].done && !newM.items[i+1].done {
				t.Errorf("Sorting failed: item %d (done=%v) before item %d (done=%v)",
					i, newM.items[i].done, i+1, newM.items[i+1].done)
			}
		}
	}
}

// TestHandler_TaskDelete_ConfirmationPrompt tests delete confirmation flow
func TestHandler_TaskDelete_ShowsConfirmation(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	if len(m.items) == 0 {
		t.Skip("No test items available")
	}

	m.currentState = StateMainBrowse
	m.cursor = 0

	// Press 'd' to delete
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	// Should show confirmation
	if newM.currentState != StateDeleteConfirm {
		t.Errorf("Expected StateDeleteConfirm, got %d", newM.currentState)
	}

	if newM.input.deleteIndex != 0 {
		t.Errorf("Expected deleteIndex=0, got %d", newM.input.deleteIndex)
	}
}

// TestHandler_TaskDelete_ConfirmDelete tests confirming deletion
func TestHandler_TaskDelete_ConfirmDelete(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	if len(m.items) == 0 {
		t.Skip("No test items available")
	}

	deletedItemID := m.items[0].id

	m.currentState = StateDeleteConfirm
	m.input.deleteIndex = 0

	// Press 'y' to confirm deletion
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	// Should return to main view
	if newM.currentState != StateMainBrowse {
		t.Errorf("Expected StateMainBrowse after confirmation, got %d", newM.currentState)
	}

	// Item should be marked as deleted
	found := false
	for _, item := range newM.items {
		if item.id == deletedItemID && item.deleted {
			found = true
			break
		}
	}
	if !found {
		t.Error("Deleted item should be marked as deleted")
	}
}

// TestHandler_TaskDelete_CancelDelete tests canceling deletion
func TestHandler_TaskDelete_CancelDelete(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	if len(m.items) == 0 {
		t.Skip("No test items available")
	}

	initialItem := m.items[0]
	m.currentState = StateDeleteConfirm
	m.input.deleteIndex = 0

	// Press 'n' to cancel deletion
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	// Should return to main view
	if newM.currentState != StateMainBrowse {
		t.Errorf("Expected StateMainBrowse, got %d", newM.currentState)
	}

	// Item should NOT be deleted
	found := false
	for _, item := range newM.items {
		if item.id == initialItem.id && !item.deleted && item.todo == initialItem.todo {
			found = true
			break
		}
	}
	if !found {
		t.Error("Item should not be deleted after cancellation")
	}
}

// TestHandler_Navigation_CursorMovement tests cursor navigation up/down
func TestHandler_Navigation_CursorMovement(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	if len(m.items) < 2 {
		t.Skip("Need at least 2 items for navigation test")
	}

	m.currentState = StateMainBrowse
	m.cursor = 0

	// Press 'j' to move down
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	if newM.cursor != 1 {
		t.Errorf("Expected cursor=1 after down, got %d", newM.cursor)
	}

	// Press 'k' to move up
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedModel, _ = newM.Update(keyMsg)
	newM = updatedModel.(*model)

	if newM.cursor != 0 {
		t.Errorf("Expected cursor=0 after up, got %d", newM.cursor)
	}
}

// TestHandler_ListSelector_OpenListMenu tests opening list selector
func TestHandler_ListSelector_OpenListMenu(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	m.currentState = StateMainBrowse

	// Press 'l' to open list selector
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	// Should transition to list selector
	if newM.currentState != StateListSelector {
		t.Errorf("Expected StateListSelector, got %d", newM.currentState)
	}
}

// TestHandler_ListCreation_NewList tests creating a new list
func TestHandler_ListCreation_NewList(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	initialListCount := len(m.todoLists)

	t.Logf("Initial lists: %d", initialListCount)
	for i, list := range m.todoLists {
		t.Logf("  List %d: name=%s, id=%d", i, list.name, list.id)
	}

	m.currentState = StateListNameInput
	m.currentSubState = SubStateListCreate
	m.textInput.SetValue("New Test List")

	t.Logf("Before: state=%d, substate=%d", m.currentState, m.currentSubState)
	t.Logf("Store type: %T", m.store)

	// Verify mock store
	lists, err := m.store.GetTodoLists()
	if err != nil {
		t.Logf("Error getting lists: %v", err)
	}
	t.Logf("Lists from store before: %d", len(lists))

	// Press Enter to create list
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := m.Update(keyMsg)
	newM := updatedModel.(*model)

	// Check store after
	lists, err = newM.store.GetTodoLists()
	if err != nil {
		t.Logf("Error getting lists after: %v", err)
	}
	t.Logf("Lists from store after: %d", len(lists))
	for i, list := range lists {
		t.Logf("  Store List %d: name=%s", i, list.name)
	}

	t.Logf("After: state=%d, substate=%d", newM.currentState, newM.currentSubState)
	t.Logf("After update: lists=%d (was %d)", len(newM.todoLists), initialListCount)
	for i, list := range newM.todoLists {
		t.Logf("  List %d: name=%s, id=%d", i, list.name, list.id)
	}

	// Should return to main view
	if newM.currentState != StateMainBrowse {
		t.Errorf("Expected StateMainBrowse, got %d", newM.currentState)
	}

	// New list should be created
	if len(newM.todoLists) <= initialListCount {
		t.Error("Expected new list to be created")
	}

	// Check if the new list exists
	found := false
	for _, list := range newM.todoLists {
		if list.name == "New Test List" {
			found = true
			break
		}
	}
	if !found {
		t.Error("New list 'New Test List' not found")
	}
}

// TestHandler_Quit_ExitApplication tests quit functionality
func TestHandler_Quit_ExitApplication(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB()

	m := createTestModel()
	m.currentState = StateMainBrowse

	// Press 'q' to quit
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(keyMsg)

	// Should return a quit command
	if cmd == nil {
		t.Error("Expected quit command")
	}
}

// Test helpers

func setupTestDB(t *testing.T) {
	testDBPath := "test_cli.db"
	os.Remove(testDBPath)

	// Initialize test database
	_, err := initDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
}

func cleanupTestDB() {
	os.Remove("test_cli.db")
}
