package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/billgoswell/commandlinetodo/internal/config"
	"github.com/billgoswell/commandlinetodo/internal/logger"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize logger before anything else
	configDir, err := config.GetConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to determine config directory: %v\n", err)
		os.Exit(1)
	}

	if err := logger.InitLogger(configDir, logger.LogLevel(cfg.LogLevel)); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.LogInfo("Application starting")

	db, err := initDB(cfg.DBPath)
	if err != nil {
		logger.LogFatal("Failed to initialize database", "error", err)
	}
	defer db.Close()

	logger.LogInfo("Database initialized", "path", cfg.DBPath)

	// Create data store (local or with sync)
	var store DataStore
	var syncStore *SyncStore
	localStore := NewLocalStore(db)

	if cfg.Sync.Enabled {
		logger.LogInfo("Sync enabled, initializing sync client")
		// Create sync client
		syncClient := NewSyncClient(cfg.Sync)

		// Create sync store
		syncStore = NewSyncStore(localStore, syncClient, cfg.Sync)

		// Perform initial sync if online
		if syncClient.IsOnline() {
			logger.LogInfo("Server is online, performing initial sync")
			if err := syncStore.FullSync(); err != nil {
				logger.LogWarn("Initial sync failed", "error", err)
			} else {
				logger.LogInfo("Initial sync completed successfully")
			}
		} else {
			logger.LogInfo("Server is offline, using local mode")
		}

		// Start background sync
		syncStore.StartBackgroundSync()
		logger.LogInfo("Background sync started", "interval", cfg.Sync.SyncIntervalSeconds)

		store = syncStore
	} else {
		logger.LogInfo("Sync disabled, using local-only mode")
		store = localStore
	}

	// Load data using the store interface
	todoLists, err := store.GetTodoLists()
	if err != nil {
		logger.LogFatal("Failed to load todo lists", "error", err)
	}

	if len(todoLists) == 0 {
		id, err := store.CreateTodoList(DefaultListName)
		if err != nil {
			logger.LogFatal("Failed to create default todo list", "error", err)
		}
		todoLists = []todoList{{id: id, name: DefaultListName, archived: false}}
		logger.LogInfo("Created default todo list")
	}

	todoItems, err := store.GetItems()
	if err != nil {
		logger.LogFatal("Failed to load items from database", "error", err)
	}

	logger.LogInfo("Data loaded successfully", "lists", len(todoLists), "items", len(todoItems))

	// Create model with store
	m := initialModel(todoItems, todoLists)
	m.store = store
	m.syncEnabled = cfg.Sync.Enabled
	if cfg.Sync.Enabled && syncStore != nil {
		m.syncStatus.online = syncStore.client.IsOnline()
		m.syncStatus.lastSyncTime = 0 // Will be set during first sync
	}
	m.sortItems()

	logger.LogInfo("Application started successfully")

	// Run the TUI
	p := tea.NewProgram(m, tea.WithAltScreen())
	defer func() {
		if syncStore != nil {
			syncStore.StopBackgroundSync()
			logger.LogInfo("Background sync stopped")
		}
		logger.LogInfo("Application shutdown")
	}()

	if _, err := p.Run(); err != nil {
		logger.LogFatal("Error starting TUI", "error", err)
	}
}
