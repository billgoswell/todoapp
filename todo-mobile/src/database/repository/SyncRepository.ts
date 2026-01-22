/**
 * Sync Repository
 *
 * Handles sync metadata management (last sync time, sync status, etc.)
 * Uses expo-sqlite for Expo compatibility.
 */

import type { SQLiteDatabase } from 'expo-sqlite';

export interface SyncMetadata {
  lastSyncTime: number; // Unix timestamp
  lastSyncStatus: 'success' | 'error' | 'pending';
  lastSyncError?: string;
  lastPullTime?: number;
  lastPushTime?: number;
}

export class SyncRepository {
  constructor(private db: SQLiteDatabase) {}

  /**
   * Get last sync timestamp
   */
  async getLastSyncTime(): Promise<number> {
    const row = this.db.getFirstSync<{ value: string }>(
      'SELECT value FROM sync_metadata WHERE key = ?',
      ['last_sync_time']
    );
    return row ? parseInt(row.value, 10) : 0;
  }

  /**
   * Set last sync timestamp
   */
  async setLastSyncTime(timestamp: number): Promise<void> {
    this.db.runSync(
      `INSERT OR REPLACE INTO sync_metadata (key, value) VALUES ('last_sync_time', ?)`,
      [timestamp.toString()]
    );
  }

  /**
   * Get last sync status
   */
  async getLastSyncStatus(): Promise<'success' | 'error' | 'pending'> {
    const row = this.db.getFirstSync<{ value: string }>(
      'SELECT value FROM sync_metadata WHERE key = ?',
      ['last_sync_status']
    );
    return (row?.value as 'success' | 'error' | 'pending') ?? 'pending';
  }

  /**
   * Set last sync status
   */
  async setLastSyncStatus(status: 'success' | 'error' | 'pending'): Promise<void> {
    this.db.runSync(
      `INSERT OR REPLACE INTO sync_metadata (key, value) VALUES ('last_sync_status', ?)`,
      [status]
    );
  }

  /**
   * Get last sync error message
   */
  async getLastSyncError(): Promise<string | null> {
    const row = this.db.getFirstSync<{ value: string }>(
      'SELECT value FROM sync_metadata WHERE key = ?',
      ['last_sync_error']
    );
    return row?.value ?? null;
  }

  /**
   * Set last sync error
   */
  async setLastSyncError(error: string | null): Promise<void> {
    if (error === null) {
      this.db.runSync(
        'DELETE FROM sync_metadata WHERE key = ?',
        ['last_sync_error']
      );
    } else {
      this.db.runSync(
        `INSERT OR REPLACE INTO sync_metadata (key, value) VALUES ('last_sync_error', ?)`,
        [error]
      );
    }
  }

  /**
   * Get all sync metadata
   */
  async getMetadata(): Promise<Record<string, string>> {
    const rows = this.db.getAllSync<{ key: string; value: string }>(
      'SELECT key, value FROM sync_metadata'
    );
    const metadata: Record<string, string> = {};
    for (const row of rows) {
      metadata[row.key] = row.value;
    }
    return metadata;
  }

  /**
   * Clear all sync metadata
   */
  async clear(): Promise<void> {
    this.db.runSync('DELETE FROM sync_metadata');
  }

  /**
   * Track pull operation
   */
  async recordPull(): Promise<void> {
    const now = Math.floor(Date.now() / 1000);
    this.db.runSync(
      `INSERT OR REPLACE INTO sync_metadata (key, value) VALUES ('last_pull_time', ?)`,
      [now.toString()]
    );
  }

  /**
   * Track push operation
   */
  async recordPush(): Promise<void> {
    const now = Math.floor(Date.now() / 1000);
    this.db.runSync(
      `INSERT OR REPLACE INTO sync_metadata (key, value) VALUES ('last_push_time', ?)`,
      [now.toString()]
    );
  }

  /**
   * Get full sync metadata summary
   */
  async getSyncSummary(): Promise<SyncMetadata> {
    const metadata = await this.getMetadata();

    return {
      lastSyncTime: parseInt(metadata['last_sync_time'] || '0', 10),
      lastSyncStatus: (metadata['last_sync_status'] || 'pending') as 'success' | 'error' | 'pending',
      lastSyncError: metadata['last_sync_error'] || undefined,
      lastPullTime: metadata['last_pull_time'] ? parseInt(metadata['last_pull_time'], 10) : undefined,
      lastPushTime: metadata['last_push_time'] ? parseInt(metadata['last_push_time'], 10) : undefined,
    };
  }
}
