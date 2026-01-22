# Handler Error Handling Integration Guide

## Overview

The `internal/handler` package provides utilities for consistent error handling in TUI event handlers. It simplifies error reporting to users and ensures consistent error code classification.

## Core Components

### 1. Error Creation Helpers

Handler-specific error creation functions that wrap the core error package:

#### `TaskOperationError(operation, err)`
For any task-related operation (create, update, delete, fetch)

```go
if err := m.store.UpdateItem(item); err != nil {
    appErr := TaskOperationError("update task", err)
    m.errorMsg = appErr.Message
    return m, nil
}
```

#### `ListOperationError(operation, err)`
For list-related operations

```go
if err := m.store.DeleteTodoList(listID); err != nil {
    appErr := ListOperationError("delete list", err)
    m.errorMsg = appErr.Message
    return m, nil
}
```

#### `SyncOperationError(operation, err)`
For synchronization operations

```go
err := syncStore.FullSync()
if err != nil {
    appErr := SyncOperationError("full sync", err)
    m.syncStatus.errorMessage = appErr.Message
}
```

#### `UserInputError(operation, field, reason)`
For input validation errors

```go
if err := validateTaskText(input); err != nil {
    appErr := UserInputError("create task", "task text", err.Error())
    m.errorMsg = appErr.Message
    return m, nil
}
```

#### `UserConfirmationError(operation)`
When user cancels an operation

```go
case KeyEsc:
    appErr := UserConfirmationError("delete task")
    // User cancelled - can handle gracefully
    m.returnToMain()
```

### 2. Message Handling

The `HandlerMessage` type provides structured feedback to users:

```go
type HandlerMessage struct {
    Level     MessageLevel  // success, info, warn, error
    Content   string        // Message to display
    ErrorCode string        // Machine-readable error code
}
```

Message levels:
- `MessageSuccess` - Operation completed successfully
- `MessageInfo` - Informational message
- `MessageWarn` - Warning that operation may have issues
- `MessageError` - Operation failed

### 3. Safe Operation Wrappers

Generic wrappers for error handling:

#### `SafeOperation(operation, fn)`
For operations that return error

```go
msg := SafeOperation("update task", func() error {
    return m.store.UpdateItem(item)
})

if msg != nil {
    m.errorMsg = msg.Content
}
```

#### `SafeDataStoreOperation[T](operation, fn)`
For operations that return a value and error

```go
item, msg := SafeDataStoreOperation("fetch task", func() (todoItem, error) {
    return m.store.GetItem(itemID)
})

if msg != nil {
    m.errorMsg = msg.Content
    return m, nil
}
```

### 4. Domain-Specific Handlers

Pre-built handlers for common operations:

#### Task Operations

```go
// Update task
msg := HandleTaskUpdate("update task", func() error {
    return m.store.UpdateItem(item)
})
if msg != nil {
    m.errorMsg = msg.Content
}

// Delete task
msg := HandleTaskDelete("delete task", func() error {
    return m.store.DeleteItem(itemID)
})
if msg != nil {
    m.errorMsg = msg.Content
}
```

#### List Operations

```go
// Create list
id, msg := HandleListCreate("create list", func() (int, error) {
    return m.store.CreateTodoList(name)
})
if msg != nil {
    m.errorMsg = msg.Content
}

// Update list
msg := HandleListUpdate("rename list", func() error {
    return m.store.UpdateTodoListName(listID, name)
})
if msg != nil {
    m.errorMsg = msg.Content
}

// Delete list
msg := HandleListDelete("delete list", func() error {
    return m.store.DeleteTodoList(listID)
})
if msg != nil {
    m.errorMsg = msg.Content
}
```

### 5. Error Classification

#### `IsRecoverableError(appErr)`
Check if user can retry the operation

Recoverable error codes:
- `ErrDataNotFound` - Item doesn't exist (show to user)
- `ErrValidationFailed` - Input invalid (user can fix)
- `ErrEmptyValue` - Required field empty (user can fix)
- `ErrLimitExceeded` - Too much input (user can fix)
- `ErrOffline` - No network (user can wait/retry)

#### `IsFatalError(appErr)`
Check if operation should terminate application

Fatal error codes:
- `ErrDatabaseSchema` - Database structure issue
- `ErrDatabaseTx` - Transaction failed

## Migration Pattern

### Before (Current Pattern in handlers.go)

```go
if err := m.store.UpdateItem(m.items[m.input.itemIndex]); err != nil {
    m.errorMsg = "Failed to update task: " + err.Error()
    logger.LogError("Failed to update task", "id", m.items[m.input.itemIndex].id, "error", err)
} else {
    logger.LogDebug("Task updated", "id", m.items[m.input.itemIndex].id)
}
```

### After (Using Handler Utilities)

```go
msg := HandleTaskUpdate("update task", func() error {
    return m.store.UpdateItem(m.items[m.input.itemIndex])
})

if msg != nil {
    m.errorMsg = msg.Content
} else {
    m.items[m.input.itemIndex].todo = editedText
    logger.LogDebug("Task updated", "id", m.items[m.input.itemIndex].id)
}
```

## Usage Examples

### Task Creation Flow

```go
func (m *model) handleTaskCreationFlow(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // ... existing logic ...

    case TaskFlowSetDueDate:
        newTask := todoItem{
            done:       false,
            todo:       m.taskFlow.text,
            priority:   m.taskFlow.priority,
            dateAdded:  time.Now().Unix(),
            dueDate:    dueDate,
            todoListID: m.currentListID,
        }

        opMsg := HandleTaskUpdate("create task", func() error {
            return m.store.SaveItem(newTask)
        })

        if opMsg != nil {
            m.errorMsg = opMsg.Content
        } else {
            m.items = append(m.items, newTask)
            m.sortItems()
        }

        m.taskFlow.reset()
        m.returnToMain()
        return m, nil
}
```

### List Management

```go
func (m *model) deleteCurrentList() {
    if m.input.listIndex >= len(m.todoLists) {
        return
    }

    selectedListID := m.todoLists[m.input.listIndex].id
    msg := HandleListDelete("delete list", func() error {
        return m.store.DeleteTodoList(selectedListID)
    })

    if msg != nil {
        m.errorMsg = msg.Content
        m.currentSubState = SubStateNone
        return
    }

    // ... clean up after successful deletion ...
}
```

### Sync Operations

```go
case KeyS:
    if m.syncEnabled {
        if syncStore, ok := m.store.(*SyncStore); ok {
            m.syncStatus.syncing = true
            m.syncStatus.errorMessage = ""
            m.spinnerFrame = 0

            syncCmd := func() tea.Msg {
                err := syncStore.FullSync()
                if err != nil {
                    appErr := SyncOperationError("full sync", err)
                    m.syncStatus.errorMessage = appErr.Message
                    m.syncStatus.syncing = false
                    return syncCompleteMsg{
                        err:      appErr,
                        syncTime: 0,
                    }
                }
                return syncCompleteMsg{
                    err:      nil,
                    syncTime: time.Now().Unix(),
                }
            }

            return m, tea.Batch(syncCmd, tickSpinner())
        }
    }
```

## Error Message Display

### In Model State

Store error messages in `m.errorMsg` field:

```go
type model struct {
    // ... other fields ...
    errorMsg string // Display errors in render.go
}
```

### In Render View

Display error messages with styling:

```go
if m.errorMsg != "" {
    s = append(s, ErrorStyle.Render("Error: "+m.errorMsg))
}
```

## Best Practices

1. **Use domain-specific handlers** - `HandleTaskUpdate()`, `HandleListCreate()` instead of generic `SafeOperation()`
2. **Provide meaningful operation names** - "update task" not just "update"
3. **Let handlers set success messages** - Domain handlers include success feedback
4. **Clear errors after display** - Set `m.errorMsg = ""` after returning to main view
5. **Check message level** - Don't treat warnings as fatal errors
6. **Log important operations** - Still log successful operations at DEBUG level

## Testing Handler Errors

```go
func TestHandlerTaskUpdate(t *testing.T) {
    m := &model{errorMsg: ""}

    msg := HandleTaskUpdate("update", func() error {
        return nil // or return an error
    })

    if msg == nil {
        t.Error("expected error message")
    }

    if msg.Level != MessageSuccess {
        t.Errorf("expected success, got %v", msg.Level)
    }
}
```

## Performance Considerations

- Error handlers add minimal overhead
- AppError.Log() uses efficient structured logging
- No additional database queries introduced
- Error message formatting lazy-evaluated

## See Also

- `internal/errors/` - Core error handling types and functions
- `internal/storage/db_errors.go` - Database operation wrappers
- `cmd/app/handlers.go` - Handler implementations
