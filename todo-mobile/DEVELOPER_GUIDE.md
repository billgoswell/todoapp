# CommandLineTodo Mobile - Developer Guide

Quick reference for working with the codebase.

## Quick Start

### 1. Initialize the App
```typescript
// In src/App.tsx (already done)
import { db } from './database';
import { apiClient } from './api';

await db.initialize();           // Create schema
await apiClient.initialize();    // Load credentials
```

### 2. Access Database
```typescript
import { db } from './database';

// Create
const task = await db.tasks.create({...});
const list = await db.lists.create({...});

// Read
const tasks = await db.tasks.findAll();
const list = await db.lists.findById(1);

// Update
await db.tasks.update(task);

// Delete (soft)
await db.tasks.delete(taskId);
```

### 3. Perform Sync
```typescript
import { performFullSync } from './api/sync';

const result = await performFullSync();
if (result.success) {
  // Sync successful
} else {
  // Handle error
  console.error(result.error);
}
```

## Module Overview

### `src/database/`
The data layer - all local storage operations.

**Key Classes:**
- `DatabaseManager` - Main entry point, manages all repos
- `TaskRepository` - Task CRUD operations
- `ListRepository` - List CRUD operations
- `SyncRepository` - Sync metadata storage
- `ChangeTracker` - Change log operations

**Usage:**
```typescript
import { db } from './database';

// All repositories are accessible
db.tasks.create(...)
db.lists.findAll()
db.sync.getLastSyncTime()
db.changes.recordTaskCreate(id)
```

### `src/api/`
The API layer - server communication.

**Key Files:**
- `client.ts` - `ApiClient` class for HTTP requests
- `sync.ts` - High-level sync operations
- `types.ts` - TypeScript interfaces

**Usage:**
```typescript
import { apiClient, performFullSync } from './api';

// Direct API calls
await apiClient.getTasks();
await apiClient.push(data);

// High-level sync
await performFullSync();
```

### `src/utils/`
Helper utilities.

**Modules:**
- `timestamps.ts` - Date/time utilities
- `uuid.ts` - UUID generation
- `storage.ts` - AsyncStorage wrapper
- `constants.ts` - App constants

**Usage:**
```typescript
import { now, formatDate } from './utils/timestamps';
import { generateTaskClientId } from './utils/uuid';
import { setApiKey, getApiKey } from './utils/storage';
import { TASK_PRIORITIES, SYNC_STATUS } from './utils/constants';
```

## Common Tasks

### Create a Task Locally
```typescript
import { db } from './database';
import { generateTaskClientId } from './utils/uuid';
import { now } from './utils/timestamps';

const task = await db.tasks.create({
  client_id: generateTaskClientId(),
  todo_list_id: 1,
  todo: 'Do something',
  priority: TASK_PRIORITIES.HIGH,
  done: false,
  date_added: now(),
  deleted: false,
  updated_at: now(),
  version: 1,
});

// Record change for sync
await db.changes.recordTaskCreate(task.id!);
```

### Update a Task
```typescript
const task = await db.tasks.findById(1);
task.todo = 'Updated description';
task.done = true;

await db.tasks.update(task);
await db.changes.recordTaskUpdate(task.id!);
```

### Get All Tasks for a List
```typescript
const tasks = await db.tasks.findByListId(listId);
```

### Sync with Server
```typescript
import { performFullSync, getSyncStatus } from './api/sync';

// Perform sync
const result = await performFullSync();

// Check status
const status = await getSyncStatus();
console.log(`Unsynced changes: ${status.unsyncedChanges}`);
```

### Check for Unsynced Changes
```typescript
const changes = await db.changes.getUnsyncedChanges();
const taskUpdates = await db.changes.getUnsyncedChangesByType('task');
const hasPending = await db.changes.hasPendingChanges('task', taskId);
```

### Get Sync Diagnostics
```typescript
import { getSyncDiagnostics } from './api/sync';

const diag = await getSyncDiagnostics();
console.log(diag.database);     // Task/list counts
console.log(diag.sync);         // Last sync info
console.log(diag.server);       // Server config
```

## Database Operations

### Transaction Pattern
```typescript
// Automatic transactions via repositories
const task = await db.tasks.create(taskData);  // Auto transaction
```

### Batch Operations
```typescript
// Bulk insert for sync
await db.tasks.bulkInsert(serverTasks);

// Bulk update for sync
await db.lists.bulkUpdate(serverLists);
```

### Change Tracking
```typescript
// Record different types of changes
await db.changes.recordTaskCreate(taskId);
await db.changes.recordTaskUpdate(taskId);
await db.changes.recordTaskDelete(taskId);

await db.changes.recordListCreate(listId);
await db.changes.recordListUpdate(listId);
await db.changes.recordListDelete(listId);

// Get unsynced
const unsyncedChanges = await db.changes.getUnsyncedChanges();

// Mark as synced after push succeeds
await db.changes.markAllSynced();

// Cleanup after sync
await db.changes.clearSyncedChanges();
```

## API Client Usage

### Set Server Configuration
```typescript
import { apiClient } from './api';

await apiClient.setCredentials(
  'http://localhost:8080/api/v1',
  'your-api-key'
);

// Credentials are saved to AsyncStorage automatically
```

### Test Connection
```typescript
const isConnected = await apiClient.testConnection();
if (!isConnected) {
  // Handle offline
}
```

### Direct API Operations
```typescript
// Get all tasks
const tasks = await apiClient.getTasks();

// Create task
const newTask = await apiClient.createTask(taskData);

// Update task
const updated = await apiClient.updateTask(taskId, taskData);

// Delete task
await apiClient.deleteTask(taskId);

// Same for lists
const lists = await apiClient.getLists();
```

### Sync Operations
```typescript
// Pull changes since timestamp
const pullResponse = await apiClient.pull(lastSyncTime);

// Push local changes
await apiClient.push({ tasks, lists });
```

## Error Handling

### API Errors
```typescript
import { apiClient } from './api';

try {
  await apiClient.pull(timestamp);
} catch (error) {
  if (apiClient.isUnauthorizedError(error)) {
    // Invalid API key - prompt user to re-authenticate
  } else if (apiClient.isNetworkError(error)) {
    // No connection - operate offline
  } else {
    // Other error
  }
}
```

### Sync Errors
```typescript
import { performFullSync, getSyncStatus } from './api/sync';

const result = await performFullSync();
if (!result.success) {
  const status = await getSyncStatus();
  console.error(status.lastSyncError);
}
```

## Type Safety

### Using TypeScript Interfaces
```typescript
import { Task, TodoList, PullResponse } from './api/types';

// Type-checked operations
const task: Task = {
  client_id: '...',
  todo_list_id: 1,
  todo: 'Task description',
  priority: 1,
  done: false,
  date_added: 123456,
  deleted: false,
  updated_at: 123456,
  version: 1,
};
```

## Constants Reference

### Task Priorities
```typescript
import { TASK_PRIORITIES, PRIORITY_LABELS, PRIORITY_COLORS } from './utils/constants';

TASK_PRIORITIES.HIGH      // 1
TASK_PRIORITIES.MEDIUM    // 2
TASK_PRIORITIES.LOW       // 3
TASK_PRIORITIES.NONE      // 4

PRIORITY_LABELS[1]        // "High"
PRIORITY_COLORS[1]        // "#FF3B30" (red)
```

### Sync Status
```typescript
import { SYNC_STATUS } from './utils/constants';

SYNC_STATUS.IDLE           // 'idle'
SYNC_STATUS.SYNCING        // 'syncing'
SYNC_STATUS.SUCCESS        // 'success'
SYNC_STATUS.ERROR          // 'error'
```

### Storage Keys
```typescript
import { STORAGE_KEYS } from './utils/constants';

STORAGE_KEYS.API_KEY
STORAGE_KEYS.SERVER_URL
STORAGE_KEYS.LAST_SYNC_TIME
STORAGE_KEYS.SYNC_ENABLED
```

## Timestamps

All timestamps are **Unix time in seconds** (not milliseconds).

```typescript
import { now, toDate, formatDate } from './utils/timestamps';

const timestamp = now();              // Current Unix timestamp
const date = toDate(timestamp);       // Convert to Date object
const formatted = formatDate(timestamp, 'MMM dd, yyyy');

// Utilities
isOverdue(timestamp)                  // Boolean
isToday(timestamp)                    // Boolean
getRelativeTime(timestamp)            // "2h ago"
```

## Debugging

### Database Diagnostics
```typescript
const stats = await db.getStats();
console.log(stats.taskCount);
console.log(stats.unsyncedChanges);
```

### Sync Diagnostics
```typescript
import { getSyncDiagnostics } from './api/sync';

const diag = await getSyncDiagnostics();
console.log(JSON.stringify(diag, null, 2));
```

### Clear All Data (Caution!)
```typescript
// Reset everything
await db.clearAllData();

// Reset only sync state
import { resetSyncState } from './api/sync';
await resetSyncState();
```

## Migration from Phase 1 to Phase 2

When implementing Phase 2 (state management):

1. **Create Context Providers**
   - Wrap app with TaskContext, ListContext, SyncContext
   - Initialize from database in useEffect

2. **Create Custom Hooks**
   - `useTasks()` - get tasks, create, update, delete
   - `useLists()` - manage lists
   - `useSync()` - trigger sync, get status

3. **Refactor Screen Usage**
   - Replace direct `db.*` calls with hooks
   - Update UI when context changes

4. **Example Hook Pattern**
   ```typescript
   export const useTasks = () => {
     const { tasks, setTasks } = useContext(TaskContext);

     const createTask = async (task: Task) => {
       const created = await db.tasks.create(task);
       await db.changes.recordTaskCreate(created.id!);
       setTasks([...tasks, created]);
       return created;
     };

     return { tasks, createTask };
   };
   ```

## State Management (Phase 2)

Access state through custom hooks:

```typescript
import { useTasks, useLists, useSync } from './state';

function MyScreen() {
  const { tasks, createTask } = useTasks();
  const { lists, createList } = useLists();
  const { performSync, isOnline } = useSync();

  // Use state and operations here
}
```

See `PHASE_2_STATE_MANAGEMENT.md` for comprehensive guide.

## Resources

- **Database Schema**: `src/database/schema.ts`
- **API Types**: `src/api/types.ts`
- **Implementation Plan**: `PLAN.md`
- **Phase 1 Details**: `PHASE_1_COMPLETE.md`
- **Phase 2 Details**: `PHASE_2_STATE_MANAGEMENT.md`
- **Phase 2 Summary**: `PHASE_2_SUMMARY.md`

## Support

For questions about the architecture or implementation:
- Check PLAN.md for architecture diagrams
- Review PHASE_1_COMPLETE.md for detailed component documentation
- Look at example usage patterns in this guide
