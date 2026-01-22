/**
 * Background Sync Scheduler
 *
 * Manages periodic background synchronization with the server.
 * Handles scheduling, execution, and error recovery.
 */

import { performFullSync, getSyncStatus } from '../api/sync';

export interface SyncScheduleConfig {
  intervalSeconds: number; // How often to sync
  maxRetries: number; // Max retry attempts
  retryDelayMs: number; // Initial retry delay
  retryBackoffMultiplier: number; // Exponential backoff multiplier
}

export interface SyncScheduleStatus {
  isRunning: boolean;
  lastSyncTime: number | null;
  lastSyncSuccess: boolean;
  nextScheduledSync: number | null;
  isSyncing: boolean;
  failureCount: number;
}

class BackgroundSyncSchedulerService {
  private intervalId: NodeJS.Timer | null = null;
  private config: SyncScheduleConfig;
  private status: SyncScheduleStatus;
  private lastSyncTime: number = 0;
  private failureCount: number = 0;
  private isSyncing: boolean = false;
  private isEnabled: boolean = true;

  constructor(config: SyncScheduleConfig = {
    intervalSeconds: 60,
    maxRetries: 3,
    retryDelayMs: 1000,
    retryBackoffMultiplier: 2,
  }) {
    this.config = config;
    this.status = {
      isRunning: false,
      lastSyncTime: null,
      lastSyncSuccess: false,
      nextScheduledSync: null,
      isSyncing: false,
      failureCount: 0,
    };
  }

  /**
   * Start background sync scheduling
   */
  start(): void {
    if (this.intervalId) {
      console.log('Background sync scheduler already started');
      return;
    }

    if (!this.isEnabled) {
      console.log('Background sync scheduler is disabled');
      return;
    }

    console.log(`Starting background sync scheduler (interval: ${this.config.intervalSeconds}s)`);

    // Perform immediate sync
    this.performSync().catch((error) => {
      console.error('Initial sync failed:', error);
    });

    // Schedule periodic syncs
    this.intervalId = setInterval(() => {
      this.performSync().catch((error) => {
        console.error('Scheduled sync failed:', error);
      });
    }, this.config.intervalSeconds * 1000);

    this.status.isRunning = true;
    this.scheduleNextSync();
  }

  /**
   * Stop background sync scheduling
   */
  stop(): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
      console.log('Background sync scheduler stopped');
    }

    this.status.isRunning = false;
    this.status.nextScheduledSync = null;
  }

  /**
   * Perform a sync operation with retries
   */
  private async performSync(): Promise<void> {
    if (this.isSyncing) {
      console.log('Sync already in progress, skipping');
      return;
    }

    this.isSyncing = true;
    this.status.isSyncing = true;

    try {
      console.log('Starting background sync...');

      const result = await performFullSync();

      if (result.success) {
        console.log(`Background sync successful: ${result.tasksReceived} tasks, ${result.listsReceived} lists`);
        this.lastSyncTime = Date.now();
        this.failureCount = 0;
        this.status.lastSyncSuccess = true;
      } else {
        console.warn(`Background sync failed: ${result.error}`);
        await this.handleSyncFailure();
      }

      this.status.lastSyncTime = this.lastSyncTime;
      this.status.failureCount = this.failureCount;
    } catch (error) {
      console.error('Background sync error:', error);
      await this.handleSyncFailure();
    } finally {
      this.isSyncing = false;
      this.status.isSyncing = false;
      this.scheduleNextSync();
    }
  }

  /**
   * Handle sync failure with exponential backoff
   */
  private async handleSyncFailure(): Promise<void> {
    this.failureCount++;
    this.status.failureCount = this.failureCount;
    this.status.lastSyncSuccess = false;

    if (this.failureCount <= this.config.maxRetries) {
      // Calculate exponential backoff delay
      const delayMs = this.config.retryDelayMs *
        Math.pow(this.config.retryBackoffMultiplier, this.failureCount - 1);

      console.log(`Sync failed (${this.failureCount}/${this.config.maxRetries}), retrying in ${delayMs}ms`);

      // Schedule retry
      setTimeout(() => {
        if (this.isEnabled && this.intervalId) {
          this.performSync().catch((error) => {
            console.error('Retry sync failed:', error);
          });
        }
      }, delayMs);
    } else {
      console.error(`Maximum sync retries exceeded (${this.config.maxRetries})`);
      // Could emit event here to notify UI of persistent failure
    }
  }

  /**
   * Schedule next sync time
   */
  private scheduleNextSync(): void {
    if (this.status.isRunning) {
      this.status.nextScheduledSync = Date.now() + (this.config.intervalSeconds * 1000);
    }
  }

  /**
   * Get current scheduler status
   */
  getStatus(): SyncScheduleStatus {
    return { ...this.status };
  }

  /**
   * Update sync interval
   */
  setInterval(intervalSeconds: number): void {
    if (intervalSeconds <= 0) {
      console.error('Sync interval must be greater than 0');
      return;
    }

    this.config.intervalSeconds = intervalSeconds;

    // Restart scheduler with new interval
    if (this.intervalId) {
      this.stop();
      this.start();
    }

    console.log(`Sync interval updated to ${intervalSeconds}s`);
  }

  /**
   * Force immediate sync
   */
  async forceSync(): Promise<void> {
    console.log('Forcing immediate sync...');
    await this.performSync();
  }

  /**
   * Enable/disable background sync
   */
  setEnabled(enabled: boolean): void {
    this.isEnabled = enabled;

    if (enabled) {
      console.log('Background sync enabled');
      if (!this.intervalId) {
        this.start();
      }
    } else {
      console.log('Background sync disabled');
      this.stop();
    }
  }

  /**
   * Check if scheduler is running
   */
  isRunning(): boolean {
    return this.status.isRunning;
  }

  /**
   * Get last sync time
   */
  getLastSyncTime(): number | null {
    return this.status.lastSyncTime;
  }

  /**
   * Get time until next scheduled sync
   */
  getTimeUntilNextSync(): number {
    if (!this.status.nextScheduledSync) {
      return -1;
    }

    const timeLeft = this.status.nextScheduledSync - Date.now();
    return Math.max(0, timeLeft);
  }

  /**
   * Reset failure count (e.g., when connection restored)
   */
  resetFailureCount(): void {
    console.log('Resetting sync failure count');
    this.failureCount = 0;
    this.status.failureCount = 0;
  }
}

// Export singleton instance
export const backgroundSyncScheduler = new BackgroundSyncSchedulerService();
