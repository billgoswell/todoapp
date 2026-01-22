# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

CommandLineTodo is a multi-platform todo application with three main components:

1. **todo-cmdline** - Go CLI application with TUI using Bubble Tea framework
2. **todo-server** - Go REST API server built with Gin framework (PostgreSQL backend)
3. **todo-mobile** - React Native mobile app with offline-first sync capability

The architecture supports offline-first operation with synchronization across devices using a last-write-wins conflict resolution strategy.

**Important Documentation:**
- **PLAN.md** - Comprehensive improvement and testing roadmap (6-week phased plan, 523 lines)
- **PHASE_1_PROGRESS.md** - Phase 1 completion summary (critical bug fix + testing infrastructure)
- **PHASE_2_PROGRESS.md** - Phase 2 progress (server testing framework with 50+ test cases)
- **DEPLOYMENT_PLAN.md** - Production deployment guide for Proxmox VE
- **TESTING.md** - API testing guide with curl examples
- **CONFIG.md** - CLI configuration reference

## Project Structure

```
todoapp/
├── shared-types/          # SHARED TYPE DEFINITIONS (canonical source of truth)
│   ├── go/
│   │   ├── models/
│   │   │   └── models.go  # Go struct definitions for Task, TodoList, User, etc.
│   │   ├── go.mod         # Shared-types module
│   │   └── go.sum
│   └── README.md          # Shared types documentation
├── todo-cmdline/          # CLI todo app
│   ├── cmd/app/           # Main application entry point and logic
│   ├── go.mod             # CLI dependencies (imports shared-types)
│   ├── Makefile           # Build commands
│   ├── README.md          # CLI documentation
│   └── CONFIG.md          # Configuration guide
├── todo-server/           # REST API backend
│   ├── cmd/server/        # Server entry point
│   ├── internal/
│   │   ├── db/            # PostgreSQL database layer
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── middleware/    # Auth middleware
│   │   └── models/        # Data models (re-exports from shared-types)
│   ├── go.mod             # Server dependencies (imports shared-types)
│   └── TESTING.md         # API testing guide
└── todo-mobile/           # React Native mobile app
    ├── src/
    │   ├── api/types.ts   # TypeScript type definitions (mirrors shared-types)
    │   ├── screens/       # Full-page components
    │   ├── components/    # Reusable UI components
    │   ├── state/         # Global state and contexts
    │   ├── database/      # SQLite repositories
    │   └── sync/          # Sync engine and API client
    ├── package.json       # Dependencies
    ├── app.json           # Expo configuration
    └── DEVELOPER_GUIDE.md # Mobile development info
```

## Building and Running

### CLI Application

```bash
# Build
cd todo-cmdline
make build

# Run (builds and executes)
make run

# Run tests
make test

# Clean build artifacts
make clean

# View help
make help
```

**Key files:**
- `cmd/app/main.go` - Application entry point; initializes DB, config, sync, and TUI
- `cmd/app/config.go` - Configuration loading from `~/.config/todo/config`
- `cmd/app/db.go` - SQLite database initialization and schema
- `cmd/app/datastore.go` - Local-only data operations interface
- `cmd/app/syncstore.go` - Wrapper for sync functionality
- `cmd/app/syncclient.go` - HTTP client for server communication
- `cmd/app/handlers.go` - TUI event handlers and keybindings
- `cmd/app/render.go` - Terminal UI rendering with Lipgloss
- `cmd/app/model.go` - Core data structures (todoItem, todoList)

### Server Application

The server requires PostgreSQL and runs on configurable host/port (default 127.0.0.1:8080).

**Architecture:**
- **Repository Pattern** - Database access logic isolated in repositories (UserRepository, TaskRepository, ListRepository)
- **Handlers** - HTTP handlers access repositories through DB instance
- **Models** - Uses shared-types for canonical data definitions

**Key files:**
- `internal/db/postgres.go` - Database connection and migration management
- `internal/db/repositories/` - Query implementations (user.go, task.go, list.go)
- `internal/handlers/` - HTTP request handlers
- `cmd/server/main.go` - Server entry point

**Configuration (environment variables):**
- `DB_HOST` (default: localhost)
- `DB_PORT` (default: 5432)
- `DB_USER` (default: postgres)
- `DB_PASSWORD` (default: postgres)
- `DB_NAME` (default: commandlinetodo)
- `SERVER_HOST` (default: 127.0.0.1)
- `SERVER_PORT` (default: 8080)

```bash
# Build and run
cd todo-server
go build -o server ./cmd/server/main.go
./server

# Or with environment variables
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres
export DB_NAME=commandlinetodo SERVER_HOST=127.0.0.1 SERVER_PORT=8080
./server
```

**Testing the server** - See `TESTING.md` for comprehensive API testing guide including:
- Health check endpoint
- Authentication and Bearer token format
- Creating/updating tasks and lists
- Pull/push sync operations
- Conflict resolution testing
- Soft delete behavior

**Deployment** - See `DEPLOYMENT_PLAN.md` for comprehensive production deployment guide to Proxmox VE including:
- 2-VM architecture (app-server + monitoring)
- PostgreSQL configuration
- SystemD service setup with auto-restart
- Automated backup strategy
- Firewall rules (local network only)
- Prometheus + Grafana monitoring setup
- Disaster recovery procedures

### Mobile Application

```bash
cd todo-mobile

# Install dependencies
npm install

# Development server
npm start

# Run on iOS (macOS only)
npm run ios

# Run on Android
npm run android

# Run tests
npm test

# Type checking
npm run type-check

# Linting
npm run lint
```

## Key Architecture Patterns

### Shared Type Definitions

All three platforms use canonical type definitions from the `shared-types/go/models/` package:

**Core Data Models:**
- `Task` - Single todo item with client_id, updated_at, version for sync
- `TodoList` - Container for tasks with sync metadata
- `User` - User account information
- `Change` - Local change tracking for pending sync
- `SyncStatus` - Synchronization state (online, syncing, pending count)

**Sync Protocol Models:**
- `SyncRequest` / `PullRequest` - Fetch changes since timestamp
- `SyncResponse` / `PullResponse` - Server response with changes
- `PushRequest` - Client sends changes to server

**Key Implementation Detail - Sync Timestamp Field:**
Use `updated_at` (last modification timestamp) for conflict resolution, NOT `date_added` (creation time). This is CRITICAL for correct last-write-wins logic.

See `shared-types/README.md` for complete field documentation.

### Sync Protocol

The system uses a **timestamp-based pull/push** architecture:

- **Pull**: Client fetches all items modified since `lastSyncTime` from server
- **Push**: Client sends local changes; server applies using `ON CONFLICT` (PostgreSQL) or `INSERT OR REPLACE` (SQLite)
- **Conflict Resolution**: Last-write-wins based on `updated_at` timestamp
- **Offline Support**: Changes queue locally until connectivity is restored

### Local Storage

- **CLI**: SQLite database at path specified in config (`~/.config/todo/config`)
- **Mobile**: SQLite via expo-sqlite with schema for tasks, lists, and sync metadata
- **Server**: PostgreSQL with user isolation (API key authentication)

### API Endpoints

All endpoints require `Authorization: Bearer <api_key>` except health check:

```
GET  /api/v1/health           # Health check (no auth)

# Lists
GET  /api/v1/lists            # Get all lists
POST /api/v1/lists            # Create list
PUT  /api/v1/lists/:id        # Update list
DELETE /api/v1/lists/:id      # Delete list (soft delete)

# Tasks
GET  /api/v1/tasks            # Get all tasks
POST /api/v1/tasks            # Create task
PUT  /api/v1/tasks/:id        # Update task
DELETE /api/v1/tasks/:id      # Delete task (soft delete)

# Sync
POST /api/v1/sync/pull        # Pull changes {since: timestamp}
POST /api/v1/sync/push        # Push changes {tasks: [...], lists: [...]}
```

## Configuration Files

### CLI Config

Located at `~/.config/todo/config` (JSON format):

```json
{
  "db_path": "./todo.db",
  "log_level": "info",
  "sync": {
    "enabled": false,
    "server_url": "",
    "api_key": "",
    "device_id": "<auto-generated-uuid>",
    "sync_interval_seconds": 60,
    "auto_sync_on_change": true,
    "retry_attempts": 3,
    "timeout_seconds": 10
  }
}
```

See `todo-cmdline/CONFIG.md` for detailed configuration options.

### Mobile App Config

Via environment variables or app settings screen:
- `SERVER_URL` - API server base URL
- `API_KEY` - Authentication key for server

## Development Workflows

### Adding a New API Endpoint

1. Define handler function in `todo-server/internal/handlers/` (receives `*db.DB`)
2. Register route in `cmd/server/main.go` under appropriate group (auth-required or public)
3. Handler automatically receives database and can perform queries
4. Return JSON responses; middleware handles errors

### Adding CLI Features

1. Add data structure to `model.go` if needed
2. Implement handler in `handlers.go` for keyboard input
3. Update `render.go` for UI display if needed
4. Modify `constants.go` for keybindings/messages

### Mobile Development

State management uses React Context API + Hooks pattern:
- `src/state/` - Global state and contexts
- `src/screens/` - Full-page components
- `src/components/` - Reusable UI components
- `src/database/` - SQLite repositories
- `src/sync/` - Sync engine and API client

## Testing

**See PLAN.md for comprehensive 6-week testing improvement roadmap:**
- Phase 1: Critical fixes & foundation
- Phase 2: Server handler & database tests
- Phase 3: CLI handler & sync tests
- Phase 4: Mobile API client & component tests
- Phase 5: Integration & E2E testing
- Phase 6: Code quality & documentation

**Current test coverage:** 3.95% (3 test files across all components)

### CLI Tests
```bash
cd todo-cmdline
go test ./...
```

Key test files:
- `cmd/app/utils_test.go` - Utility function tests (7 tests)
- `cmd/app/syncstore_test.go` - Sync logic tests (6 tests)

### Server Tests

Covered in `TESTING.md`. Manual curl-based testing recommended for API endpoints.

**Note:** Only conversion helpers currently tested; full handler tests in PLAN.md Phase 2.

### Mobile Tests
```bash
npm test                      # Run all tests
npm test -- --testPathPattern="unit"        # Unit tests only
npm test -- --testPathPattern="integration" # Integration tests
npm test -- --coverage                      # With coverage report
```

**Note:** Jest is configured but no tests implemented yet; see PLAN.md Phase 4.

## Dependencies

### CLI
- `charmbracelet/bubbletea` - TUI framework
- `charmbracelet/bubbles` - TUI components
- `charmbracelet/lipgloss` - Terminal styling
- `modernc.org/sqlite` - SQLite driver
- `google/uuid` - UUID generation
- `mergestat/timediff` - Relative time formatting

### Server
- `gin-gonic/gin` - HTTP framework
- `lib/pq` - PostgreSQL driver

### Mobile
- `react-native` 0.81.5 / `react` 19.1.0
- `expo` 54.0.31 - Build system and native modules
- `expo-sqlite` - Local database
- `react-navigation` - Screen navigation
- `axios` - HTTP client

## Logging

### CLI
- Location: `~/.config/todo/logs/todo.log`
- Format: JSON with timestamp, level, message
- Rotation: Auto-rotates at 10MB to `todo.log.old`
- Levels: debug, info, warn, error (configurable in config)

### Server
- Uses Go's standard log package
- Output to stdout/stderr

## Sync Behavior

**Background Sync (CLI):**
- Starts automatically if sync is enabled
- Runs at intervals specified in config (default 60s)
- Auto-syncs on local changes if `auto_sync_on_change: true`
- Manual sync with 's' key

**Offline Behavior:**
- All apps work offline with local database
- Changes queue until connectivity restored
- No data loss on disconnect

## Common Tasks

### Enable Sync on CLI

Edit `~/.config/todo/config`:
```json
{
  "db_path": "./todo.db",
  "log_level": "info",
  "sync": {
    "enabled": true,
    "server_url": "http://localhost:8080/api/v1",
    "api_key": "your-api-key",
    "sync_interval_seconds": 30,
    "auto_sync_on_change": true
  }
}
```

### Test Server with CLI Client

1. Start PostgreSQL
2. Start server (see Server Application section)
3. Create API key in database: `INSERT INTO users (api_key) VALUES ('test-key');`
4. Update CLI config with sync enabled
5. Watch logs: `tail -f ~/.config/todo/logs/todo.log`

### Debug Sync Issues

1. Check CLI logs at `~/.config/todo/logs/todo.log`
2. Set `log_level: "debug"` for verbose output
3. Verify server is reachable: `curl http://localhost:8080/api/v1/health`
4. Check API key is correct in config
5. Review server logs (stdout/stderr)
