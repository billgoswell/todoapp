# Repository Pattern

This package implements the **Repository Pattern** for database access, separating data access logic from business logic and providing a clean interface for interacting with the database.

## Overview

The repository pattern provides:
- **Clean separation of concerns** - Database queries isolated from handler logic
- **Testability** - Easy to mock repositories in unit tests
- **Maintainability** - Changes to database logic only affect repositories
- **Consistency** - All repositories follow the same patterns and conventions

## Structure

```
repositories/
├── user.go        # UserRepository - User account operations
├── task.go        # TaskRepository - Task/todo item operations
├── list.go        # ListRepository - Todo list operations
└── README.md      # This file
```

## Repositories

### UserRepository

Handles all user-related database operations.

```go
repo := repositories.NewUserRepository(conn)

// Create a new user
userID, err := repo.CreateUser("api-key-123")

// Get user by API key
userID, err := repo.GetUserByAPIKey("api-key-123")
```

**Methods:**
- `CreateUser(apiKey string) (int, error)` - Create a new user
- `GetUserByAPIKey(apiKey string) (int, error)` - Retrieve user by API key

### TaskRepository

Handles all task-related database operations.

```go
repo := repositories.NewTaskRepository(conn)

// Get tasks modified since timestamp
tasks, err := repo.GetTasksSince(userID, since)

// Insert or update a task (with conflict resolution)
id, err := repo.UpsertTask(userID, taskData)
```

**Methods:**
- `GetTasksSince(userID int, since int64) ([]map[string]interface{}, error)` - Fetch tasks modified since timestamp
- `UpsertTask(userID int, task map[string]interface{}) (int, error)` - Insert or update using last-write-wins

### ListRepository

Handles all list-related database operations.

```go
repo := repositories.NewListRepository(conn)

// Get lists modified since timestamp
lists, err := repo.GetListsSince(userID, since)

// Insert or update a list (with conflict resolution)
id, err := repo.UpsertList(userID, listData)
```

**Methods:**
- `GetListsSince(userID int, since int64) ([]map[string]interface{}, error)` - Fetch lists modified since timestamp
- `UpsertList(userID int, list map[string]interface{}) (int, error)` - Insert or update using last-write-wins

## Usage in Handlers

Handlers access repositories through the `DB` instance:

```go
// In a handler
func GetTasksHandler(db *db.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt("user_id")
        since := c.Query("since")

        // Access repository through db.Tasks
        tasks, err := db.Tasks.GetTasksSince(userID, timestamp)
        if err != nil {
            // Handle error
            return
        }

        c.JSON(http.StatusOK, gin.H{"tasks": tasks})
    }
}
```

## Backward Compatibility

The `DB` struct provides delegation methods for backward compatibility:

```go
// These methods delegate to the repositories
db.CreateUser(apiKey)        // → db.Users.CreateUser(apiKey)
db.GetUserByAPIKey(apiKey)   // → db.Users.GetUserByAPIKey(apiKey)
db.GetTasksSince(userID, ts) // → db.Tasks.GetTasksSince(userID, ts)
db.UpsertTask(userID, task)  // → db.Tasks.UpsertTask(userID, task)
db.GetListsSince(userID, ts) // → db.Lists.GetListsSince(userID, ts)
db.UpsertList(userID, list)  // → db.Lists.UpsertList(userID, list)
```

Existing code continues to work without modification.

## Testing

To test repositories in isolation:

```go
// Create a mock connection or use testcontainers
repo := repositories.NewTaskRepository(testConn)

// Test directly
tasks, err := repo.GetTasksSince(userID, since)
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}
```

## Conflict Resolution

Both `UpsertTask` and `UpsertList` implement **last-write-wins** conflict resolution:

- If the incoming version has `updated_at > existing.updated_at`, accept the new version
- Otherwise, keep the existing version
- This logic is implemented in the SQL `ON CONFLICT` clause

See `shared-types/sync/conflict.go` for the canonical algorithm documentation.

## Data Format

Repositories return data as `[]map[string]interface{}` to provide flexibility:

```go
task := map[string]interface{}{
    "id":              123,
    "client_id":       "uuid-456",
    "todo_list_id":    1,
    "todo":            "Buy groceries",
    "priority":        3,
    "done":            false,
    "date_added":      1234567890,    // Unix timestamp
    "date_completed":  nil,           // Optional
    "due_date":        1234567990,    // Optional
    "deleted":         false,
    "deleted_at":      nil,           // Optional
    "created_at":      1234567890,    // Unix timestamp
    "updated_at":      1234567900,    // Unix timestamp (for sync)
    "version":         1,             // For conflict detection
}
```

**Note:** In the future, this could be refactored to return typed structs instead of maps for better type safety.

## Adding New Repositories

To add a new repository (e.g., for a new entity):

1. Create `entity.go` in this directory
2. Define `EntityRepository` struct
3. Implement `NewEntityRepository(conn *sql.DB)` constructor
4. Implement query methods
5. Add field to `DB` struct in `postgres.go`
6. Initialize in `NewDB()` function

Example:
```go
// In entities/comment.go
type CommentRepository struct {
    conn *sql.DB
}

func NewCommentRepository(conn *sql.DB) *CommentRepository {
    return &CommentRepository{conn: conn}
}

func (r *CommentRepository) CreateComment(userID int, comment map[string]interface{}) (int, error) {
    // Implementation
}
```

Then update `postgres.go`:
```go
type DB struct {
    conn     *sql.DB
    Users    *repositories.UserRepository
    Tasks    *repositories.TaskRepository
    Lists    *repositories.ListRepository
    Comments *repositories.CommentRepository  // Add this
}

func NewDB(...) (*DB, error) {
    // ... existing code ...
    return &DB{
        conn:     conn,
        Users:    repositories.NewUserRepository(conn),
        Tasks:    repositories.NewTaskRepository(conn),
        Lists:    repositories.NewListRepository(conn),
        Comments: repositories.NewCommentRepository(conn),  // Add this
    }, nil
}
```

## Future Improvements

- [ ] Return typed structs instead of `map[string]interface{}`
- [ ] Add query builder helpers to reduce duplication
- [ ] Implement batch operations (UpsertTasksBatch, UpsertListsBatch)
- [ ] Add indexes for common queries
- [ ] Implement connection pooling strategies
- [ ] Add pagination support
- [ ] Add transaction support for multi-step operations

See `CODE_AUDIT_SUMMARY.md` for a complete list of planned improvements.
