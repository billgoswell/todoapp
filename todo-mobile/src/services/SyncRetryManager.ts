/**
 * Sync Retry Manager
 *
 * Manages retry logic for failed sync operations.
 * Implements exponential backoff and jitter for robustness.
 */

import { performFullSync } from '../api/sync';

export interface RetryConfig {
  maxAttempts: number;
  initialDelayMs: number;
  maxDelayMs: number;
  backoffMultiplier: number;
  useJitter: boolean; // Add random jitter to avoid thundering herd
}

export interface RetryAttempt {
  attempt: number;
  timestamp: number;
  error?: Error;
  success: boolean;
}

class SyncRetryManagerService {
  private config: RetryConfig;
  private attempts: Map<string, RetryAttempt[]> = new Map();
  private activeRetries: Set<string> = new Set();
  private retryTimers: Map<string, ReturnType<typeof setTimeout>> = new Map();

  constructor(config: RetryConfig = {
    maxAttempts: 5,
    initialDelayMs: 1000,
    maxDelayMs: 30000,
    backoffMultiplier: 2,
    useJitter: true,
  }) {
    this.config = config;
  }

  /**
   * Attempt sync with automatic retries
   */
  async syncWithRetry(operationId: string = 'default'): Promise<boolean> {
    if (this.activeRetries.has(operationId)) {
      console.log(`Retry already in progress for ${operationId}`);
      return false;
    }

    this.activeRetries.add(operationId);
    this.attempts.set(operationId, []);

    try {
      return await this.executeWithRetry(operationId, 1);
    } finally {
      this.activeRetries.delete(operationId);
    }
  }

  /**
   * Execute sync with retry logic
   */
  private async executeWithRetry(
    operationId: string,
    attemptNumber: number
  ): Promise<boolean> {
    try {
      console.log(`[${operationId}] Sync attempt ${attemptNumber}/${this.config.maxAttempts}`);

      const result = await performFullSync();

      const attempt: RetryAttempt = {
        attempt: attemptNumber,
        timestamp: Date.now(),
        success: result.success,
      };

      this.recordAttempt(operationId, attempt);

      if (result.success) {
        console.log(`[${operationId}] Sync succeeded on attempt ${attemptNumber}`);
        return true;
      } else {
        throw new Error(result.error || 'Sync failed');
      }
    } catch (error) {
      const attempt: RetryAttempt = {
        attempt: attemptNumber,
        timestamp: Date.now(),
        error: error instanceof Error ? error : new Error(String(error)),
        success: false,
      };

      this.recordAttempt(operationId, attempt);

      if (attemptNumber < this.config.maxAttempts) {
        const delayMs = this.calculateDelay(attemptNumber);
        console.log(
          `[${operationId}] Sync failed, retrying in ${delayMs}ms (attempt ${attemptNumber}/${this.config.maxAttempts})`
        );

        return new Promise((resolve) => {
          const timerKey = `${operationId}-${attemptNumber}`;
          const timer = setTimeout(() => {
            this.retryTimers.delete(timerKey);
            this.executeWithRetry(operationId, attemptNumber + 1)
              .then(resolve)
              .catch(() => resolve(false));
          }, delayMs);

          this.retryTimers.set(timerKey, timer);
        });
      } else {
        console.error(
          `[${operationId}] Sync failed after ${attemptNumber} attempts`
        );
        return false;
      }
    }
  }

  /**
   * Calculate exponential backoff delay with optional jitter
   */
  private calculateDelay(attemptNumber: number): number {
    // Exponential backoff: base * multiplier^(attempt-1)
    let delayMs = this.config.initialDelayMs *
      Math.pow(this.config.backoffMultiplier, attemptNumber - 1);

    // Cap at max delay
    delayMs = Math.min(delayMs, this.config.maxDelayMs);

    // Add random jitter (±25%)
    if (this.config.useJitter) {
      const jitter = delayMs * 0.25 * (Math.random() * 2 - 1);
      delayMs += jitter;
    }

    return Math.max(this.config.initialDelayMs, Math.round(delayMs));
  }

  /**
   * Record sync attempt
   */
  private recordAttempt(operationId: string, attempt: RetryAttempt): void {
    const attempts = this.attempts.get(operationId) || [];
    attempts.push(attempt);
    this.attempts.set(operationId, attempts);
  }

  /**
   * Get retry history for operation
   */
  getHistory(operationId: string): RetryAttempt[] {
    return [...(this.attempts.get(operationId) || [])];
  }

  /**
   * Get last attempt for operation
   */
  getLastAttempt(operationId: string): RetryAttempt | null {
    const attempts = this.attempts.get(operationId);
    return attempts && attempts.length > 0 ? attempts[attempts.length - 1] : null;
  }

  /**
   * Cancel ongoing retries for operation
   */
  cancel(operationId: string): void {
    this.activeRetries.delete(operationId);

    // Clear any pending timers for this operation
    Array.from(this.retryTimers.entries()).forEach(([key, timer]) => {
      if (key.startsWith(operationId)) {
        clearTimeout(timer);
        this.retryTimers.delete(key);
      }
    });

    console.log(`[${operationId}] Retry cancelled`);
  }

  /**
   * Cancel all retries
   */
  cancelAll(): void {
    this.activeRetries.forEach((id) => this.cancel(id));
    console.log('All retries cancelled');
  }

  /**
   * Update retry config
   */
  setConfig(config: Partial<RetryConfig>): void {
    this.config = { ...this.config, ...config };
    console.log('Retry config updated:', this.config);
  }

  /**
   * Get current config
   */
  getConfig(): RetryConfig {
    return { ...this.config };
  }

  /**
   * Check if retry is in progress
   */
  isRetrying(operationId: string): boolean {
    return this.activeRetries.has(operationId);
  }

  /**
   * Clear history for operation
   */
  clearHistory(operationId: string): void {
    this.attempts.delete(operationId);
  }

  /**
   * Clear all history
   */
  clearAllHistory(): void {
    this.attempts.clear();
  }

  /**
   * Reset failure state after connectivity is restored.
   * Cancels any pending backoff retries and clears attempt history
   * so the next sync starts with a fresh backoff schedule.
   */
  resetFailureCount(): void {
    this.cancelAll();
    this.clearAllHistory();
  }
}

// Export singleton instance
export const syncRetryManager = new SyncRetryManagerService();
