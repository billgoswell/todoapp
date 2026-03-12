package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/billgoswell/commandlinetodo-server/internal/db/repositories"
)

// IDMapping is a client-to-server ID mapping returned from batch operations
type IDMapping = repositories.IDMapping

// DB wraps the database connection and provides access to repositories
type DB struct {
	conn *sql.DB

	// Repositories for different entities
	Users *repositories.UserRepository
	Tasks *repositories.TaskRepository
	Lists *repositories.ListRepository
}

// NewDB creates a new database connection
func NewDB(host, port, user, password, dbname string) (*DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		return nil, err
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	return &DB{
		conn:  conn,
		Users: repositories.NewUserRepository(conn),
		Tasks: repositories.NewTaskRepository(conn),
		Lists: repositories.NewListRepository(conn),
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Ping checks if the database is reachable
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// Migrate runs database migrations
func (db *DB) Migrate() error {
	schema := `
	-- Create users table
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		api_key VARCHAR(64) UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create todo_lists table
	CREATE TABLE IF NOT EXISTS todo_lists (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		client_id VARCHAR(36) NOT NULL,
		name TEXT NOT NULL,
		display_order INTEGER DEFAULT 0,
		archived BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		version INTEGER DEFAULT 1,
		CONSTRAINT unique_list_per_user UNIQUE(user_id, name)
	);

	-- Create tasks table
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		client_id VARCHAR(36) NOT NULL,
		todo_list_id INTEGER NOT NULL REFERENCES todo_lists(id) ON DELETE CASCADE,
		todo_list_client_id VARCHAR(36),
		todo TEXT NOT NULL,
		priority INTEGER DEFAULT 4,
		done BOOLEAN DEFAULT FALSE,
		date_added TIMESTAMP NOT NULL,
		date_completed TIMESTAMP,
		due_date TIMESTAMP,
		deleted BOOLEAN DEFAULT FALSE,
		deleted_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		version INTEGER DEFAULT 1,
		CONSTRAINT unique_task_per_user UNIQUE(user_id, client_id)
	);

	-- Create indexes for common queries
	-- Single column indexes (maintained for backward compatibility)
	CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at);
	CREATE INDEX IF NOT EXISTS idx_tasks_todo_list_id ON tasks(todo_list_id);
	CREATE INDEX IF NOT EXISTS idx_lists_user_id ON todo_lists(user_id);
	CREATE INDEX IF NOT EXISTS idx_lists_updated_at ON todo_lists(updated_at);
	CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

	-- Composite indexes for sync queries (user_id + updated_at)
	-- These cover the WHERE clause in GetTasksSince and GetListsSince
	CREATE INDEX IF NOT EXISTS idx_tasks_user_updated ON tasks(user_id, updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_lists_user_updated ON todo_lists(user_id, updated_at DESC);

	-- Partial indexes for active (non-deleted/non-archived) items
	-- These smaller indexes are faster for queries filtering by deleted/archived status
	CREATE INDEX IF NOT EXISTS idx_tasks_active ON tasks(user_id, updated_at DESC)
		WHERE deleted = FALSE;
	CREATE INDEX IF NOT EXISTS idx_lists_active ON todo_lists(user_id, updated_at DESC)
		WHERE archived = FALSE;

	-- Index for task filtering by list and completion status
	-- Supports dashboard queries like "tasks in list X where done = Y"
	CREATE INDEX IF NOT EXISTS idx_tasks_by_list_status ON tasks(todo_list_id, done)
		WHERE deleted = FALSE;

	-- Index for resolving tasks by list client_id
	-- Supports sync operations that reference lists by UUID
	CREATE INDEX IF NOT EXISTS idx_tasks_list_client_id ON tasks(todo_list_client_id);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// GetListIDByClientID resolves a list's clientID to its server ID
// Used during sync to convert stable UUID references to database IDs
func (db *DB) GetListIDByClientID(userID int, clientID string) (int, error) {
	return db.Lists.GetIDByClientID(userID, clientID)
}

// CreateUser creates a new user with an API key (delegates to UserRepository)
func (db *DB) CreateUser(apiKey string) (int, error) {
	return db.Users.CreateUser(apiKey)
}

// GetUserByAPIKey retrieves a user by API key (delegates to UserRepository)
func (db *DB) GetUserByAPIKey(apiKey string) (int, error) {
	return db.Users.GetUserByAPIKey(apiKey)
}

// GetTasksSince retrieves all tasks modified since the given timestamp (delegates to TaskRepository)
func (db *DB) GetTasksSince(userID int, since int64) ([]map[string]interface{}, error) {
	return db.Tasks.GetTasksSince(userID, since)
}

// GetListsSince retrieves all lists modified since the given timestamp (delegates to ListRepository)
func (db *DB) GetListsSince(userID int, since int64) ([]map[string]interface{}, error) {
	return db.Lists.GetListsSince(userID, since)
}

// UpsertTask inserts or updates a task (delegates to TaskRepository)
func (db *DB) UpsertTask(userID int, task map[string]interface{}) (int, error) {
	return db.Tasks.UpsertTask(userID, task)
}

// UpsertList inserts or updates a list (delegates to ListRepository)
func (db *DB) UpsertList(userID int, list map[string]interface{}) (int, error) {
	return db.Lists.UpsertList(userID, list)
}

// UpsertTasksBatch inserts or updates multiple tasks in a single operation
// Returns ID mappings to update client's local database with server IDs
func (db *DB) UpsertTasksBatch(userID int, tasks []map[string]interface{}) ([]IDMapping, error) {
	return db.Tasks.UpsertTasksBatch(userID, tasks)
}

// UpsertListsBatch inserts or updates multiple lists in a single operation
// Returns ID mappings to update client's local database with server IDs
func (db *DB) UpsertListsBatch(userID int, lists []map[string]interface{}) ([]IDMapping, error) {
	return db.Lists.UpsertListsBatch(userID, lists)
}
