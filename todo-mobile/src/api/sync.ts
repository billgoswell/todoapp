/**
 * Sync API Methods
 *
 * Higher-level sync operations that coordinate between local database
 * and remote API server. Built on top of ApiClient.
 */

import { apiClient } from './client';
import { db } from '../database';
import { Task, TodoList, PullResponse, PushRequest } from './types';

/**
 * Pull changes from server
 * Fetches all tasks and lists modified since the given timestamp
 */
export async function pullFromServer(since: number = 0): Promise<PullResponse> {
  try {
    const response = await apiClient.pull(since);
    return response;
  } catch (error) {
    console.error('Pull from server failed:', error);
    throw error;
  }
}

/**
 * Push local changes to server
 * Sends all pending tasks and lists to server
 */
export async function pushToServer(): Promise<void> {
  try {
    // Gather all tasks and lists
    const tasks = await db.tasks.findAll();
    const lists = await db.lists.findAll();

    const request: PushRequest = {
      tasks,
      lists,
    };

    // Send to server
    await apiClient.push(request);
  } catch (error) {
    console.error('Push to server failed:', error);
    throw error;
  }
}

/**
 * Perform full sync: pull from server, apply changes, then push local changes
 * This is the main sync operation that should be called periodically
 */
export async function performFullSync(): Promise<{
  success: boolean;
  tasksReceived: number;
  listsReceived: number;
  error?: string;
}> {
  try {
    // 1. Get last sync time
    const lastSyncTime = await db.sync.getLastSyncTime();
    console.log(`Syncing changes since ${lastSyncTime}`);

    // 2. Pull changes from server
    console.log('Pulling changes from server...');
    const pullResponse = await pullFromServer(lastSyncTime);

    // 3. Apply server changes locally (handle null responses from server)
    const serverTasks = pullResponse.tasks ?? [];
    const serverLists = pullResponse.lists ?? [];
    console.log(`Received ${serverTasks.length} tasks and ${serverLists.length} lists`);

    if (serverTasks.length > 0) {
      await applyServerTasks(serverTasks);
    }

    if (serverLists.length > 0) {
      await applyServerLists(serverLists);
    }

    // 4. Push local changes to server
    console.log('Pushing local changes to server...');
    await pushToServer();

    // 5. Update sync metadata
    const now = Math.floor(Date.now() / 1000);
    await db.sync.setLastSyncTime(now);
    await db.sync.setLastSyncStatus('success');
    await db.sync.setLastSyncError(null);
    await db.changes.markAllSynced();

    console.log('Sync completed successfully');

    return {
      success: true,
      tasksReceived: serverTasks.length,
      listsReceived: serverLists.length,
    };
  } catch (error: any) {
    console.error('Full sync failed:', error);

    const errorMessage = error?.message || 'Sync failed';
    await db.sync.setLastSyncStatus('error');
    await db.sync.setLastSyncError(errorMessage);

    return {
      success: false,
      tasksReceived: 0,
      listsReceived: 0,
      error: errorMessage,
    };
  }
}

/**
 * Apply server tasks locally using conflict resolution
 * Strategy: last-write-wins based on updated_at timestamp
 */
async function applyServerTasks(serverTasks: Task[]): Promise<void> {
  for (const serverTask of serverTasks) {
    // Try to find local task by client_id
    const localTask = await db.tasks.findByClientId(serverTask.client_id);

    if (!localTask) {
      // New task from server - insert it
      console.log(`Creating new task from server: ${serverTask.client_id}`);
      await db.tasks.create(serverTask);
    } else {
      // Task exists locally - apply conflict resolution
      if (serverTask.updated_at > localTask.updated_at) {
        // Server is newer - update local with server version
        console.log(`Updating task from server: ${serverTask.client_id}`);
        await db.tasks.update(serverTask);
      } else {
        // Local is newer or same - keep local version
        console.log(`Keeping local task (local is newer): ${serverTask.client_id}`);
      }
    }
  }
}

/**
 * Apply server lists locally using conflict resolution
 * Strategy: last-write-wins based on updated_at timestamp
 */
async function applyServerLists(serverLists: TodoList[]): Promise<void> {
  for (const serverList of serverLists) {
    // Try to find local list by client_id
    const localList = await db.lists.findByClientId(serverList.client_id);

    if (!localList) {
      // New list from server - insert it
      console.log(`Creating new list from server: ${serverList.client_id}`);
      await db.lists.create(serverList);
    } else {
      // List exists locally - apply conflict resolution
      if (serverList.updated_at > localList.updated_at) {
        // Server is newer - update local with server version
        console.log(`Updating list from server: ${serverList.client_id}`);
        await db.lists.update(serverList);
      } else {
        // Local is newer or same - keep local version
        console.log(`Keeping local list (local is newer): ${serverList.client_id}`);
      }
    }
  }
}

/**
 * Test connection to server and validate API key
 */
export async function testServerConnection(): Promise<boolean> {
  try {
    const isConnected = await apiClient.testConnection();
    return isConnected;
  } catch (error) {
    console.error('Connection test failed:', error);
    return false;
  }
}

/**
 * Get sync status summary
 */
export async function getSyncStatus() {
  const metadata = await db.sync.getSyncSummary();
  const unsyncedChanges = await db.changes.countUnsyncedChanges();

  return {
    lastSyncTime: metadata.lastSyncTime,
    lastSyncStatus: metadata.lastSyncStatus,
    lastSyncError: metadata.lastSyncError,
    unsyncedChanges,
    hasUnsynced: unsyncedChanges > 0,
  };
}

/**
 * Force reset sync state (careful - for debugging only)
 */
export async function resetSyncState(): Promise<void> {
  await db.sync.clear();
  await db.changes.clearAllChanges();
  console.log('Sync state reset');
}

/**
 * Get detailed sync diagnostics
 */
export async function getSyncDiagnostics() {
  const stats = await db.getStats();
  const syncStatus = await getSyncStatus();
  const serverUrl = apiClient.getServerUrl();
  const hasApiKey = !!apiClient.getApiKey();

  return {
    database: stats,
    sync: syncStatus,
    server: {
      url: serverUrl || 'Not configured',
      authenticated: hasApiKey,
    },
  };
}
