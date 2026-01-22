/**
 * App Provider
 *
 * Wraps all context providers and initializes app state.
 * This component combines TaskContext, ListContext, SyncContext,
 * and sets up background services for sync and network monitoring.
 */

import React, { useEffect } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { TaskContextProvider } from './context/TaskContext';
import { ListContextProvider } from './context/ListContext';
import { SyncContextProvider } from './context/SyncContext';
import { useTasks, useLists, useSync } from './hooks';
import { networkMonitor, backgroundSyncScheduler, syncRetryManager } from '../services';

interface AppProviderProps {
  children: React.ReactNode;
}

/**
 * Inner component that uses hooks (must be inside providers)
 */
const AppInitializer: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { loadTasks } = useTasks();
  const { loadLists } = useLists();
  const { refreshSyncStatus, testConnection, setIsOnline } = useSync();

  useEffect(() => {
    const initializeApp = async () => {
      try {
        // Load initial data from database
        console.log('Loading initial app data...');
        await Promise.all([loadTasks(), loadLists()]);

        // Refresh sync status
        await refreshSyncStatus();

        // Test connection
        const isOnline = await testConnection();
        console.log('Server connection:', isOnline ? 'online' : 'offline');
      } catch (error) {
        console.error('Failed to initialize app state:', error);
      }
    };

    initializeApp();
  }, [loadTasks, loadLists, refreshSyncStatus, testConnection]);

  // Set up background services
  useEffect(() => {
    console.log('Setting up background services...');

    // Start network monitoring
    networkMonitor.startMonitoring();

    // Add network change listener
    const networkListener = {
      onNetworkChange: (status: 'online' | 'offline' | 'unknown') => {
        console.log('Network status changed:', status);
        setIsOnline(status === 'online');

        // Reset retry manager on connection recovery
        if (status === 'online') {
          syncRetryManager.resetFailureCount();

          // Trigger immediate sync when coming back online
          console.log('Connection restored, triggering sync...');
          backgroundSyncScheduler.forceSync().catch((error) => {
            console.error('Failed to sync after reconnection:', error);
          });
        }
      },
    };

    networkMonitor.addListener(networkListener);

    // Start background sync scheduler
    backgroundSyncScheduler.start();

    // Handle app state changes
    let appStateSubscription: any = null;

    const handleAppStateChange = (state: AppStateStatus) => {
      if (state === 'active') {
        console.log('App came to foreground - starting sync scheduler');
        if (!backgroundSyncScheduler.isRunning()) {
          backgroundSyncScheduler.start();
        }
        // Force a sync when app comes to foreground
        backgroundSyncScheduler.forceSync().catch((error) => {
          console.error('Failed to sync on foreground:', error);
        });
      } else if (state === 'background') {
        console.log('App went to background');
        // Could keep scheduler running or pause it
        // For now, let it continue running
      }
    };

    appStateSubscription = AppState.addEventListener('change', handleAppStateChange);

    // Cleanup
    return () => {
      console.log('Cleaning up background services...');
      networkMonitor.removeListener(networkListener);
      backgroundSyncScheduler.stop();
      syncRetryManager.cancelAll();
      networkMonitor.stopMonitoring();

      if (appStateSubscription) {
        appStateSubscription.remove();
      }
    };
  }, [setIsOnline]);

  return <>{children}</>;
};

/**
 * Root provider that wraps all contexts
 * Usage: <AppProvider><App /></AppProvider>
 */
export const AppProvider: React.FC<AppProviderProps> = ({ children }) => {
  return (
    <TaskContextProvider>
      <ListContextProvider>
        <SyncContextProvider>
          <AppInitializer>{children}</AppInitializer>
        </SyncContextProvider>
      </ListContextProvider>
    </TaskContextProvider>
  );
};

export default AppProvider;
