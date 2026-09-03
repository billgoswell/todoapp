# CommandLineTodo

A multi-platform todo application with a command-line interface, REST API server, and mobile app—all built from a unified monorepo architecture.

## Overview

This monorepo contains three integrated components for managing tasks and todo lists:

- **CLI** (`todo-cmdline/`) - A Go-based command-line interface using Bubble Tea TUI framework
- **Server** (`todo-server/`) - A Go REST API server with PostgreSQL database using Gin framework
- **Mobile** (`todo-mobile/`) - A React Native application for iOS and Android

All components share canonical type definitions from `shared-types/` and are tested with cross-component integration tests.

## Project Status

- ✅ **335+ tests passing** across all components (126 CLI, 128 server, 66 mobile, 15 integration)
- ✅ **CLI**: Fully functional with task management, list organization, and local storage
- ✅ **Server**: REST API with PostgreSQL, optimized batch operations, sync functionality
- ✅ **Mobile**: React Native app with complete feature parity
- ✅ **Integration Tests**: Cross-component test suite validating interactions (requires a running server + database)

## Repository Structure

```
todoapp/
├── todo-cmdline/          # CLI application (Go)
│   ├── cmd/app/          # Main CLI application
│   ├── internal/         # Handler, storage, sync logic
│   └── README.md         # CLI-specific documentation
│
├── todo-server/          # REST API server (Go + PostgreSQL)
│   ├── internal/         # Handlers, database, models
│   ├── TESTING.md        # Server API testing guide
│   └── README.md         # Server-specific documentation
│
├── todo-mobile/          # React Native mobile app
│   ├── src/              # React components and logic
│   ├── DEVELOPER_GUIDE.md # Mobile development guide
│   └── README.md         # Mobile-specific documentation
│
├── shared-types/         # Canonical type definitions
│   └── README.md         # Shared types documentation
│
├── integration-tests/    # Cross-component test suite
│   └── README.md         # Integration test documentation
│
├── docs/                 # Project documentation
│   ├── CLAUDE.md         # AI development guidelines
│   ├── PLAN.md           # Project roadmap
│   ├── STATUS.md         # Current project status
│   ├── REPO_STRUCTURE.md # Detailed repository structure
│   └── ...
│
└── .gitignore            # Standard exclusions
```

## Quick Start

### Prerequisites

- **Go 1.21+** - For CLI and Server development
- **Node.js 18+** - For Mobile development
- **PostgreSQL 15+** - For Server database (if running locally)
- **Git** - For version control

### CLI Development

```bash
cd todo-cmdline
go test ./cmd/app -v
go run cmd/app/main.go
```

See `todo-cmdline/README.md` for detailed CLI documentation.

### Server Development

```bash
cd todo-server
# Start PostgreSQL (or use existing instance)
go test ./... -v
go run main.go
```

See `todo-server/README.md` and `todo-server/TESTING.md` for server documentation.

### Mobile Development

```bash
cd todo-mobile
npm install
npm start
```

See `todo-mobile/README.md` and `todo-mobile/DEVELOPER_GUIDE.md` for mobile documentation.

### Integration Tests

```bash
cd integration-tests
go test ./... -v
```

See `integration-tests/README.md` for integration test documentation.

## Testing

This monorepo includes comprehensive test coverage:

- **CLI Tests**: Handler tests, storage tests, sync tests (126 tests)
- **Server Tests**: API endpoint tests, repository tests, sync/conversion tests (128 tests)
- **Mobile Tests**: Component, database, sync, and state tests (66 tests)
- **Integration Tests**: Cross-component synchronization tests (15 tests)

Run all tests locally by running the test command in each component directory.

> **Note**: `todo-server/internal/db` uses [testcontainers](https://testcontainers.com/) and needs a running Docker daemon; `integration-tests/` needs a running server and database (`localhost:8080`). Both are skipped/fail without those dependencies available.

## Documentation

Comprehensive documentation is available:

- **Development**: Check the README in each component directory
- **Architecture**: See `docs/PLAN.md` for the project roadmap
- **API Testing**: See `todo-server/TESTING.md` for REST API testing guide
- **Mobile Development**: See `todo-mobile/DEVELOPER_GUIDE.md`
- **CLI Configuration**: See `todo-cmdline/CONFIG.md`
- **Deployment**: See `todo-server/DEPLOYMENT_PLAN.md`

## Development Workflow

1. **Create a feature branch** from `main`
2. **Make your changes** in the relevant component directory
3. **Run tests** to ensure nothing breaks
4. **Commit and push** your changes
5. **Merge to main** once you're satisfied with the changes

## Key Features

### CLI Application
- Interactive task management
- List organization and filtering
- Local storage with configurable paths
- Sync with server for multi-device access
- Keyboard-driven interface with Bubble Tea

### Server
- RESTful API for all task operations
- PostgreSQL database with optimized batch operations
- User authentication and task synchronization
- Efficient query patterns and indexing
- Docker containerization support

### Mobile App
- Native-like experience on iOS and Android
- Full feature parity with CLI
- Real-time synchronization with server
- Offline support with local caching
- Push notifications

## Architecture Highlights

- **Monorepo**: Single Git repository for atomic commits across components
- **Shared Types**: Canonical Go type definitions prevent data inconsistencies
- **Last-Write-Wins Conflict Resolution**: Simple, predictable sync behavior
- **Batch Operations**: PostgreSQL multi-row inserts for 100x-1000x performance gains
- **Comprehensive Testing**: Unit, integration, and cross-component tests

## Contributing

For personal development:

1. All changes should have corresponding tests
2. Run full test suite before committing
3. Update documentation if behavior changes
4. Keep components loosely coupled through shared types

## License

This project is provided as-is for personal use.

## Getting Help

- **CLI Issues**: Check `todo-cmdline/README.md` and run tests
- **Server Issues**: Check `todo-server/TESTING.md` and database logs
- **Mobile Issues**: Check `todo-mobile/DEVELOPER_GUIDE.md`
- **Integration Issues**: Run integration tests in `integration-tests/`

## Quick Reference

| Component | Language | Framework | Tests | Status |
|-----------|----------|-----------|-------|--------|
| CLI | Go | Bubble Tea | 126 | ✅ Complete |
| Server | Go | Gin + PostgreSQL | 128 | ✅ Complete |
| Mobile | TypeScript/React | React Native | 66 | ✅ Complete |
| Integration | Go | Testify | 15 | ✅ Complete |

---

**Last Updated**: 2026-09-03
**Monorepo Established**: 2026-01-22
**Project Status**: All phases complete, ready for active development
