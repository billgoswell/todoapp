# Database Error Handling Integration Guide

## Overview

The `db_errors.go` module provides structured error handling helpers for database operations. These functions wrap raw database calls with the centralized error handling from `internal/errors/`.

## Integration Pattern

### Before (Current Pattern in db.go)

```go
import (
    "database/sql"
    "github.com/billgoswell/commandlinetodo/internal/logger"
)

func getItemsFromDB() ([]todoItem, error) {
    rows, err := db.Query("SELECT ... FROM tasks WHERE deleted = 0")
    if err != nil {
        logger.LogError("Failed to query items", "error", err)
        return []todoItem{}, err
    }
    defer rows.Close()

    items := []todoItem{}
    for rows.Next() {
        var item todoItem
        if err := rows.Scan(&item.id, ...); err != nil {
            logger.LogError("Failed to scan item", "error", err)
            return []todoItem{}, err
        }
        items = append(items, item)
    }

    if err := rows.Err(); err != nil {
        logger.LogError("Error iterating items", "error", err)
        return []todoItem{}, err
    }
    return items, nil
}
```

### After (With Error Handling)

```go
import (
    "database/sql"
    "github.com/billgoswell/commandlinetodo/internal/storage"
)

func getItemsFromDB() ([]todoItem, error) {
    rows, err := storage.QueryWithError(db, "query items",
        "SELECT ... FROM tasks WHERE deleted = 0")
    if err != nil {
        return []todoItem{}, err
    }
    defer rows.Close()

    items := []todoItem{}
    for rows.Next() {
        var item todoItem
        if err := rows.Scan(&item.id, ...); err != nil {
            appErr := errors.QueryError("scan item", err)
            appErr.Log()
            return []todoItem{}, appErr
        }
        items = append(items, item)
    }

    if _, err := storage.ScanRowsWithError(rows, "iterate items"); err != nil {
        return []todoItem{}, err
    }
    return items, nil
}
```

## Helper Functions

### Query Operations

#### `QueryWithError(db, operation, query, args...)`
Wraps `db.Query()` with error handling.

**Returns:**
- `(*sql.Rows, error)` - Query results or structured AppError
- AppError code: `ErrDatabaseQuery`

```go
rows, err := storage.QueryWithError(db, "get all tasks",
    "SELECT id, todo FROM tasks WHERE deleted = 0")
if err != nil {
    return err
}
defer rows.Close()
```

#### `QueryRowWithError(db, operation, query, args...)`
Wraps `db.QueryRow()` (returns *sql.Row directly, no error handling at call time).

```go
var count int
row := storage.QueryRowWithError(db, "count tasks",
    "SELECT COUNT(*) FROM tasks WHERE deleted = 0")
err := row.Scan(&count)
```

#### `ScanRowsWithError(rows, operation)`
Checks row iteration errors and returns AppError if needed.

```go
rows, err := storage.QueryWithError(db, "get items", query)
if err != nil { return err }
defer rows.Close()

// ... iterate rows ...

if _, err := storage.ScanRowsWithError(rows, "iterate items"); err != nil {
    return err
}
```

### Execution Operations

#### `ExecuteWithError(db, operation, query, args...)`
Wraps `db.Exec()` with error handling (INSERT/UPDATE/DELETE).

**Returns:**
- `error` - nil on success, AppError on failure
- AppError code: `ErrDatabaseIO`

```go
err := storage.ExecuteWithError(db, "insert task",
    "INSERT INTO tasks (todo, priority) VALUES (?, ?)",
    "Buy milk", 1)
if err != nil {
    return err
}
```

#### `ExecuteWithIDError(db, operation, query, args...)`
Wraps `db.Exec()` and retrieves `LastInsertId()` with error handling.

**Returns:**
- `(int, error)` - New ID or AppError
- AppError code: `ErrDatabaseIO`

```go
newID, err := storage.ExecuteWithIDError(db, "create todo list",
    "INSERT INTO todoLists (name, created_at) VALUES (?, ?)",
    "My List", now())
if err != nil {
    return 0, err
}
```

### Transaction Operations

#### `BeginWithError(db, operation)`
Wraps `db.Begin()` with error handling.

```go
tx, err := storage.BeginWithError(db, "delete list and tasks")
if err != nil {
    return err
}
defer storage.RollbackWithError(tx, "delete list and tasks")
```

#### `TxExecuteWithError(tx, operation, query, args...)`
Wraps `tx.Exec()` within a transaction.

```go
err := storage.TxExecuteWithError(tx, "archive list",
    "UPDATE todoLists SET archived = 1 WHERE id = ?", listID)
if err != nil {
    return err
}
```

#### `CommitWithError(tx, operation)`
Wraps `tx.Commit()` with error handling.

```go
if err := storage.CommitWithError(tx, "finalize delete list"); err != nil {
    return err
}
```

#### `RollbackWithError(tx, operation)`
Safely rolls back transaction (logs but doesn't return error).

```go
defer storage.RollbackWithError(tx, "delete list and tasks")
```

## Migration Checklist

For each database function in `cmd/app/db.go`:

- [ ] Identify database operations (Query, Exec, Begin, Commit, Rollback)
- [ ] Replace with corresponding `storage.*WithError()` function
- [ ] Remove manual `logger.LogError()` calls (AppError.Log() handles it)
- [ ] Test that operation still works
- [ ] Verify error messages are meaningful
- [ ] Check that AppError codes match operation type

## Error Code Reference

| Helper | Error Code | Use Case |
|--------|-----------|----------|
| QueryWithError | `ErrDatabaseQuery` | SELECT operations |
| ExecuteWithError | `ErrDatabaseIO` | INSERT/UPDATE/DELETE |
| ExecuteWithIDError | `ErrDatabaseIO` | INSERT with LastInsertId |
| BeginWithError | `ErrDatabaseTx` | Transaction start |
| TxExecuteWithError | `ErrDatabaseTx` | Transaction operations |
| CommitWithError | `ErrDatabaseTx` | Transaction commit |

## Backward Compatibility

The storage helpers can coexist with existing `executeStmt()` and `executeStmtWithID()` functions. Migrate functions gradually:

1. Keep old functions unchanged
2. Add new functions using storage helpers
3. Gradually replace old function internals
4. Remove old functions after full migration

## Error Details Map

AppError.Details map can include context-specific information:

```go
// Example with custom details
appErr := errors.QueryError("get task", dbErr)
appErr.Details["query"] = "SELECT * FROM tasks WHERE id = ?"
appErr.Details["id"] = taskID

// Later access
if errors.IsErrorCode(err, errors.ErrDatabaseQuery) {
    taskID := err.(*errors.AppError).Details["id"]
}
```

## Testing

When testing database operations with error handling:

```go
func TestGetItems_QueryError(t *testing.T) {
    // Create invalid database to trigger error
    db.Close()
    _, err := getItemsFromDB()

    if err == nil {
        t.Error("expected error, got nil")
    }
    if !errors.IsErrorCode(err, errors.ErrDatabaseQuery) {
        t.Errorf("expected ErrDatabaseQuery, got %s",
            errors.ErrorCode(err))
    }
}
```

## Performance Considerations

- Error helpers add minimal overhead (primarily function call + error wrapping)
- AppError.Log() uses structured logging (efficient)
- No additional database round-trips introduced
- String formatting only happens on error path

## See Also

- `internal/errors/errors.go` - AppError type definitions
- `cmd/app/db.go` - Current database implementation
- `internal/logger/logger.go` - Structured logging
