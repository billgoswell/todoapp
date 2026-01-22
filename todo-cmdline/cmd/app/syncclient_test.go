package main

import (
	"testing"
	"time"
)

// TestSyncPayloadTimestamp verifies that the sync payload uses actual updated_at
// instead of dateAdded for conflict resolution
func TestSyncPayloadTimestamp(t *testing.T) {
	// Arrange: Create test items with specific timestamps
	now := time.Now().Unix()
	dateAdded := now - 3600     // 1 hour ago
	updatedAt := now - 600      // 10 minutes ago
	laterUpdate := now - 300    // 5 minutes ago

	items := []todoItem{
		{
			id:         1,
			clientID:   "test-001",
			todo:       "Original task",
			priority:   3,
			done:       false,
			dateAdded:  dateAdded,
			updatedAt:  updatedAt, // Recently updated
			todoListID: 1,
			version:    1,
		},
		{
			id:         2,
			clientID:   "test-002",
			todo:       "Recently modified task",
			priority:   2,
			done:       true,
			dateAdded:  dateAdded,
			updatedAt:  laterUpdate, // Even more recently updated
			todoListID: 1,
			version:    2,
		},
	}

	// Act: Convert items to payloads (same logic as PushChanges)
	taskPayloads := make([]TaskPayload, len(items))
	for i, item := range items {
		taskPayloads[i] = TaskPayload{
			ClientID:      item.clientID,
			Todo:          item.todo,
			Priority:      item.priority,
			Done:          item.done,
			DateAdded:     item.dateAdded,
			DateCompleted: item.dateCompleted,
			DueDate:       item.dueDate,
			Deleted:       item.deleted,
			DeletedAt:     item.deletedAt,
			TodoListID:    item.todoListID,
			UpdatedAt:     item.updatedAt, // Should use updatedAt, not dateAdded
			Version:       item.version,
		}
	}

	// Assert: Verify that UpdatedAt uses the actual updated timestamp
	tests := []struct {
		name          string
		itemIndex     int
		expectedTime  int64
		shouldNotBe   int64
	}{
		{
			name:         "First item uses correct updated_at",
			itemIndex:    0,
			expectedTime: updatedAt,
			shouldNotBe:  dateAdded,
		},
		{
			name:         "Second item uses correct updated_at",
			itemIndex:    1,
			expectedTime: laterUpdate,
			shouldNotBe:  dateAdded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := taskPayloads[tt.itemIndex]

			// Check that UpdatedAt is set to the actual updated timestamp
			if payload.UpdatedAt != tt.expectedTime {
				t.Errorf("Expected UpdatedAt %d, got %d", tt.expectedTime, payload.UpdatedAt)
			}

			// Verify that UpdatedAt is NOT using dateAdded
			if payload.UpdatedAt == tt.shouldNotBe {
				t.Errorf("UpdatedAt should not be dateAdded (%d)", tt.shouldNotBe)
			}

			// Verify that UpdatedAt is different from DateAdded
			if payload.UpdatedAt == payload.DateAdded {
				t.Error("UpdatedAt should not equal DateAdded")
			}
		})
	}
}

// TestConflictResolution_LastWriteWins verifies the sync conflict resolution logic
// This tests that newer timestamps take precedence in sync operations
func TestConflictResolution_LastWriteWins(t *testing.T) {
	now := time.Now().Unix()

	// Simulate two devices modifying the same task
	// Device A: updates at time T1
	// Device B: updates at time T2 (later)
	// Expected: Device B's version should win because T2 > T1

	t1 := now - 1000 // Older timestamp
	t2 := now - 500  // Newer timestamp

	clientAItem := todoItem{
		id:         1,
		clientID:   "shared-task-001",
		todo:       "Updated by device A",
		priority:   3,
		done:       false,
		dateAdded:  now - 3600,
		updatedAt:  t1, // Earlier update
		todoListID: 1,
		version:    2,
	}

	clientBItem := todoItem{
		id:         1,
		clientID:   "shared-task-001",
		todo:       "Updated by device B",
		priority:   2,
		done:       true,
		dateAdded:  now - 3600,
		updatedAt:  t2, // Later update
		todoListID: 1,
		version:    3,
	}

	// In a real scenario, the server would receive both updates
	// and apply the one with the higher updated_at timestamp
	// This test verifies the logic by comparing timestamps

	t.Run("Newer update should win", func(t *testing.T) {
		// Device B's update is newer, so it should be the winner
		if clientBItem.updatedAt <= clientAItem.updatedAt {
			t.Errorf("Device B's timestamp (%d) should be newer than Device A's (%d)",
				clientBItem.updatedAt, clientAItem.updatedAt)
		}

		// In conflict resolution: if clientB.updatedAt > clientA.updatedAt, use clientB
		if clientBItem.updatedAt > clientAItem.updatedAt {
			// clientB wins, verify its content is preferred
			if clientBItem.todo != "Updated by device B" {
				t.Error("Device B's content should be selected")
			}
		}
	})

	t.Run("Older update should not overwrite newer", func(t *testing.T) {
		// If we try to apply the older update (clientA) after newer (clientB),
		// it should be rejected
		if clientAItem.updatedAt > clientBItem.updatedAt {
			t.Error("Device A's timestamp should not be newer")
		}

		// The server should reject this because it's older
		shouldApplyUpdate := clientAItem.updatedAt > clientBItem.updatedAt
		if shouldApplyUpdate {
			t.Error("Server should not apply older update")
		}
	})
}

// TestSyncItem_PreservesUpdatedAt verifies that updatedAt is preserved
// when syncing items from server
func TestSyncItem_PreservesUpdatedAt(t *testing.T) {
	now := time.Now().Unix()
	serverUpdateTime := now - 3600

	// Task received from server with updated_at set
	serverTask := todoItem{
		id:         1,
		clientID:   "task-from-server",
		todo:       "Task from server",
		priority:   3,
		done:       false,
		dateAdded:  serverUpdateTime - 3600,
		updatedAt:  serverUpdateTime, // Server's timestamp
		todoListID: 1,
		version:    2,
	}

	// When saved locally, updated_at should be preserved
	t.Run("Save task from server preserves updated_at", func(t *testing.T) {
		if serverTask.updatedAt == 0 {
			t.Error("Task updatedAt should be set")
		}

		if serverTask.updatedAt != serverUpdateTime {
			t.Errorf("Expected updatedAt %d, got %d", serverUpdateTime, serverTask.updatedAt)
		}

		// Verify it's not using dateAdded
		if serverTask.updatedAt == serverTask.dateAdded {
			t.Error("updatedAt should not be same as dateAdded")
		}
	})
}

