/**
 * useSync Hook
 *
 * Custom hook for accessing and managing sync state.
 * Provides sync operations and state from SyncContext.
 */

import { useCallback, useContext, useEffect } from 'react';
import { SyncContext, SyncContextType } from '../context/SyncContext';
import { useTasks } from './useTasks';
import { useLists } from './useLists';

/**
 * Use sync state and operations
 * Must be used within SyncContextProvider
 */
export const useSync = (): SyncContextType => {
  const context = useContext(SyncContext);

  if (context === undefined) {
    throw new Error('useSync must be used within a SyncContextProvider');
  }

  return context;
};

/**
 * Hook to sync and reload data
 * Performs sync then reloads tasks and lists
 */
export const useSyncAndReload = () => {
  const { performSync, refreshSyncStatus } = useSync();
  const { loadTasks } = useTasks();
  const { loadLists } = useLists();

  const syncAndReload = useCallback(async () => {
    try {
      const success = await performSync();

      if (success) {
        // Reload all data after successful sync
        await Promise.all([loadTasks(), loadLists()]);
        await refreshSyncStatus();
      }

      return success;
    } catch (error) {
      console.error('Sync and reload failed:', error);
      return false;
    }
  }, [performSync, loadTasks, loadLists, refreshSyncStatus]);

  return { syncAndReload };
};

/**
 * Hook for periodic auto-sync
 * Triggers sync at specified interval when app is in foreground
 */
export const useAutoSync = (intervalSeconds: number = 60) => {
  const { performSync, isOnline } = useSync();

  useEffect(() => {
    if (!isOnline) {
      return;
    }

    // Initial sync
    performSync().catch(error => console.error('Auto-sync failed:', error));

    // Set up interval
    const interval = setInterval(() => {
      performSync().catch(error => console.error('Auto-sync failed:', error));
    }, intervalSeconds * 1000);

    return () => clearInterval(interval);
  }, [isOnline, intervalSeconds, performSync]);
};

/**
 * Hook to display sync status message
 */
export const useSyncStatusMessage = (): string => {
  const { syncStatus, lastSyncError, isSyncing } = useSync();

  if (isSyncing) {
    return 'Syncing...';
  }

  if (syncStatus === 'success') {
    return 'Synced';
  }

  if (syncStatus === 'error') {
    return `Sync failed: ${lastSyncError || 'Unknown error'}`;
  }

  return 'Not synced';
};

/**
 * Hook to sync on demand with error handling
 */
export const useSyncOnDemand = () => {
  const { performSync, clearSyncError } = useSync();
  const { loadTasks } = useTasks();
  const { loadLists } = useLists();

  const sync = useCallback(async () => {
    try {
      clearSyncError();
      const success = await performSync();

      if (success) {
        // Reload data after sync
        await Promise.all([loadTasks(), loadLists()]);
      }

      return success;
    } catch (error) {
      console.error('On-demand sync failed:', error);
      return false;
    }
  }, [performSync, loadTasks, loadLists, clearSyncError]);

  return { sync };
};
