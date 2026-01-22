/**
 * Task Repository
 *
 * Handles all CRUD operations for tasks in SQLite database.
 * Implements repository pattern for data access abstraction.
 * Uses expo-sqlite for Expo compatibility.
 */

import { Task } from '../../api/types';
import type { SQLiteDatabase } from 'expo-sqlite';

export class TaskRepository {
  constructor(private db: SQLiteDatabase) {}

  /**
   * Get all tasks
   */
  async findAll(): Promise<Task[]> {
    const rows = this.db.getAllSync<any>(
      'SELECT * FROM tasks WHERE deleted = 0 ORDER BY created_at DESC'
    );
    return rows.map(row => this.mapRowToTask(row));
  }

  /**
   * Get tasks for a specific list
   */
  async findByListId(listId: number): Promise<Task[]> {
    const rows = this.db.getAllSync<any>(
      'SELECT * FROM tasks WHERE todo_list_id = ? AND deleted = 0 ORDER BY created_at DESC',
      [listId]
    );
    return rows.map(row => this.mapRowToTask(row));
  }

  /**
   * Get task by ID
   */
  async findById(id: number): Promise<Task | null> {
    const row = this.db.getFirstSync<any>(
      'SELECT * FROM tasks WHERE id = ?',
      [id]
    );
    return row ? this.mapRowToTask(row) : null;
  }

  /**
   * Get task by client ID
   */
  async findByClientId(clientId: string): Promise<Task | null> {
    const row = this.db.getFirstSync<any>(
      'SELECT * FROM tasks WHERE client_id = ?',
      [clientId]
    );
    return row ? this.mapRowToTask(row) : null;
  }

  /**
   * Get tasks modified since timestamp (for sync)
   */
  async findModifiedSince(timestamp: number): Promise<Task[]> {
    const rows = this.db.getAllSync<any>(
      'SELECT * FROM tasks WHERE updated_at > ? ORDER BY updated_at DESC',
      [timestamp]
    );
    return rows.map(row => this.mapRowToTask(row));
  }

  /**
   * Get soft-deleted tasks
   */
  async findDeleted(): Promise<Task[]> {
    const rows = this.db.getAllSync<any>(
      'SELECT * FROM tasks WHERE deleted = 1 ORDER BY deleted_at DESC'
    );
    return rows.map(row => this.mapRowToTask(row));
  }

  /**
   * Create a new task
   */
  async create(task: Task): Promise<Task> {
    const now = Math.floor(Date.now() / 1000);
    const result = this.db.runSync(
      `INSERT INTO tasks
       (client_id, server_id, todo_list_id, todo, priority, done, date_added,
        date_completed, due_date, deleted, deleted_at, created_at, updated_at, version)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        task.client_id,
        task.id || 0,
        task.todo_list_id,
        task.todo,
        task.priority,
        task.done ? 1 : 0,
        task.date_added,
        task.date_completed || null,
        task.due_date || null,
        task.deleted ? 1 : 0,
        task.deleted_at || null,
        task.created_at || now,
        task.updated_at || now,
        task.version || 1,
      ]
    );

    return {
      ...task,
      id: result.lastInsertRowId,
      created_at: task.created_at || now,
      updated_at: task.updated_at || now,
    };
  }

  /**
   * Update an existing task
   */
  async update(task: Task): Promise<Task> {
    const now = Math.floor(Date.now() / 1000);
    this.db.runSync(
      `UPDATE tasks SET
       todo = ?, priority = ?, done = ?, date_completed = ?, due_date = ?,
       deleted = ?, deleted_at = ?, updated_at = ?, version = version + 1
       WHERE id = ?`,
      [
        task.todo,
        task.priority,
        task.done ? 1 : 0,
        task.date_completed ?? null,
        task.due_date ?? null,
        task.deleted ? 1 : 0,
        task.deleted_at ?? null,
        now,
        task.id ?? null,
      ]
    );

    return {
      ...task,
      updated_at: now,
      version: (task.version || 1) + 1,
    };
  }

  /**
   * Soft delete a task (mark as deleted)
   */
  async delete(id: number): Promise<void> {
    const now = Math.floor(Date.now() / 1000);
    this.db.runSync(
      'UPDATE tasks SET deleted = 1, deleted_at = ?, updated_at = ? WHERE id = ?',
      [now, now, id]
    );
  }

  /**
   * Hard delete a task (permanent deletion)
   * Only use this for cleanup after server confirmation
   */
  async hardDelete(id: number): Promise<void> {
    this.db.runSync('DELETE FROM tasks WHERE id = ?', [id]);
  }

  /**
   * Count total tasks (not deleted)
   */
  async count(): Promise<number> {
    const row = this.db.getFirstSync<{ count: number }>(
      'SELECT COUNT(*) as count FROM tasks WHERE deleted = 0'
    );
    return row?.count ?? 0;
  }

  /**
   * Bulk update tasks (for sync operations)
   */
  async bulkUpdate(tasks: Task[]): Promise<void> {
    this.db.withTransactionSync(() => {
      for (const task of tasks) {
        this.db.runSync(
          `UPDATE tasks SET
           todo = ?, priority = ?, done = ?, date_completed = ?, due_date = ?,
           deleted = ?, deleted_at = ?, updated_at = ?, version = ?,
           server_id = COALESCE(?, server_id)
           WHERE client_id = ?`,
          [
            task.todo,
            task.priority,
            task.done ? 1 : 0,
            task.date_completed || null,
            task.due_date || null,
            task.deleted ? 1 : 0,
            task.deleted_at || null,
            task.updated_at,
            task.version || 1,
            task.id || null,
            task.client_id,
          ]
        );
      }
    });
  }

  /**
   * Bulk insert tasks (for sync operations)
   */
  async bulkInsert(tasks: Task[]): Promise<void> {
    this.db.withTransactionSync(() => {
      for (const task of tasks) {
        this.db.runSync(
          `INSERT OR IGNORE INTO tasks
           (client_id, server_id, todo_list_id, todo, priority, done, date_added,
            date_completed, due_date, deleted, deleted_at, created_at, updated_at, version)
           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
          [
            task.client_id,
            task.id || 0,
            task.todo_list_id,
            task.todo,
            task.priority,
            task.done ? 1 : 0,
            task.date_added,
            task.date_completed || null,
            task.due_date || null,
            task.deleted ? 1 : 0,
            task.deleted_at || null,
            task.created_at || Math.floor(Date.now() / 1000),
            task.updated_at || Math.floor(Date.now() / 1000),
            task.version || 1,
          ]
        );
      }
    });
  }

  /**
   * Clear all tasks from a list
   */
  async clearList(listId: number): Promise<void> {
    const now = Math.floor(Date.now() / 1000);
    this.db.runSync(
      'UPDATE tasks SET deleted = 1, deleted_at = ?, updated_at = ? WHERE todo_list_id = ?',
      [now, now, listId]
    );
  }

  /**
   * Map database row to Task object
   */
  private mapRowToTask(row: any): Task {
    return {
      id: row.server_id || row.id,
      client_id: row.client_id,
      todo_list_id: row.todo_list_id,
      todo: row.todo,
      priority: row.priority,
      done: Boolean(row.done),
      date_added: row.date_added,
      date_completed: row.date_completed || null,
      due_date: row.due_date || null,
      deleted: Boolean(row.deleted),
      deleted_at: row.deleted_at || null,
      created_at: row.created_at,
      updated_at: row.updated_at,
      version: row.version,
    };
  }
}
