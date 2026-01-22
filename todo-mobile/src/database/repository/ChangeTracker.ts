/**
 * Change Tracker
 *
 * Tracks all local changes (create, update, delete) for sync purposes.
 * Implements change log pattern for offline-first synchronization.
 * Uses expo-sqlite for Expo compatibility.
 */

import type { SQLiteDatabase } from 'expo-sqlite';

export type ChangeType = 'create' | 'update' | 'delete';
export type EntityType = 'task' | 'list';

export interface Change {
  id: number;
  entityType: EntityType;
  entityId: number;
  changeType: ChangeType;
  timestamp: number;
  synced: boolean;
}

export class ChangeTracker {
  constructor(private db: SQLiteDatabase) {}

  /**
   * Record a change (create, update, or delete)
   */
  async recordChange(entityType: EntityType, entityId: number, changeType: ChangeType): Promise<Change> {
    const timestamp = Math.floor(Date.now() / 1000);

    const result = this.db.runSync(
      `INSERT INTO change_log (entity_type, entity_id, change_type, timestamp, synced)
       VALUES (?, ?, ?, ?, ?)`,
      [entityType, entityId, changeType, timestamp, 0]
    );

    return {
      id: result.lastInsertRowId,
      entityType,
      entityId,
      changeType,
      timestamp,
      synced: false,
    };
  }

  /**
   * Record task creation
   */
  async recordTaskCreate(taskId: number): Promise<Change> {
    return this.recordChange('task', taskId, 'create');
  }

  /**
   * Record task update
   */
  async recordTaskUpdate(taskId: number): Promise<Change> {
    return this.recordChange('task', taskId, 'update');
  }

  /**
   * Record task deletion
   */
  async recordTaskDelete(taskId: number): Promise<Change> {
    return this.recordChange('task', taskId, 'delete');
  }

  /**
   * Record list creation
   */
  async recordListCreate(listId: number): Promise<Change> {
    return this.recordChange('list', listId, 'create');
  }

  /**
   * Record list update
   */
  async recordListUpdate(listId: number): Promise<Change> {
    return this.recordChange('list', listId, 'update');
  }

  /**
   * Record list deletion
   */
  async recordListDelete(listId: number): Promise<Change> {
    return this.recordChange('list', listId, 'delete');
  }

  /**
   * Get all unsynced changes
   */
  async getUnsyncedChanges(): Promise<Change[]> {
    const rows = this.db.getAllSync<any>(
      `SELECT * FROM change_log WHERE synced = 0 ORDER BY timestamp ASC`
    );
    return rows.map(row => this.mapRowToChange(row));
  }

  /**
   * Get unsynced changes for a specific entity type
   */
  async getUnsyncedChangesByType(entityType: EntityType): Promise<Change[]> {
    const rows = this.db.getAllSync<any>(
      `SELECT * FROM change_log WHERE synced = 0 AND entity_type = ? ORDER BY timestamp ASC`,
      [entityType]
    );
    return rows.map(row => this.mapRowToChange(row));
  }

  /**
   * Get all changes (synced and unsynced)
   */
  async getAllChanges(): Promise<Change[]> {
    const rows = this.db.getAllSync<any>(
      `SELECT * FROM change_log ORDER BY timestamp ASC`
    );
    return rows.map(row => this.mapRowToChange(row));
  }

  /**
   * Mark a change as synced
   */
  async markSynced(changeId: number): Promise<void> {
    this.db.runSync(
      'UPDATE change_log SET synced = 1 WHERE id = ?',
      [changeId]
    );
  }

  /**
   * Mark multiple changes as synced
   */
  async markMultipleSynced(changeIds: number[]): Promise<void> {
    if (changeIds.length === 0) return;

    const placeholders = changeIds.map(() => '?').join(',');
    this.db.runSync(
      `UPDATE change_log SET synced = 1 WHERE id IN (${placeholders})`,
      changeIds
    );
  }

  /**
   * Mark all changes as synced
   */
  async markAllSynced(): Promise<void> {
    this.db.runSync('UPDATE change_log SET synced = 1');
  }

  /**
   * Clear synced changes (cleanup after successful sync)
   */
  async clearSyncedChanges(): Promise<void> {
    this.db.runSync('DELETE FROM change_log WHERE synced = 1');
  }

  /**
   * Clear all changes
   */
  async clearAllChanges(): Promise<void> {
    this.db.runSync('DELETE FROM change_log');
  }

  /**
   * Get changes for specific entity
   */
  async getChangesForEntity(entityType: EntityType, entityId: number): Promise<Change[]> {
    const rows = this.db.getAllSync<any>(
      `SELECT * FROM change_log WHERE entity_type = ? AND entity_id = ? ORDER BY timestamp ASC`,
      [entityType, entityId]
    );
    return rows.map(row => this.mapRowToChange(row));
  }

  /**
   * Count unsynced changes
   */
  async countUnsyncedChanges(): Promise<number> {
    const row = this.db.getFirstSync<{ count: number }>(
      'SELECT COUNT(*) as count FROM change_log WHERE synced = 0'
    );
    return row?.count ?? 0;
  }

  /**
   * Check if entity has pending changes
   */
  async hasPendingChanges(entityType: EntityType, entityId: number): Promise<boolean> {
    const row = this.db.getFirstSync<{ count: number }>(
      `SELECT COUNT(*) as count FROM change_log WHERE entity_type = ? AND entity_id = ? AND synced = 0`,
      [entityType, entityId]
    );
    return (row?.count ?? 0) > 0;
  }

  /**
   * Consolidate changes for an entity (keep only latest change type)
   * This optimizes the change log by removing redundant intermediate states
   * e.g., create + update + update becomes just create
   */
  async consolidateChanges(entityType: EntityType, entityId: number): Promise<void> {
    // Get the latest change for this entity
    const latestRow = this.db.getFirstSync<{ id: number; change_type: string }>(
      `SELECT id, change_type FROM change_log
       WHERE entity_type = ? AND entity_id = ? AND synced = 0
       ORDER BY timestamp DESC LIMIT 1`,
      [entityType, entityId]
    );

    if (!latestRow) return;

    // Delete all changes for this entity except the latest one
    this.db.runSync(
      `DELETE FROM change_log
       WHERE entity_type = ? AND entity_id = ? AND synced = 0 AND id != ?`,
      [entityType, entityId, latestRow.id]
    );
  }

  /**
   * Map database row to Change object
   */
  private mapRowToChange(row: any): Change {
    return {
      id: row.id,
      entityType: row.entity_type as EntityType,
      entityId: row.entity_id,
      changeType: row.change_type as ChangeType,
      timestamp: row.timestamp,
      synced: Boolean(row.synced),
    };
  }
}
