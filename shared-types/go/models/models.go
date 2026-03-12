package models

// User represents a user account in the system
type User struct {
	ID        int    `json:"id"`
	APIKey    string `json:"api_key"`
	CreatedAt int64  `json:"created_at"`
}

// TodoList represents a list that organizes tasks
// This is the single source of truth for TodoList schema across all platforms
type TodoList struct {
	// Database identifier (assigned by server)
	ID int `json:"id"`

	// Client-side identifier (UUID) - used for offline-first sync
	// Ensures tasks created offline can be synced correctly later
	ClientID string `json:"client_id"`

	// Display name of the list
	Name string `json:"name"`

	// Order in which lists appear in the UI
	DisplayOrder int `json:"display_order"`

	// Whether this list is archived (soft delete)
	Archived bool `json:"archived"`

	// Timestamp when list was first created
	CreatedAt int64 `json:"created_at"`

	// Timestamp of last modification - CRITICAL for conflict resolution
	// Used in last-write-wins comparison during sync
	UpdatedAt int64 `json:"updated_at"`

	// Version number for optimistic locking / conflict detection
	Version int `json:"version"`
}

// Task represents a single todo item in a list
// This is the single source of truth for Task schema across all platforms
type Task struct {
	// Database identifier (assigned by server)
	ID int `json:"id"`

	// Client-side identifier (UUID) - used for offline-first sync
	// Ensures tasks created offline can be synced correctly later
	ClientID string `json:"client_id"`

	// Reference to the TodoList this task belongs to by numeric database ID
	// Used for database foreign keys
	TodoListID int `json:"todo_list_id"`

	// Reference to the TodoList this task belongs to by stable UUID (clientID)
	// This is used during sync to ensure correct relationships across devices
	// Even if server assigns a different numeric ID to the list, the UUID reference
	// allows the sync process to correctly associate tasks with their lists
	TodoListClientID string `json:"todo_list_client_id"`

	// The task description/text content
	Todo string `json:"todo"`

	// Priority level (1 = lowest, 4 = highest)
	Priority int `json:"priority"`

	// Whether this task is marked as complete
	Done bool `json:"done"`

	// Timestamp when task was originally created
	// Note: This is immutable and differs from UpdatedAt
	DateAdded int64 `json:"date_added"`

	// Timestamp when task was marked complete (null if not done)
	DateCompleted *int64 `json:"date_completed"`

	// Timestamp for when the task is due (null if no due date)
	DueDate *int64 `json:"due_date"`

	// Whether this task is soft-deleted
	Deleted bool `json:"deleted"`

	// Timestamp when task was deleted (null if not deleted)
	DeletedAt *int64 `json:"deleted_at"`

	// Timestamp when task was first created in database
	CreatedAt int64 `json:"created_at"`

	// Timestamp of last modification - CRITICAL for conflict resolution
	// Used in last-write-wins comparison during sync
	// IMPORTANT: Must compare against this field, not DateAdded
	UpdatedAt int64 `json:"updated_at"`

	// Version number for optimistic locking / conflict detection
	Version int `json:"version"`
}

// SyncRequest is the request format for pulling changes from server
type SyncRequest struct {
	// Timestamp from which to fetch changes
	// Only returns items modified after this timestamp
	Since int64 `json:"since"`
}

// SyncResponse contains all changes from the server since a given timestamp
type SyncResponse struct {
	// Array of tasks modified since the requested timestamp
	Tasks []Task `json:"tasks"`

	// Array of lists modified since the requested timestamp
	Lists []TodoList `json:"lists"`
}

// PushRequest contains changes to push from client to server
type PushRequest struct {
	// Tasks to create, update, or mark as deleted
	Tasks []Task `json:"tasks"`

	// Lists to create, update, or mark as deleted
	Lists []TodoList `json:"lists"`
}

// IDMapping represents the server ID assigned to a client-created item
// Used to map local IDs to server IDs after initial sync
type IDMapping struct {
	// Client-side identifier (UUID) used when creating the item
	ClientID string `json:"client_id"`

	// Server-assigned identifier after successful creation/update
	ServerID int `json:"server_id"`

	// Type of entity: "list" or "task"
	EntityType string `json:"entity_type"`
}

// PushResponse contains the result of a push operation
// Returns server-assigned IDs for items that were created/updated
type PushResponse struct {
	// Status of the push operation
	Status string `json:"status"`

	// List ID mappings from client to server
	// Used to update local database with server-assigned IDs
	ListIDMappings []IDMapping `json:"list_id_mappings"`

	// Task ID mappings from client to server
	// Used to update local database with server-assigned IDs
	TaskIDMappings []IDMapping `json:"task_id_mappings"`
}

// HealthResponse is the response from the health check endpoint
// Used to verify server is running and accessible
type HealthResponse struct {
	Status string `json:"status"`
}

// ErrorResponse is a standard error response format
// All API errors should follow this format
type ErrorResponse struct {
	Error string `json:"error"`
}

// Change represents a local change pending sync
// Used by CLI and mobile to track what needs to be uploaded to server
type Change struct {
	ID         int    `json:"id"`
	EntityType string `json:"entity_type"` // "task" or "list"
	EntityID   int    `json:"entity_id"`
	ChangeType string `json:"change_type"` // "create", "update", "delete"
	Timestamp  int64  `json:"timestamp"`
	Synced     bool   `json:"synced"`
}

// SyncStatus tracks the synchronization state
// Used by CLI and mobile to display sync progress/status
type SyncStatus struct {
	Online        bool   `json:"online"`
	Syncing       bool   `json:"syncing"`
	LastSyncTime  int64  `json:"last_sync_time"`
	PendingCount  int    `json:"pending_count"`
	ErrorMessage  string `json:"error_message"`
}
