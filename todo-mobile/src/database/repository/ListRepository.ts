/**
 * List Repository
 *
 * Handles all CRUD operations for todo lists in SQLite database.
 * Implements repository pattern for data access abstraction.
 * Uses expo-sqlite for Expo compatibility.
 */

import { TodoList } from '../../api/types';
import type { SQLiteDatabase } from 'expo-sqlite';

export class ListRepository {
  constructor(private db: SQLiteDatabase) {}

  /**
   * Get all lists (not archived)
   */
  async findAll(): Promise<TodoList[]> {
    const rows = this.db.getAllSync<any>(
      'SELECT * FROM todo_lists WHERE archived = 0 ORDER BY display_order ASC, created_at DESC'
    );
    return rows.map(row => this.mapRowToList(row));
  }

  /**
   * Get all lists including archived
   */
  async findAllIncludingArchived(): Promise<TodoList[]> {
    const rows = this.db.getAllSync<any>(
      'SELECT * FROM todo_lists ORDER BY display_order ASC, created_at DESC'
    );
    return rows.map(row => this.mapRowToList(row));
  }

  /**
   * Get list by ID
   */
  async findById(id: number): Promise<TodoList | null> {
    const row = this.db.getFirstSync<any>(
      'SELECT * FROM todo_lists WHERE id = ?',
      [id]
    );
    return row ? this.mapRowToList(row) : null;
  }

  /**
   * Get list by client ID
   */
  async findByClientId(clientId: string): Promise<TodoList | null> {
    const row = this.db.getFirstSync<any>(
      'SELECT * FROM todo_lists WHERE client_id = ?',
      [clientId]
    );
    return row ? this.mapRowToList(row) : null;
  }

  /**
   * Get lists modified since timestamp (for sync)
   */
  async findModifiedSince(timestamp: number): Promise<TodoList[]> {
    const rows = this.db.getAllSync<any>(
      'SELECT * FROM todo_lists WHERE updated_at > ? ORDER BY updated_at DESC',
      [timestamp]
    );
    return rows.map(row => this.mapRowToList(row));
  }

  /**
   * Get archived lists
   */
  async findArchived(): Promise<TodoList[]> {
    const rows = this.db.getAllSync<any>(
      'SELECT * FROM todo_lists WHERE archived = 1 ORDER BY created_at DESC'
    );
    return rows.map(row => this.mapRowToList(row));
  }

  /**
   * Create a new list
   */
  async create(list: TodoList): Promise<TodoList> {
    const now = Math.floor(Date.now() / 1000);
    const result = this.db.runSync(
      `INSERT INTO todo_lists
       (client_id, server_id, name, display_order, archived, created_at, updated_at, version)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        list.client_id,
        list.id ?? 0,
        list.name,
        list.display_order ?? 0,
        list.archived ? 1 : 0,
        list.created_at ?? now,
        list.updated_at ?? now,
        list.version ?? 1,
      ]
    );

    return {
      ...list,
      id: result.lastInsertRowId,
      created_at: list.created_at ?? now,
      updated_at: list.updated_at ?? now,
    };
  }

  /**
   * Update an existing list
   */
  async update(list: TodoList): Promise<TodoList> {
    const now = Math.floor(Date.now() / 1000);
    this.db.runSync(
      `UPDATE todo_lists SET
       name = ?, display_order = ?, archived = ?, updated_at = ?, version = version + 1
       WHERE id = ?`,
      [
        list.name,
        list.display_order ?? 0,
        list.archived ? 1 : 0,
        now,
        list.id ?? null,
      ]
    );

    return {
      ...list,
      updated_at: now,
      version: (list.version ?? 1) + 1,
    };
  }

  /**
   * Archive a list (soft delete)
   */
  async archive(id: number): Promise<void> {
    const now = Math.floor(Date.now() / 1000);
    this.db.runSync(
      'UPDATE todo_lists SET archived = 1, updated_at = ? WHERE id = ?',
      [now, id]
    );
  }

  /**
   * Unarchive a list
   */
  async unarchive(id: number): Promise<void> {
    const now = Math.floor(Date.now() / 1000);
    this.db.runSync(
      'UPDATE todo_lists SET archived = 0, updated_at = ? WHERE id = ?',
      [now, id]
    );
  }

  /**
   * Hard delete a list (permanent deletion)
   * Should cascade delete tasks
   */
  async hardDelete(id: number): Promise<void> {
    this.db.withTransactionSync(() => {
      // Delete all tasks in the list first
      this.db.runSync('DELETE FROM tasks WHERE todo_list_id = ?', [id]);
      // Then delete the list
      this.db.runSync('DELETE FROM todo_lists WHERE id = ?', [id]);
    });
  }

  /**
   * Count total lists (not archived)
   */
  async count(): Promise<number> {
    const row = this.db.getFirstSync<{ count: number }>(
      'SELECT COUNT(*) as count FROM todo_lists WHERE archived = 0'
    );
    return row?.count ?? 0;
  }

  /**
   * Reorder lists
   */
  async reorder(lists: TodoList[]): Promise<void> {
    const now = Math.floor(Date.now() / 1000);
    this.db.withTransactionSync(() => {
      lists.forEach((list, index) => {
        this.db.runSync(
          'UPDATE todo_lists SET display_order = ?, updated_at = ? WHERE id = ?',
          [index, now, list.id ?? null]
        );
      });
    });
  }

  /**
   * Bulk update lists (for sync operations)
   */
  async bulkUpdate(lists: TodoList[]): Promise<void> {
    this.db.withTransactionSync(() => {
      for (const list of lists) {
        this.db.runSync(
          `UPDATE todo_lists SET
           name = ?, display_order = ?, archived = ?, updated_at = ?, version = ?,
           server_id = COALESCE(?, server_id)
           WHERE client_id = ?`,
          [
            list.name,
            list.display_order ?? 0,
            list.archived ? 1 : 0,
            list.updated_at,
            list.version ?? 1,
            list.id ?? null,
            list.client_id,
          ]
        );
      }
    });
  }

  /**
   * Bulk insert lists (for sync operations)
   */
  async bulkInsert(lists: TodoList[]): Promise<void> {
    this.db.withTransactionSync(() => {
      for (const list of lists) {
        this.db.runSync(
          `INSERT OR IGNORE INTO todo_lists
           (client_id, server_id, name, display_order, archived, created_at, updated_at, version)
           VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
          [
            list.client_id,
            list.id ?? 0,
            list.name,
            list.display_order ?? 0,
            list.archived ? 1 : 0,
            list.created_at ?? Math.floor(Date.now() / 1000),
            list.updated_at ?? Math.floor(Date.now() / 1000),
            list.version ?? 1,
          ]
        );
      }
    });
  }

  /**
   * Map database row to TodoList object
   */
  private mapRowToList(row: any): TodoList {
    return {
      id: row.server_id || row.id,
      client_id: row.client_id,
      name: row.name,
      display_order: row.display_order ?? 0,
      archived: Boolean(row.archived),
      created_at: row.created_at,
      updated_at: row.updated_at,
      version: row.version,
    };
  }
}
