package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	_ "modernc.org/sqlite"
)

func main() {
	// Create temp database
	dbPath := filepath.Join(os.TempDir(), "test_todo_fresh.db")
	defer os.Remove(dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create tables (mimicking the client app)
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS todoLists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		display_order INTEGER DEFAULT 0,
		archived BOOLEAN DEFAULT 0,
		created_at INTEGER,
		updated_at INTEGER,
		client_id TEXT,
		server_id INTEGER DEFAULT 0,
		version INTEGER DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS tasks (
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
		client_id TEXT,
		server_id INTEGER DEFAULT 0,
		version INTEGER DEFAULT 1,
		list_client_id TEXT,
		FOREIGN KEY (todoList_id) REFERENCES todoLists(id)
	);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	fmt.Println("✓ Created tables successfully")

	// Check list_client_id column exists
	rows, err := db.Query("PRAGMA table_info(tasks)")
	if err != nil {
		log.Fatalf("Failed to check schema: %v", err)
	}
	defer rows.Close()

	hasListClientID := false
	for rows.Next() {
		var cid, name, type_, notnull, dfltvalue, pk interface{}
		if err := rows.Scan(&cid, &name, &type_, &notnull, &dfltvalue, &pk); err != nil {
			log.Fatalf("Failed to scan column: %v", err)
		}
		if name == "list_client_id" {
			hasListClientID = true
			fmt.Printf("✓ Found list_client_id column: type=%v\n", type_)
		}
	}

	if !hasListClientID {
		log.Fatal("✗ list_client_id column not found!")
	}

	// Test inserting data
	now := int64(1704067200) // 2024-01-01

	// Insert list
	result, err := db.Exec(
		"INSERT INTO todoLists (name, display_order, created_at, updated_at, client_id, server_id, version) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"Work", 0, now, now, "list-uuid-001", 0, 1,
	)
	if err != nil {
		log.Fatalf("Failed to insert list: %v", err)
	}

	listID, _ := result.LastInsertId()
	fmt.Printf("✓ Created list: id=%d, clientID=list-uuid-001\n", listID)

	// Insert task with listClientID
	result, err = db.Exec(
		"INSERT INTO tasks (todo, priority, done, dateAdded, dueDate, deleted, todoList_id, client_id, list_client_id, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"Buy groceries", 2, 0, now, 0, 0, listID, "task-uuid-001", "list-uuid-001", 1,
	)
	if err != nil {
		log.Fatalf("Failed to insert task: %v", err)
	}

	taskID, _ := result.LastInsertId()
	fmt.Printf("✓ Created task: id=%d, clientID=task-uuid-001, listClientID=list-uuid-001\n", taskID)

	// Verify the data
	var taskTodo, taskListClientID string
	var taskListID int
	err = db.QueryRow(
		"SELECT todo, todoList_id, list_client_id FROM tasks WHERE id = ?",
		taskID,
	).Scan(&taskTodo, &taskListID, &taskListClientID)

	if err != nil {
		log.Fatalf("Failed to query task: %v", err)
	}

	fmt.Printf("✓ Verified task: todo='%s', todoListID=%d, listClientID='%s'\n", taskTodo, taskListID, taskListClientID)

	if taskListClientID != "list-uuid-001" {
		log.Fatalf("✗ listClientID mismatch: expected 'list-uuid-001', got '%s'", taskListClientID)
	}

	// Test populating task with list's clientID
	_, err = db.Exec(`
		UPDATE tasks
		SET list_client_id = (
		    SELECT client_id FROM todoLists WHERE id = tasks.todoList_id
		)
		WHERE list_client_id IS NULL OR list_client_id = ''
	`)
	if err != nil {
		log.Fatalf("Failed to populate list_client_id: %v", err)
	}

	fmt.Println("✓ populateTaskListClientIDs migration works")

	// Test with task without listClientID
	result, err = db.Exec(
		"INSERT INTO tasks (todo, priority, done, dateAdded, dueDate, deleted, todoList_id, client_id, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"Task without list_client_id", 3, 0, now, 0, 0, listID, "task-uuid-002", 1,
	)
	if err != nil {
		log.Fatalf("Failed to insert task 2: %v", err)
	}

	// Run migration
	_, err = db.Exec(`
		UPDATE tasks
		SET list_client_id = (
		    SELECT client_id FROM todoLists WHERE id = tasks.todoList_id
		)
		WHERE list_client_id IS NULL OR list_client_id = ''
	`)
	if err != nil {
		log.Fatalf("Failed to populate list_client_id: %v", err)
	}

	// Verify
	err = db.QueryRow(
		"SELECT list_client_id FROM tasks WHERE client_id = 'task-uuid-002'",
	).Scan(&taskListClientID)

	if err != nil {
		log.Fatalf("Failed to query task 2: %v", err)
	}

	if taskListClientID != "list-uuid-001" {
		log.Fatalf("✗ Migration failed: expected 'list-uuid-001', got '%s'", taskListClientID)
	}

	fmt.Printf("✓ Migration populated task 2: listClientID='%s'\n", taskListClientID)
	fmt.Println("\n✅ All database schema tests passed!")
}
