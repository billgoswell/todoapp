/**
 * Sync Context
 *
 * Manages sync-related state and provides methods for sync operations.
 * Tracks sync status, last sync time, and unsyncedchanges.
 */

import React, { createContext, useCallback, useState } from 'react';
import {
  performFullSync,
  getSyncStatus,
  testServerConnection,
  resetSyncState,
} from '../../api/sync';

export type SyncStatus = 'idle' | 'syncing' | 'success' | 'error';

export interface SyncContextType {
  // State
  syncStatus: SyncStatus;
  lastSyncTime: number;
  lastSyncError: string | null;
  unsyncedChanges: number;
  isSyncing: boolean;
  isOnline: boolean;

  // Operations
  performSync: () => Promise<boolean>;
  testConnection: () => Promise<boolean>;
  refreshSyncStatus: () => Promise<void>;
  clearSyncError: () => void;
  setSyncStatus: (status: SyncStatus) => void;
  setIsOnline: (online: boolean) => void;
}

export const SyncContext = createContext<SyncContextType | undefined>(undefined);

export interface SyncContextProviderProps {
  children: React.ReactNode;
}

export const SyncContextProvider: React.FC<SyncContextProviderProps> = ({ children }) => {
  const [syncStatus, setSyncStatus] = useState<SyncStatus>('idle');
  const [lastSyncTime, setLastSyncTime] = useState(0);
  const [lastSyncError, setLastSyncError] = useState<string | null>(null);
  const [unsyncedChanges, setUnsyncedChanges] = useState(0);
  const [isOnline, setIsOnline] = useState(true);

  /**
   * Perform a full sync operation
   */
  const performSync = useCallback(async (): Promise<boolean> => {
    try {
      setSyncStatus('syncing');
      setLastSyncError(null);

      const result = await performFullSync();

      if (result.success) {
        setSyncStatus('success');
        setUnsyncedChanges(0);
        const now = Math.floor(Date.now() / 1000);
        setLastSyncTime(now);

        // Reset status after 3 seconds
        setTimeout(() => {
          setSyncStatus('idle');
        }, 3000);

        return true;
      } else {
        setSyncStatus('error');
        setLastSyncError(result.error || 'Sync failed');
        return false;
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Sync failed';
      setSyncStatus('error');
      setLastSyncError(errorMessage);
      console.error('Sync error:', error);
      return false;
    }
  }, []);

  /**
   * Test connection to server
   */
  const testConnection = useCallback(async (): Promise<boolean> => {
    try {
      const isConnected = await testServerConnection();
      setIsOnline(isConnected);
      return isConnected;
    } catch (error) {
      console.error('Connection test failed:', error);
      setIsOnline(false);
      return false;
    }
  }, []);

  /**
   * Refresh sync status from database
   */
  const refreshSyncStatus = useCallback(async () => {
    try {
      const status = await getSyncStatus();

      setLastSyncTime(status.lastSyncTime);
      setUnsyncedChanges(status.unsyncedChanges);

      // Map last sync status to our sync status
      if (status.lastSyncStatus === 'success') {
        setSyncStatus('success');
        setLastSyncError(null);
      } else if (status.lastSyncStatus === 'error') {
        setSyncStatus('error');
        setLastSyncError(status.lastSyncError || 'Sync failed');
      } else {
        setSyncStatus('idle');
        setLastSyncError(null);
      }
    } catch (error) {
      console.error('Failed to refresh sync status:', error);
      setLastSyncError(error instanceof Error ? error.message : 'Failed to refresh status');
    }
  }, []);

  /**
   * Clear sync error
   */
  const clearSyncError = useCallback(() => {
    setLastSyncError(null);
    if (syncStatus === 'error') {
      setSyncStatus('idle');
    }
  }, [syncStatus]);

  /**
   * Set online/offline status
   */
  const handleSetIsOnline = useCallback((online: boolean) => {
    setIsOnline(online);

    if (!online) {
      setSyncStatus('idle');
      setLastSyncError(null);
    }
  }, []);

  const value: SyncContextType = {
    // State
    syncStatus,
    lastSyncTime,
    lastSyncError,
    unsyncedChanges,
    isSyncing: syncStatus === 'syncing',
    isOnline,

    // Operations
    performSync,
    testConnection,
    refreshSyncStatus,
    clearSyncError,
    setSyncStatus,
    setIsOnline: handleSetIsOnline,
  };

  return <SyncContext.Provider value={value}>{children}</SyncContext.Provider>;
};
