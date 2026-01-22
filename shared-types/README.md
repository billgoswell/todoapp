# Shared Types Package

This package contains the canonical type definitions for the CommandLineTodo application, shared across all three platforms (Server, CLI, and Mobile).

## Purpose

The shared-types package serves as the **single source of truth** for data model definitions, eliminating duplication and ensuring consistency across:

1. **todo-server** (Go) - REST API backend with PostgreSQL
2. **todo-cmdline** (Go) - CLI TUI application with SQLite
3. **todo-mobile** (TypeScript) - React Native mobile app with SQLite

## Structure

```
shared-types/
├── go/
│   ├── models/
│   │   └── models.go       # Go struct definitions
│   ├── go.mod             # Go module definition
│   └── go.sum             # Go dependencies
└── README.md              # This file
```

## Types

### Core Data Models

#### Task
Represents a single todo item.

**Fields:**
- `id` (int) - Server-assigned unique identifier
- `client_id` (string) - UUID for offline-first sync
- `todo_list_id` (int) - Reference to parent list
- `todo` (string) - Task description
- `priority` (int) - Priority level (1-4, where 1=low, 4=high)
- `done` (bool) - Completion status
- `date_added` (int64) - Original creation time (immutable)
- `date_completed` (int64) - Time marked complete (nullable)
- `due_date` (int64) - Due date target (nullable)
- `deleted` (bool) - Soft delete flag
- `deleted_at` (int64) - When deleted (nullable)
- `created_at` (int64) - Database creation timestamp
- `updated_at` (int64) - **Last modification timestamp (CRITICAL for sync)**
- `version` (int) - Version for conflict detection

**Key Implementation Note:**
When resolving sync conflicts, always compare `updated_at` timestamps, NOT `date_added`. The `date_added` field represents task creation time and never changes. The `updated_at` field reflects the actual last modification, which is necessary for proper last-write-wins conflict resolution.

#### TodoList
Represents an organized collection of tasks.

**Fields:**
- `id` (int) - Server-assigned unique identifier
- `client_id` (string) - UUID for offline-first sync
- `name` (string) - Display name
- `display_order` (int) - Ordering in the UI
- `archived` (bool) - Archived status (soft delete)
- `created_at` (int64) - Database creation timestamp
- `updated_at` (int64) - **Last modification timestamp (CRITICAL for sync)**
- `version` (int) - Version for conflict detection

#### User
Represents a system user/account.

**Fields:**
- `id` (int) - Unique identifier
- `api_key` (string) - API authentication key
- `created_at` (int64) - Account creation timestamp

### Sync Models

#### SyncRequest / PullRequest
Request format for pulling changes from the server.

```go
type SyncRequest struct {
    Since int64 // Only return items modified after this timestamp
}
```

#### SyncResponse / PullResponse
Server response containing changes since requested timestamp.

```go
type SyncResponse struct {
    Tasks []Task     // Modified tasks
    Lists []TodoList // Modified lists
}
```

#### PushRequest
Request format for pushing local changes to server.

```go
type PushRequest struct {
    Tasks []Task     // Tasks to create/update/delete
    Lists []TodoList // Lists to create/update/delete
}
```

### Change Tracking
Used by clients to track local modifications pending sync.

```go
type Change struct {
    ID         int    // Database identifier
    EntityType string // "task" or "list"
    EntityID   int    // ID of the entity
    ChangeType string // "create", "update", or "delete"
    Timestamp  int64  // When the change occurred
    Synced     bool   // Whether it's been synced to server
}
```

### Sync Status
Tracks the current synchronization state.

```go
type SyncStatus struct {
    Online       bool   // Server connectivity status
    Syncing      bool   // Currently performing sync
    LastSyncTime int64  // Timestamp of last successful sync
    PendingCount int    // Number of changes awaiting sync
    ErrorMessage string // Last sync error, if any
}
```

## Usage

### Go Applications (Server and CLI)

Both Go applications import these types via Go modules:

```go
import "github.com/billgoswell/commandlinetodo/shared-types/models"

// Use the types
task := models.Task{
    ClientID: "uuid-123",
    Todo: "Buy groceries",
    Priority: 3,
}
```

**Configuration in go.mod:**
```
require github.com/billgoswell/commandlinetodo/shared-types v0.0.0

replace github.com/billgoswell/commandlinetodo/shared-types => ../shared-types/go
```

### TypeScript Applications (Mobile)

TypeScript types are defined in `todo-mobile/src/api/types.ts` using the same field names and structure, maintaining API compatibility:

```typescript
import type { Task, TodoList } from "@/api/types";

const task: Task = {
  client_id: "uuid-123",
  todo: "Buy groceries",
  priority: 3,
  done: false,
  // ... other fields
};
```

## Sync Algorithm & Conflict Resolution

All platforms implement **last-write-wins** conflict resolution using timestamps.

### Sync Flow

1. **Pull changes:** Fetch all items modified since `lastSyncTime`
2. **Resolve conflicts:** For each item, compare `updated_at` timestamps
   - If server's `updated_at` > local `updated_at`: Accept server version
   - Otherwise: Keep local version
3. **Push changes:** Send all local items (including changes)
4. **Update sync time:** Set `lastSyncTime` to current timestamp

### Canonical Conflict Resolution

The shared sync strategy is defined in `sync/conflict.go`:

```go
// Returns ServerVersionWins if server is newer, LocalVersionWins otherwise
func ResolveTaskConflict(serverTask, localTask *models.Task) ResolutionResult

// Apply server changes to local task
func ApplyTaskUpdate(serverTask, localTask *models.Task)
```

### CRITICAL IMPLEMENTATION DETAILS

**Always use `updated_at` for timestamp comparison, NEVER `date_added`:**
- `date_added`: Original creation time (immutable)
- `updated_at`: Last modification time (used for conflict resolution)

Example of WRONG comparison:
```go
// WRONG - compares creation time, not last change time
if serverTask.UpdatedAt > localTask.DateAdded { }

// CORRECT - compares modification times
if serverTask.UpdatedAt > localTask.UpdatedAt { }
```

### Field Update Rules

When applying server changes to local entities:
- **Always update:** Done, Todo, Priority, DateCompleted, DueDate, Deleted, DeletedAt, UpdatedAt, Version
- **Never update:** DateAdded (creation time), ID (local), ClientID (local)
- **Similar rules for lists:** Update Name, DisplayOrder, Archived, UpdatedAt, Version; Never update CreatedAt, ID, ClientID

### Conflict Scenarios

| Scenario | Resolution | Example |
|----------|-----------|---------|
| New on server | Accept server | Task created on mobile, first sync |
| Server updated more recently | Accept server | User edited task on web, mobile has stale copy |
| Local updated more recently | Keep local | User edited task on CLI after last sync |
| Same timestamp | Keep local | Simultaneous edits (ties go to local) |

### Testing Conflict Resolution

See `sync/conflict_test.go` for comprehensive test cases demonstrating all scenarios.

## JSON Field Naming

Types use `snake_case` JSON tags for API compatibility:

```
Go struct field  →  JSON field
id              →  id
clientID        →  client_id
todoListID      →  todo_list_id
dateAdded       →  date_added
dateCompleted   →  date_completed
dueDate         →  due_date
deletedAt       →  deleted_at
createdAt       →  created_at
updatedAt       →  updated_at
```

## Compatibility Matrix

| Type | Server (Go) | CLI (Go) | Mobile (TS) | Notes |
|------|------------|---------|------------|-------|
| Task | ✅ | ✅ | ✅ | Direct struct or interface |
| TodoList | ✅ | ✅ | ✅ | Direct struct or interface |
| User | ✅ | ⏳ | ⏳ | Server only for now |
| Change | ✅ | ✅ | ✅ | Change log tracking |
| SyncStatus | ✅ | ✅ | ✅ | UI display |
| SyncRequest | ✅ | ✅ | ✅ | API communication |
| SyncResponse | ✅ | ✅ | ✅ | API communication |
| PushRequest | ✅ | ✅ | ✅ | API communication |

## Migration Guide

### For Server Handlers
When importing types, use the models package:

```go
// Old (before refactoring)
import "github.com/billgoswell/commandlinetodo-server/internal/models"

// New (after refactoring)
import "github.com/billgoswell/commandlinetodo/shared-types/models"

// Or via type re-export
import "github.com/billgoswell/commandlinetodo-server/internal/models"
// models.Task is now re-exported from shared-types
```

### For CLI Code
Keep existing `todoItem` and `todoList` struct names for TUI code. They maintain compatibility with the shared types through field alignment.

### For Mobile Code
Continue using TypeScript interfaces defined in `src/api/types.ts`. These are compatible with the server API and are designed separately from the Go types.

## Future Enhancements

- [ ] TypeScript/JavaScript npm package export
- [ ] OpenAPI/Swagger schema generation
- [ ] API documentation generation
- [ ] GraphQL schema generation
- [ ] Database schema documentation
- [ ] Type validation library

## Contributing

When modifying types in `shared-types/go/models/models.go`:

1. Update the Go definition
2. Update TypeScript interfaces in `todo-mobile/src/api/types.ts`
3. Update this README if fields change
4. Run `go mod tidy` in all Go projects
5. Test all three platforms

## Documentation References

- **CLAUDE.md** - Project overview and architecture
- **PLAN.md** - Comprehensive improvement roadmap
- **TESTING.md** - API testing guide
- **CONFIG.md** - Configuration reference
