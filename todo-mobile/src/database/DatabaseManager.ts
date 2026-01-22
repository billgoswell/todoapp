/**
 * Database Manager
 *
 * Centralizes database initialization and repository management.
 * Provides singleton access to all repositories and database utilities.
 * Uses expo-sqlite for Expo compatibility.
 */

import * as SQLite from 'expo-sqlite';
import { TaskRepository } from './repository/TaskRepository';
import { ListRepository } from './repository/ListRepository';
import { SyncRepository } from './repository/SyncRepository';
import { ChangeTracker } from './repository/ChangeTracker';
import { DATABASE_SCHEMA, DATABASE_NAME } from './schema';

export class DatabaseManager {
  private static instance: DatabaseManager;
  private db: SQLite.SQLiteDatabase | null = null;
  private initialized = false;

  public tasks!: TaskRepository;
  public lists!: ListRepository;
  public sync!: SyncRepository;
  public changes!: ChangeTracker;

  private constructor() {}

  /**
   * Get singleton instance
   */
  static getInstance(): DatabaseManager {
    if (!DatabaseManager.instance) {
      DatabaseManager.instance = new DatabaseManager();
    }
    return DatabaseManager.instance;
  }

  /**
   * Initialize database connection and create repositories
   */
  async initialize(): Promise<void> {
    if (this.initialized) {
      return;
    }

    try {
      // Open database connection
      this.db = SQLite.openDatabaseSync(DATABASE_NAME);

      console.log('Database connection established');

      // Create tables if they don't exist
      await this.createSchema();

      // Initialize repositories
      this.tasks = new TaskRepository(this.db);
      this.lists = new ListRepository(this.db);
      this.sync = new SyncRepository(this.db);
      this.changes = new ChangeTracker(this.db);

      this.initialized = true;
      console.log('Database initialized successfully');
    } catch (error) {
      console.error('Failed to initialize database:', error);
      throw error;
    }
  }

  /**
   * Create database schema
   */
  private async createSchema(): Promise<void> {
    if (!this.db) {
      throw new Error('Database not initialized');
    }

    const statements = DATABASE_SCHEMA.split(';').filter(stmt => stmt.trim());

    for (const statement of statements) {
      if (statement.trim()) {
        try {
          this.db.runSync(statement.trim());
        } catch (error) {
          console.error('Schema creation error:', error);
          throw error;
        }
      }
    }

    console.log('Schema created successfully');
  }

  /**
   * Get database connection (for advanced operations)
   */
  getDatabase(): SQLite.SQLiteDatabase {
    if (!this.db) {
      throw new Error('Database not initialized. Call initialize() first.');
    }
    return this.db;
  }

  /**
   * Execute raw SQL query (for advanced operations)
   */
  async executeSql(sql: string, params: any[] = []): Promise<any> {
    if (!this.db) {
      throw new Error('Database not initialized');
    }

    return this.db.runSync(sql, params);
  }

  /**
   * Perform a transaction
   */
  async transaction<T>(callback: () => T): Promise<T> {
    if (!this.db) {
      throw new Error('Database not initialized');
    }

    return this.db.withTransactionSync(callback);
  }

  /**
   * Clear all data (use with caution)
   */
  async clearAllData(): Promise<void> {
    if (!this.db) {
      throw new Error('Database not initialized');
    }

    this.db.withTransactionSync(() => {
      this.db!.runSync('DELETE FROM change_log');
      this.db!.runSync('DELETE FROM tasks');
      this.db!.runSync('DELETE FROM todo_lists');
      this.db!.runSync('DELETE FROM sync_metadata');
    });
  }

  /**
   * Get database statistics
   */
  async getStats(): Promise<{
    taskCount: number;
    listCount: number;
    changeLogCount: number;
    unsyncedChanges: number;
  }> {
    const taskCount = await this.tasks.count();
    const listCount = await this.lists.count();
    const changeLogCount = await this.changes.countUnsyncedChanges();

    return {
      taskCount,
      listCount,
      changeLogCount,
      unsyncedChanges: changeLogCount,
    };
  }

  /**
   * Close database connection
   */
  async close(): Promise<void> {
    if (this.db) {
      try {
        this.db.closeSync();
        this.db = null;
        this.initialized = false;
        console.log('Database closed');
      } catch (error) {
        console.error('Error closing database:', error);
        throw error;
      }
    }
  }

  /**
   * Check if database is initialized
   */
  isInitialized(): boolean {
    return this.initialized;
  }
}

// Export singleton instance
export const db = DatabaseManager.getInstance();
