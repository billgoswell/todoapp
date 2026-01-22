package sync

import (
	"testing"
	"time"

	"github.com/billgoswell/commandlinetodo/shared-types/models"
)

func TestResolveTaskConflict_NewTask(t *testing.T) {
	serverTask := &models.Task{
		ClientID:   "task-1",
		Todo:       "New task",
		UpdatedAt:  time.Now().Unix(),
	}

	result := ResolveTaskConflict(serverTask, nil)
	if result != NoConflict {
		t.Errorf("expected NoConflict for new task, got %s", result)
	}
}

func TestResolveTaskConflict_ServerNewer(t *testing.T) {
	now := time.Now().Unix()

	serverTask := &models.Task{
		ClientID:   "task-1",
		Todo:       "Updated on server",
		UpdatedAt:  now + 100, // Server is 100 seconds newer
	}

	localTask := &models.Task{
		ClientID:   "task-1",
		Todo:       "Old local version",
		UpdatedAt:  now,
	}

	result := ResolveTaskConflict(serverTask, localTask)
	if result != ServerVersionWins {
		t.Errorf("expected ServerVersionWins when server is newer, got %s", result)
	}
}

func TestResolveTaskConflict_LocalNewer(t *testing.T) {
	now := time.Now().Unix()

	serverTask := &models.Task{
		ClientID:   "task-1",
		Todo:       "Stale server version",
		UpdatedAt:  now,
	}

	localTask := &models.Task{
		ClientID:   "task-1",
		Todo:       "Recent local changes",
		UpdatedAt:  now + 100, // Local is 100 seconds newer
	}

	result := ResolveTaskConflict(serverTask, localTask)
	if result != LocalVersionWins {
		t.Errorf("expected LocalVersionWins when local is newer, got %s", result)
	}
}

func TestResolveTaskConflict_SameTimestamp(t *testing.T) {
	now := time.Now().Unix()

	serverTask := &models.Task{
		ClientID:   "task-1",
		Todo:       "Server version",
		UpdatedAt:  now,
	}

	localTask := &models.Task{
		ClientID:   "task-1",
		Todo:       "Local version",
		UpdatedAt:  now, // Same timestamp
	}

	result := ResolveTaskConflict(serverTask, localTask)
	if result != LocalVersionWins {
		t.Errorf("expected LocalVersionWins when timestamps equal, got %s", result)
	}
}

func TestResolveListConflict_NewList(t *testing.T) {
	serverList := &models.TodoList{
		ClientID:   "list-1",
		Name:       "New List",
		UpdatedAt:  time.Now().Unix(),
	}

	result := ResolveListConflict(serverList, nil)
	if result != NoConflict {
		t.Errorf("expected NoConflict for new list, got %s", result)
	}
}

func TestResolveListConflict_ServerNewer(t *testing.T) {
	now := time.Now().Unix()

	serverList := &models.TodoList{
		ClientID:   "list-1",
		Name:       "Updated Name",
		UpdatedAt:  now + 100,
	}

	localList := &models.TodoList{
		ClientID:   "list-1",
		Name:       "Old Name",
		UpdatedAt:  now,
	}

	result := ResolveListConflict(serverList, localList)
	if result != ServerVersionWins {
		t.Errorf("expected ServerVersionWins when server is newer, got %s", result)
	}
}

func TestApplyTaskUpdate(t *testing.T) {
	serverTask := &models.Task{
		Done:       true,
		Todo:       "Updated description",
		Priority:   4,
		UpdatedAt:  time.Now().Unix(),
		Version:    5,
	}

	localTask := &models.Task{
		Done:       false,
		Todo:       "Old description",
		Priority:   2,
		DateAdded:  time.Now().Unix() - 86400, // Original creation time
		UpdatedAt:  time.Now().Unix() - 3600,
		Version:    3,
	}

	originalDateAdded := localTask.DateAdded
	ApplyTaskUpdate(serverTask, localTask)

	if localTask.Done != true {
		t.Error("Done field not updated")
	}
	if localTask.Todo != "Updated description" {
		t.Error("Todo field not updated")
	}
	if localTask.Priority != 4 {
		t.Error("Priority field not updated")
	}
	if localTask.DateAdded != originalDateAdded {
		t.Error("DateAdded was modified - should remain unchanged")
	}
	if localTask.Version != 5 {
		t.Error("Version not updated")
	}
}

func TestApplyListUpdate(t *testing.T) {
	serverList := &models.TodoList{
		Name:        "New Name",
		DisplayOrder: 5,
		Archived:    true,
		UpdatedAt:   time.Now().Unix(),
		Version:     3,
	}

	localList := &models.TodoList{
		Name:        "Old Name",
		DisplayOrder: 1,
		Archived:    false,
		CreatedAt:   time.Now().Unix() - 86400,
		UpdatedAt:   time.Now().Unix() - 3600,
		Version:     1,
	}

	originalCreatedAt := localList.CreatedAt
	ApplyListUpdate(serverList, localList)

	if localList.Name != "New Name" {
		t.Error("Name field not updated")
	}
	if localList.DisplayOrder != 5 {
		t.Error("DisplayOrder field not updated")
	}
	if localList.Archived != true {
		t.Error("Archived field not updated")
	}
	if localList.CreatedAt != originalCreatedAt {
		t.Error("CreatedAt was modified - should remain unchanged")
	}
	if localList.Version != 3 {
		t.Error("Version not updated")
	}
}
