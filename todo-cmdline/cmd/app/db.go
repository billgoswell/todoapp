package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"time"

	"github.com/google/uuid"
	"github.com/billgoswell/commandlinetodo/internal/logger"
	"github.com/billgoswell/commandlinetodo/internal/storage"
)

type todoItem struct {
	id              int
	clientID        string // UUID for offline-first sync
	serverID        int    // Server's ID (0 if not synced yet)
	done            bool
	todo            string
	priority        int
	dateCompleted   int64
	dateAdded       int64
	dueDate         int64
	deleted         bool
	deletedAt       int64
	todoListID      int
	listClientID    string // UUID reference to the list (for sync stability)
	updatedAt       int64  // Timestamp of last update (for sync conflict resolution)
	version         int    // For conflict detection
}

var db *sql.DB

func now() int64 {
	return time.Now().Unix()
}

func executeStmt(operation string, query string, args ...interface{}) error {
	return storage.ExecuteWithError(db, operation, query, args...)
}

func executeStmtWithID(operation string, query string, args ...interface{}) (int, error) {
	return storage.ExecuteWithIDError(db, operation, query, args...)
}

func initDB(dbPath string) (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		logger.LogError("Failed to open database", "path", dbPath, "error", err)
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if err := executeStmt("enable foreign keys", "PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	if err := createTableIfNotExists(); err != nil {
		logger.LogError("Failed to create tables", "error", err)
		return nil, err
	}

	if err := createSyncTables(); err != nil {
		logger.LogError("Failed to create sync tables", "error", err)
		return nil, err
	}

	if err := migrateSyncColumns(); err != nil {
		logger.LogError("Failed to migrate sync columns", "error", err)
		return nil, err
	}

	if err := populateTaskListClientIDs(); err != nil {
		logger.LogWarn("Failed to populate task list client IDs", "error", err)
	}

	if err := fixExistingTaskListIDs(); err != nil {
		logger.LogWarn("Failed to fix task list IDs", "error", err)
	}

	return db, nil
}

func createTableIfNotExists() error {
	if err := executeStmt("create todoLists table", `CREATE TABLE IF NOT EXISTS todoLists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		display_order INTEGER DEFAULT 0,
		archived BOOLEAN DEFAULT 0,
		created_at INTEGER,
		updated_at INTEGER
	)`); err != nil {
		return err
	}

	return executeStmt("create tasks table", `CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		todo TEXT NOT NULL,
		priority INTEGER DEFAULT 4,
		done BOOLEAN DEFAULT 0,
		dateAdded INTEGER,
		dateCompleted INTEGER DEFAULT 0,
		dueDate INTEGER DEFAULT 0,
		deleted BOOLEAN DEFAULT 0,
		deletedAt INTEGER DEFAULT 0,
		todoList_id INTEGER DEFAULT 1,
		FOREIGN KEY (todoList_id) REFERENCES todoLists(id)
	)`)
}

func getItemsFromDB() ([]todoItem, error) {
	rows, err := storage.QueryWithError(db, "query items",
		"SELECT id, todo, priority, done, dateAdded, dateCompleted, dueDate, deleted, deletedAt, todoList_id, COALESCE(client_id, ''), COALESCE(server_id, 0), COALESCE(version, 1), COALESCE(list_client_id, '') FROM tasks WHERE deleted = 0 ORDER BY id")
	if err != nil {
		return []todoItem{}, err
	}
	defer rows.Close()

	items := []todoItem{}
	for rows.Next() {
		var item todoItem
		if err := rows.Scan(&item.id, &item.todo, &item.priority, &item.done, &item.dateAdded, &item.dateCompleted, &item.dueDate, &item.deleted, &item.deletedAt, &item.todoListID, &item.clientID, &item.serverID, &item.version, &item.listClientID); err != nil {
			logger.LogError("Failed to scan item", "error", err)
			return []todoItem{}, err
		}
		// Validate priority to ensure it's a valid value (1-4)
		item.priority = validatePriority(item.priority)
		// Generate client ID if missing (for backward compatibility)
		if item.clientID == "" {
			item.clientID = generateClientID()
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		logger.LogError("Error iterating items", "error", err)
		return []todoItem{}, err
	}

	return items, nil
}

func saveItemToDB(item todoItem) error {
	return executeStmt("insert item",
		"INSERT INTO tasks (todo, priority, done, dateAdded, dueDate, deleted, todoList_id, client_id, list_client_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		item.todo, item.priority, item.done, now(), item.dueDate, 0, item.todoListID, item.clientID, item.listClientID,
	)
}

func updateItemInDB(item todoItem) error {
	return executeStmt("update item",
		"UPDATE tasks SET todo = ?, done = ?, priority = ?, dateCompleted = ?, dueDate = ?, todoList_id = ?, list_client_id = ? WHERE id = ?",
		item.todo, item.done, item.priority, item.dateCompleted, item.dueDate, item.todoListID, item.listClientID, item.id,
	)
}

func markItemAsDeleted(id int) error {
	return executeStmt("delete item",
		"UPDATE tasks SET deleted = 1, deletedAt = ? WHERE id = ?",
		now(), id,
	)
}

func getTodoLists() ([]todoList, error) {
	rows, err := storage.QueryWithError(db, "query todoLists",
		"SELECT id, name, display_order, archived, created_at, updated_at, COALESCE(client_id, ''), COALESCE(server_id, 0), COALESCE(version, 1) FROM todoLists WHERE archived = 0 ORDER BY display_order")
	if err != nil {
		return []todoList{}, err
	}
	defer rows.Close()

	lists := []todoList{}
	for rows.Next() {
		var list todoList
		if err := rows.Scan(&list.id, &list.name, &list.displayOrder, &list.archived, &list.createdAt, &list.updatedAt, &list.clientID, &list.serverID, &list.version); err != nil {
			logger.LogError("Failed to scan todoList", "error", err)
			return []todoList{}, err
		}
		// Generate client ID if missing (for backward compatibility)
		if list.clientID == "" {
			list.clientID = generateClientID()
		}
		lists = append(lists, list)
	}
	if err := rows.Err(); err != nil {
		logger.LogError("Error iterating todoLists", "error", err)
		return []todoList{}, err
	}

	return lists, nil
}

func getTodoListByID(id int) (todoList, error) {
	rows, err := storage.QueryWithError(db, "query todoList by id",
		"SELECT id, name, display_order, archived, created_at, updated_at, COALESCE(client_id, ''), COALESCE(server_id, 0), COALESCE(version, 1) FROM todoLists WHERE id = ? LIMIT 1", id)
	if err != nil {
		return todoList{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var list todoList
		if err := rows.Scan(&list.id, &list.name, &list.displayOrder, &list.archived, &list.createdAt, &list.updatedAt, &list.clientID, &list.serverID, &list.version); err != nil {
			logger.LogError("Failed to scan todoList", "error", err)
			return todoList{}, err
		}
		// Generate client ID if missing (for backward compatibility)
		if list.clientID == "" {
			list.clientID = generateClientID()
		}
		return list, nil
	}

	return todoList{}, fmt.Errorf("list not found with id: %d", id)
}

func createTodoList(name string) (int, error) {
	return executeStmtWithID("create todo list",
		"INSERT INTO todoLists (name, display_order, archived, created_at, updated_at) VALUES (?, (SELECT COUNT(*) FROM todoLists), 0, ?, ?)",
		name, now(), now(),
	)
}

func updateTodoListName(id int, name string) error {
	return executeStmt("update todo list name",
		"UPDATE todoLists SET name = ?, updated_at = ? WHERE id = ?",
		name, now(), id,
	)
}

func deleteTodoList(id int) error {
	tx, err := storage.BeginWithError(db, "delete todo list")
	if err != nil {
		return err
	}
	defer storage.RollbackWithError(tx, "delete todo list")

	timestamp := now()

	if err := storage.TxExecuteWithError(tx, "archive todo list",
		"UPDATE todoLists SET archived = 1, updated_at = ? WHERE id = ?",
		timestamp, id,
	); err != nil {
		return err
	}

	if err := storage.TxExecuteWithError(tx, "delete tasks in list",
		"UPDATE tasks SET deleted = 1, deletedAt = ? WHERE todoList_id = ? AND deleted = 0",
		timestamp, id,
	); err != nil {
		return err
	}

	return storage.CommitWithError(tx, "delete todo list")
}

func setTodoListArchived(id int, archived bool) error {
	operation := "archive todo list"
	if !archived {
		operation = "unarchive todo list"
	}

	return executeStmt(operation,
		"UPDATE todoLists SET archived = ?, updated_at = ? WHERE id = ?",
		archived, now(), id,
	)
}

func archiveTodoList(id int) error {
	return setTodoListArchived(id, true)
}

func unarchiveTodoList(id int) error {
	return setTodoListArchived(id, false)
}

func fixExistingTaskListIDs() error {
	if err := executeStmt("fix task list IDs",
		"UPDATE tasks SET todoList_id = 1 WHERE todoList_id IS NULL OR todoList_id = 0",
	); err != nil {
		return err
	}

	return verifyTaskListIDs()
}

func verifyTaskListIDs() error {
	var orphanedCount int
	err := storage.QueryRowWithError(db, "verify task list IDs",
		"SELECT COUNT(*) FROM tasks WHERE deleted = 0 AND (todoList_id IS NULL OR todoList_id = 0)").Scan(&orphanedCount)
	if err != nil {
		logger.LogError("Failed to verify task list IDs", "error", err)
		return err
	}

	if orphanedCount > 0 {
		return fmt.Errorf("data integrity error: found %d orphaned tasks with invalid list IDs", orphanedCount)
	}

	return nil
}

// Sync infrastructure functions

func generateClientID() string {
	return uuid.New().String()
}

func createSyncTables() error {
	// Create sync metadata table
	if err := executeStmt("create sync_metadata table", `CREATE TABLE IF NOT EXISTS sync_metadata (
		key TEXT PRIMARY KEY,
		value TEXT
	)`); err != nil {
		return err
	}

	// Create change log table
	return executeStmt("create change_log table", `CREATE TABLE IF NOT EXISTS change_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		change_type TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		synced BOOLEAN DEFAULT 0
	)`)
}

func columnExists(tableName, columnName string) (bool, error) {
	var name string
	err := storage.QueryRowWithError(db, "check column exists",
		"SELECT name FROM pragma_table_info(?) WHERE name = ?",
		tableName, columnName,
	).Scan(&name)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func migrateSyncColumns() error {
	// Check and add sync columns to todoLists
	if exists, err := columnExists("todoLists", "client_id"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add client_id to todoLists",
			"ALTER TABLE todoLists ADD COLUMN client_id TEXT"); err != nil {
			return err
		}
	}

	if exists, err := columnExists("todoLists", "server_id"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add server_id to todoLists",
			"ALTER TABLE todoLists ADD COLUMN server_id INTEGER DEFAULT 0"); err != nil {
			return err
		}
	}

	if exists, err := columnExists("todoLists", "version"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add version to todoLists",
			"ALTER TABLE todoLists ADD COLUMN version INTEGER DEFAULT 1"); err != nil {
			return err
		}
	}

	// Check and add sync columns to tasks
	if exists, err := columnExists("tasks", "client_id"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add client_id to tasks",
			"ALTER TABLE tasks ADD COLUMN client_id TEXT"); err != nil {
			return err
		}
	}

	if exists, err := columnExists("tasks", "server_id"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add server_id to tasks",
			"ALTER TABLE tasks ADD COLUMN server_id INTEGER DEFAULT 0"); err != nil {
			return err
		}
	}

	if exists, err := columnExists("tasks", "version"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add version to tasks",
			"ALTER TABLE tasks ADD COLUMN version INTEGER DEFAULT 1"); err != nil {
			return err
		}
	}

	// Check and add list_client_id column to tasks
	if exists, err := columnExists("tasks", "list_client_id"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add list_client_id to tasks",
			"ALTER TABLE tasks ADD COLUMN list_client_id TEXT"); err != nil {
			return err
		}
	}

	return nil
}

func populateTaskListClientIDs() error {
	// Update tasks to have their list's clientID
	return executeStmt("populate task list_client_id",
		`UPDATE tasks
		 SET list_client_id = (
		     SELECT client_id FROM todoLists WHERE id = tasks.todoList_id
		 )
		 WHERE list_client_id IS NULL OR list_client_id = ''`)
}

func logChange(entityType string, entityID int, changeType string) error {
	return executeStmt("log change",
		"INSERT INTO change_log (entity_type, entity_id, change_type, timestamp, synced) VALUES (?, ?, ?, ?, 0)",
		entityType, entityID, changeType, now(),
	)
}

func getPendingChanges() ([]Change, error) {
	rows, err := storage.QueryWithError(db, "query pending changes",
		"SELECT id, entity_type, entity_id, change_type, timestamp, synced FROM change_log WHERE synced = 0 ORDER BY timestamp")
	if err != nil {
		return []Change{}, err
	}
	defer rows.Close()

	changes := []Change{}
	for rows.Next() {
		var change Change
		if err := rows.Scan(&change.id, &change.entityType, &change.entityID, &change.changeType, &change.timestamp, &change.synced); err != nil {
			logger.LogError("Failed to scan change", "error", err)
			return []Change{}, err
		}
		changes = append(changes, change)
	}

	if err := rows.Err(); err != nil {
		logger.LogError("Error iterating changes", "error", err)
		return []Change{}, err
	}

	return changes, nil
}

func markChangeSynced(changeID int) error {
	return executeStmt("mark change synced",
		"UPDATE change_log SET synced = 1 WHERE id = ?",
		changeID,
	)
}

func getLastSyncTime() (int64, error) {
	var timestamp int64
	err := storage.QueryRowWithError(db, "get last sync time",
		"SELECT value FROM sync_metadata WHERE key = 'last_sync_time'").Scan(&timestamp)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		logger.LogError("Failed to get last sync time", "error", err)
		return 0, err
	}

	return timestamp, nil
}

func setLastSyncTime(timestamp int64) error {
	// Use INSERT OR REPLACE to handle both insert and update
	return executeStmt("set last sync time",
		"INSERT OR REPLACE INTO sync_metadata (key, value) VALUES ('last_sync_time', ?)",
		fmt.Sprintf("%d", timestamp),
	)
}

func getTodoListByClientID(clientID string) (todoList, error) {
	rows, err := storage.QueryWithError(db, "query todoList by client_id",
		"SELECT id, name, display_order, archived, created_at, updated_at, COALESCE(client_id, ''), COALESCE(server_id, 0), COALESCE(version, 1) FROM todoLists WHERE client_id = ? LIMIT 1", clientID)
	if err != nil {
		return todoList{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var list todoList
		if err := rows.Scan(&list.id, &list.name, &list.displayOrder, &list.archived, &list.createdAt, &list.updatedAt, &list.clientID, &list.serverID, &list.version); err != nil {
			logger.LogError("Failed to scan todoList by client_id", "error", err)
			return todoList{}, err
		}
		return list, nil
	}

	return todoList{}, fmt.Errorf("list not found with client_id: %s", clientID)
}

func saveTodoList(list todoList) error {
	return executeStmt("save todo list from server",
		"INSERT INTO todoLists (name, display_order, archived, created_at, updated_at, client_id, version) VALUES (?, ?, ?, ?, ?, ?, ?)",
		list.name, list.displayOrder, list.archived, list.createdAt, list.updatedAt, list.clientID, list.version,
	)
}

func updateTodoListFromServer(list todoList) error {
	return executeStmt("update todo list from server",
		"UPDATE todoLists SET name = ?, display_order = ?, archived = ?, updated_at = ?, version = ? WHERE client_id = ?",
		list.name, list.displayOrder, list.archived, list.updatedAt, list.version, list.clientID,
	)
}

// updateListServerID updates a list's server ID after successful sync to server
// This maps the client ID to the server-assigned ID
func updateListServerID(clientID string, serverID int) error {
	return executeStmt("update list server_id",
		"UPDATE todoLists SET server_id = ? WHERE client_id = ?",
		serverID, clientID,
	)
}

// updateTaskServerID updates a task's server ID after successful sync to server
// This maps the client ID to the server-assigned ID
func updateTaskServerID(clientID string, serverID int) error {
	return executeStmt("update task server_id",
		"UPDATE tasks SET server_id = ? WHERE client_id = ?",
		serverID, clientID,
	)
}
