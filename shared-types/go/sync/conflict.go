package sync

import "github.com/billgoswell/commandlinetodo/shared-types/models"

// ConflictResolutionStrategy defines how conflicts are resolved during sync
// The application uses a "last-write-wins" strategy based on update timestamps
const (
	StrategyLastWriteWins = "last_write_wins"
)

// ResolutionResult indicates the outcome of conflict resolution
type ResolutionResult string

const (
	// ServerVersionWins indicates the server version is newer
	ServerVersionWins ResolutionResult = "server"
	// LocalVersionWins indicates the local version is newer
	LocalVersionWins ResolutionResult = "local"
	// NoConflict indicates there was no existing local version
	NoConflict ResolutionResult = "new"
)

// ResolveTaskConflict applies last-write-wins conflict resolution for a task
//
// Algorithm:
// 1. If no local task exists (new from server), accept server version
// 2. If local task exists, compare updated_at timestamps:
//    - If server.updated_at > local.updated_at: Accept server version
//    - Otherwise: Keep local version
//
// CRITICAL: Always use updated_at field, NOT date_added field
// - date_added: Original creation time (immutable)
// - updated_at: Last modification time (what we need for conflicts)
//
// Returns: ServerVersionWins if server is newer, LocalVersionWins otherwise
func ResolveTaskConflict(serverTask, localTask *models.Task) ResolutionResult {
	if localTask == nil {
		// No local version exists
		return NoConflict
	}

	// Compare timestamps - server version wins if newer
	if serverTask.UpdatedAt > localTask.UpdatedAt {
		return ServerVersionWins
	}

	// Local version is same age or newer - keep it
	return LocalVersionWins
}

// ResolveListConflict applies last-write-wins conflict resolution for a list
//
// Algorithm:
// 1. If no local list exists (new from server), accept server version
// 2. If local list exists, compare updated_at timestamps:
//    - If server.updated_at > local.updated_at: Accept server version
//    - Otherwise: Keep local version
//
// Returns: ServerVersionWins if server is newer, LocalVersionWins otherwise
func ResolveListConflict(serverList, localList *models.TodoList) ResolutionResult {
	if localList == nil {
		// No local version exists
		return NoConflict
	}

	// Compare timestamps - server version wins if newer
	if serverList.UpdatedAt > localList.UpdatedAt {
		return ServerVersionWins
	}

	// Local version is same age or newer - keep it
	return LocalVersionWins
}

// ApplyTaskUpdate copies relevant fields from server task to local task
// Call this after ResolveTaskConflict returns ServerVersionWins
func ApplyTaskUpdate(serverTask *models.Task, localTask *models.Task) {
	localTask.Done = serverTask.Done
	localTask.Todo = serverTask.Todo
	localTask.Priority = serverTask.Priority
	localTask.DateCompleted = serverTask.DateCompleted
	localTask.DueDate = serverTask.DueDate
	localTask.Deleted = serverTask.Deleted
	localTask.DeletedAt = serverTask.DeletedAt
	localTask.TodoListID = serverTask.TodoListID
	localTask.UpdatedAt = serverTask.UpdatedAt
	localTask.Version = serverTask.Version
	// NOTE: Do NOT update DateAdded - it's the original creation time
}

// ApplyListUpdate copies relevant fields from server list to local list
// Call this after ResolveListConflict returns ServerVersionWins
func ApplyListUpdate(serverList *models.TodoList, localList *models.TodoList) {
	localList.Name = serverList.Name
	localList.DisplayOrder = serverList.DisplayOrder
	localList.Archived = serverList.Archived
	localList.UpdatedAt = serverList.UpdatedAt
	localList.Version = serverList.Version
	// NOTE: Do NOT update CreatedAt - it's the original creation time
}
