/**
 * SQLite database schema
 *
 * Defines all tables and indexes for local offline storage.
 * Schema matches the server PostgreSQL database structure.
 */

export const DATABASE_NAME = 'commandlinetodo.db';
export const DATABASE_VERSION = 1;

export const DATABASE_SCHEMA = `
-- Todo Lists Table
CREATE TABLE IF NOT EXISTS todo_lists (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  client_id TEXT UNIQUE NOT NULL,
  server_id INTEGER DEFAULT 0,
  name TEXT NOT NULL,
  display_order INTEGER DEFAULT 0,
  archived INTEGER DEFAULT 0,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL,
  version INTEGER DEFAULT 1
);

-- Tasks Table
CREATE TABLE IF NOT EXISTS tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  client_id TEXT UNIQUE NOT NULL,
  server_id INTEGER DEFAULT 0,
  todo_list_id INTEGER NOT NULL,
  todo TEXT NOT NULL,
  priority INTEGER DEFAULT 4,
  done INTEGER DEFAULT 0,
  date_added INTEGER NOT NULL,
  date_completed INTEGER DEFAULT 0,
  due_date INTEGER DEFAULT 0,
  deleted INTEGER DEFAULT 0,
  deleted_at INTEGER DEFAULT 0,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL,
  version INTEGER DEFAULT 1,
  FOREIGN KEY (todo_list_id) REFERENCES todo_lists(id) ON DELETE CASCADE
);

-- Sync Metadata Table
CREATE TABLE IF NOT EXISTS sync_metadata (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

-- Change Log Table (tracks pending changes for push)
CREATE TABLE IF NOT EXISTS change_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  entity_type TEXT NOT NULL,
  entity_id INTEGER NOT NULL,
  change_type TEXT NOT NULL,
  timestamp INTEGER NOT NULL,
  synced INTEGER DEFAULT 0
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_tasks_list_id ON tasks(todo_list_id);
CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX IF NOT EXISTS idx_tasks_client_id ON tasks(client_id);
CREATE INDEX IF NOT EXISTS idx_lists_updated_at ON todo_lists(updated_at);
CREATE INDEX IF NOT EXISTS idx_lists_client_id ON todo_lists(client_id);
CREATE INDEX IF NOT EXISTS idx_change_log_synced ON change_log(synced);
`;

// Also export as SCHEMA for backward compatibility
export const SCHEMA = DATABASE_SCHEMA;
