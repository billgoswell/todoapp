package main

import (
	"fmt"
	"sync"
	"time"

	cfg "github.com/billgoswell/commandlinetodo/internal/config"
	"github.com/billgoswell/commandlinetodo/internal/logger"
)

// SyncStore wraps LocalStore with synchronization capabilities
type SyncStore struct {
	local   *LocalStore
	client  *SyncClient
	config  cfg.SyncConfig
	mu      sync.RWMutex
	stopCh  chan struct{}
	running bool
}

// NewSyncStore creates a new sync store instance
func NewSyncStore(local *LocalStore, client *SyncClient, config cfg.SyncConfig) *SyncStore {
	return &SyncStore{
		local:   local,
		client:  client,
		config:  config,
		stopCh:  make(chan struct{}),
		running: false,
	}
}

// GetTodoLists retrieves all todo lists
func (s *SyncStore) GetTodoLists() ([]todoList, error) {
	return s.local.GetTodoLists()
}

// GetTodoListByID retrieves a list by its ID
func (s *SyncStore) GetTodoListByID(id int) (todoList, error) {
	return s.local.GetTodoListByID(id)
}

// GetTodoListByClientID retrieves a list by its client ID
func (s *SyncStore) GetTodoListByClientID(clientID string) (todoList, error) {
	return s.local.GetTodoListByClientID(clientID)
}

// CreateTodoList creates a new todo list
func (s *SyncStore) CreateTodoList(name string) (int, error) {
	id, err := s.local.CreateTodoList(name)
	if err != nil {
		return 0, err
	}

	s.local.LogChange("list", id, "create")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return id, nil
}

// UpdateTodoListName updates a todo list name
func (s *SyncStore) UpdateTodoListName(id int, name string) error {
	err := s.local.UpdateTodoListName(id, name)
	if err != nil {
		return err
	}

	s.local.LogChange("list", id, "update")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// DeleteTodoList deletes a todo list
func (s *SyncStore) DeleteTodoList(id int) error {
	err := s.local.DeleteTodoList(id)
	if err != nil {
		return err
	}

	s.local.LogChange("list", id, "delete")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// ArchiveTodoList archives a todo list
func (s *SyncStore) ArchiveTodoList(id int) error {
	err := s.local.ArchiveTodoList(id)
	if err != nil {
		return err
	}

	s.local.LogChange("list", id, "update")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// UnarchiveTodoList unarchives a todo list
func (s *SyncStore) UnarchiveTodoList(id int) error {
	err := s.local.UnarchiveTodoList(id)
	if err != nil {
		return err
	}

	s.local.LogChange("list", id, "update")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// SaveTodoList saves a new list from the server (doesn't trigger sync)
func (s *SyncStore) SaveTodoList(list todoList) error {
	return s.local.SaveTodoList(list)
}

// UpdateTodoListFromServer updates a list from the server (doesn't trigger sync)
func (s *SyncStore) UpdateTodoListFromServer(list todoList) error {
	return s.local.UpdateTodoListFromServer(list)
}

// UpdateListServerID updates a list's server ID after successful sync
func (s *SyncStore) UpdateListServerID(clientID string, serverID int) error {
	return s.local.UpdateListServerID(clientID, serverID)
}

// GetItems retrieves all items
func (s *SyncStore) GetItems() ([]todoItem, error) {
	return s.local.GetItems()
}

// GetItemByID retrieves an item by ID
func (s *SyncStore) GetItemByID(id int) (todoItem, error) {
	return s.local.GetItemByID(id)
}

// GetItemByClientID retrieves an item by client ID
func (s *SyncStore) GetItemByClientID(clientID string) (todoItem, error) {
	return s.local.GetItemByClientID(clientID)
}

// SaveItem saves a new item
func (s *SyncStore) SaveItem(item todoItem) error {
	if item.clientID == "" {
		item.clientID = generateClientID()
	}

	err := s.local.SaveItem(item)
	if err != nil {
		return err
	}

	// Get the item back to get the ID
	savedItem, err := s.local.GetItemByClientID(item.clientID)
	if err == nil {
		s.local.LogChange("task", savedItem.id, "create")
	}

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// UpdateItem updates an existing item
func (s *SyncStore) UpdateItem(item todoItem) error {
	err := s.local.UpdateItem(item)
	if err != nil {
		return err
	}

	s.local.LogChange("task", item.id, "update")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// UpdateTaskServerID updates a task's server ID after successful sync
func (s *SyncStore) UpdateTaskServerID(clientID string, serverID int) error {
	return s.local.UpdateTaskServerID(clientID, serverID)
}

// DeleteItem deletes an item
func (s *SyncStore) DeleteItem(id int) error {
	err := s.local.DeleteItem(id)
	if err != nil {
		return err
	}

	s.local.LogChange("task", id, "delete")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// GetLastSyncTime retrieves the last sync timestamp
func (s *SyncStore) GetLastSyncTime() (int64, error) {
	return s.local.GetLastSyncTime()
}

// SetLastSyncTime updates the last sync timestamp
func (s *SyncStore) SetLastSyncTime(timestamp int64) error {
	return s.local.SetLastSyncTime(timestamp)
}

// GetPendingChanges retrieves pending changes
func (s *SyncStore) GetPendingChanges() ([]Change, error) {
	return s.local.GetPendingChanges()
}

// MarkChangeSynced marks a change as synced
func (s *SyncStore) MarkChangeSynced(changeID int) error {
	return s.local.MarkChangeSynced(changeID)
}

// LogChange logs a change
func (s *SyncStore) LogChange(entityType string, entityID int, changeType string) error {
	return s.local.LogChange(entityType, entityID, changeType)
}

// FullSync performs a complete sync (pull then push)
func (s *SyncStore) FullSync() error {
	logger.LogInfo("Full sync started")
	start := time.Now()

	lastSync, _ := s.local.GetLastSyncTime()

	// Pull first to get latest state from server
	if err := s.PullChanges(lastSync); err != nil {
		logger.LogError("Full sync failed during pull", "duration", time.Since(start), "error", err)
		return fmt.Errorf("pull changes failed: %w", err)
	}

	// Push local changes
	if err := s.PushChanges(); err != nil {
		logger.LogError("Full sync failed during push", "duration", time.Since(start), "error", err)
		return fmt.Errorf("push changes failed: %w", err)
	}

	// Update last sync time
	s.local.SetLastSyncTime(time.Now().Unix())

	logger.LogInfo("Full sync completed successfully", "duration", time.Since(start))
	return nil
}

// PullChanges pulls changes from server and applies them locally
func (s *SyncStore) PullChanges(since int64) error {
	logger.LogDebug("Pulling changes from server", "since", since)

	resp, err := s.client.PullChanges(since)
	if err != nil {
		logger.LogError("Failed to pull changes", "error", err)
		return err
	}

	if resp == nil {
		logger.LogDebug("No changes received from server")
		return nil
	}

	// Apply task changes
	for _, serverTask := range resp.Tasks {
		// Try to find existing local task by client ID
		localTask, err := s.local.GetItemByClientID(serverTask.ClientID)
		if err != nil {
			// Doesn't exist locally - create it
			newItem := todoItem{
				clientID:     serverTask.ClientID,
				serverID:     0,
				done:         serverTask.Done,
				todo:         serverTask.Todo,
				priority:     serverTask.Priority,
				dateCompleted: serverTask.DateCompleted,
				dateAdded:    serverTask.DateAdded,
				dueDate:      serverTask.DueDate,
				deleted:      serverTask.Deleted,
				deletedAt:    serverTask.DeletedAt,
				todoListID:   serverTask.TodoListID,
				listClientID: serverTask.TodoListClientID,
				version:      serverTask.Version,
			}
			s.local.SaveItem(newItem)
			continue
		}

		// Task exists - resolve conflict using "last write wins"
		if serverTask.UpdatedAt > localTask.dateAdded {
			// Server is newer - update local copy
			localTask.done = serverTask.Done
			localTask.todo = serverTask.Todo
			localTask.priority = serverTask.Priority
			localTask.dateCompleted = serverTask.DateCompleted
			localTask.dueDate = serverTask.DueDate
			localTask.deleted = serverTask.Deleted
			localTask.deletedAt = serverTask.DeletedAt
			localTask.todoListID = serverTask.TodoListID
			localTask.listClientID = serverTask.TodoListClientID
			localTask.version = serverTask.Version
			s.local.UpdateItem(localTask)
		}
		// Otherwise local is newer, leave it as is
	}

	// Apply list changes
	for _, serverList := range resp.Lists {
		// For now, we'll skip list syncing as it requires more complex logic
		// with archived lists
		_ = serverList
	}

	return nil
}

// PushChanges pushes pending local changes to the server and updates local IDs
func (s *SyncStore) PushChanges() error {
	items, err := s.local.GetItems()
	if err != nil {
		logger.LogError("Failed to get items for push", "error", err)
		return err
	}

	lists, err := s.local.GetTodoLists()
	if err != nil {
		logger.LogError("Failed to get lists for push", "error", err)
		return err
	}

	logger.LogDebug("Pushing changes to server", "items", len(items), "lists", len(lists))

	// Push changes and get ID mappings back
	pushResp, err := s.client.PushChanges(items, lists)
	if err != nil {
		logger.LogError("Failed to push changes", "error", err)
		return err
	}

	// Process list ID mappings and update local database
	if pushResp != nil && len(pushResp.ListIDMappings) > 0 {
		logger.LogDebug("Processing list ID mappings", "count", len(pushResp.ListIDMappings))
		for _, mapping := range pushResp.ListIDMappings {
			if err := s.local.UpdateListServerID(mapping.ClientID, mapping.ServerID); err != nil {
				logger.LogWarn("Failed to update list server ID",
					"clientID", mapping.ClientID,
					"serverID", mapping.ServerID,
					"error", err)
			}
		}
	}

	// Process task ID mappings and update local database
	if pushResp != nil && len(pushResp.TaskIDMappings) > 0 {
		logger.LogDebug("Processing task ID mappings", "count", len(pushResp.TaskIDMappings))
		for _, mapping := range pushResp.TaskIDMappings {
			if err := s.local.UpdateTaskServerID(mapping.ClientID, mapping.ServerID); err != nil {
				logger.LogWarn("Failed to update task server ID",
					"clientID", mapping.ClientID,
					"serverID", mapping.ServerID,
					"error", err)
			}
		}
	}

	// Mark all changes as synced
	changes, err := s.local.GetPendingChanges()
	if err != nil {
		logger.LogError("Failed to get pending changes", "error", err)
		return err
	}

	for _, change := range changes {
		s.local.MarkChangeSynced(change.id)
	}

	logger.LogDebug("Push completed and changes marked as synced")
	return nil
}

// StartBackgroundSync starts the background sync goroutine
func (s *SyncStore) StartBackgroundSync() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.LogInfo("Background sync started", "interval", s.config.SyncIntervalSeconds)

	go func() {
		ticker := time.NewTicker(time.Duration(s.config.SyncIntervalSeconds) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logger.LogDebug("Background sync tick")
				if s.client.IsOnline() {
					if err := s.FullSync(); err != nil {
						logger.LogWarn("Background sync cycle failed", "error", err)
					} else {
						logger.LogDebug("Background sync cycle completed successfully")
					}
				} else {
					logger.LogDebug("Skipping background sync - server offline")
				}
			case <-s.stopCh:
				logger.LogInfo("Background sync stopped")
				return
			}
		}
	}()
}

// StopBackgroundSync stops the background sync goroutine
func (s *SyncStore) StopBackgroundSync() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		close(s.stopCh)
		s.running = false
	}
}
