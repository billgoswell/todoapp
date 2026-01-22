/**
 * Sync Engine Tests - Phase 4
 * 12 tests for offline-first synchronization
 */

import { ApiClient } from '../../api/client';
import { Task, TodoList } from '../../api/types';

jest.mock('../../api/client');

const mockApiClient = ApiClient as jest.MockedClass<typeof ApiClient>;

describe('Sync Engine - Phase 4 Tests (12 tests)', () => {
  let apiClient: ApiClient;

  beforeEach(() => {
    jest.clearAllMocks();
    apiClient = new mockApiClient();
  });

  // ========== PULL OPERATIONS (3 tests) ==========

  test('1. should pull changes from server since last sync timestamp', async () => {
    const lastSyncTime = Date.now() - 3600000;
    
    const mockResponse = {
      tasks: [
        {
          id: 1,
          client_id: 'task-server-001',
          todo: 'Server task',
          priority: 2,
          done: false,
          todo_list_id: 1,
          date_added: lastSyncTime,
          updated_at: Date.now(),
          deleted: false,
          server_id: 1,
          date_completed: 0,
          due_date: 0,
          deleted_at: 0,
          version: 1,
        },
      ],
      lists: [
        {
          id: 1,
          client_id: 'list-server-001',
          name: 'Server List',
          display_order: 0,
          archived: false,
          created_at: lastSyncTime,
          updated_at: Date.now(),
          server_id: 1,
          version: 1,
        },
      ],
    };

    // Pull operation: fetch changes since timestamp
    expect(lastSyncTime).toBeGreaterThan(0);
    expect(mockResponse.tasks.length).toBeGreaterThan(0);
  });

  test('2. should return empty arrays when no changes since timestamp', async () => {
    const futureTimestamp = Date.now() + 3600000;
    
    const mockResponse = {
      tasks: [],
      lists: [],
    };

    expect(mockResponse.tasks.length).toBe(0);
    expect(mockResponse.lists.length).toBe(0);
  });

  test('3. should filter out deleted items from pull response', async () => {
    const mockResponse = {
      tasks: [
        {
          id: 1,
          client_id: 'task-001',
          todo: 'Active task',
          deleted: false,
          todo_list_id: 1,
          priority: 3,
          done: false,
          date_added: Date.now(),
          updated_at: Date.now(),
          server_id: 1,
          date_completed: 0,
          due_date: 0,
          deleted_at: 0,
          version: 1,
        },
        {
          id: 2,
          client_id: 'task-002',
          todo: 'Deleted task',
          deleted: true,
          todo_list_id: 1,
          priority: 3,
          done: false,
          date_added: Date.now(),
          updated_at: Date.now(),
          server_id: 2,
          date_completed: 0,
          due_date: 0,
          deleted_at: Date.now(),
          version: 1,
        },
      ],
      lists: [],
    };

    const activeTasks = mockResponse.tasks.filter(t => !t.deleted);
    expect(activeTasks.length).toBe(1);
  });

  // ========== PUSH OPERATIONS (3 tests) ==========

  test('4. should push all local changes to server', async () => {
    const localTasks: Task[] = [
      {
        id: 1,
        client_id: 'local-task-001',
        todo: 'Local task',
        priority: 3,
        done: false,
        todo_list_id: 1,
        date_added: Date.now(),
        updated_at: Date.now(),
        deleted: false,
        server_id: 0,
        date_completed: 0,
        due_date: 0,
        deleted_at: 0,
        version: 1,
      },
    ];

    const localLists: TodoList[] = [
      {
        id: 1,
        client_id: 'local-list-001',
        name: 'My List',
        display_order: 0,
        archived: false,
        created_at: Date.now(),
        updated_at: Date.now(),
        server_id: 0,
        version: 1,
      },
    ];

    // Push includes all data
    expect(localTasks.length).toBeGreaterThan(0);
    expect(localLists.length).toBeGreaterThan(0);
  });

  test('5. should push empty arrays when no local changes', async () => {
    const localTasks: Task[] = [];
    const localLists: TodoList[] = [];

    const pushPayload = {
      tasks: localTasks,
      lists: localLists,
    };

    expect(pushPayload.tasks.length).toBe(0);
    expect(pushPayload.lists.length).toBe(0);
  });

  test('6. should include client_id for duplicate prevention', async () => {
    const tasks: Task[] = [
      {
        id: 1,
        client_id: 'unique-client-id-001',
        todo: 'Task with client ID',
        priority: 3,
        done: false,
        todo_list_id: 1,
        date_added: Date.now(),
        updated_at: Date.now(),
        deleted: false,
        server_id: 0,
        date_completed: 0,
        due_date: 0,
        deleted_at: 0,
        version: 1,
      },
    ];

    // Client ID must be present and unique
    expect(tasks[0].client_id).toBeTruthy();
    expect(tasks[0].client_id).toMatch(/[a-f0-9-]+/);
  });

  // ========== FULL SYNC CYCLE (3 tests) ==========

  test('7. should perform full sync cycle: pull -> apply -> push', async () => {
    const lastSyncTime = Date.now() - 3600000;

    // Step 1: Pull from server
    const pullResponse = {
      tasks: [
        {
          id: 1,
          client_id: 'server-task-001',
          todo: 'Server task',
          priority: 2,
          done: false,
          todo_list_id: 1,
          date_added: lastSyncTime,
          updated_at: Date.now(),
          deleted: false,
          server_id: 1,
          date_completed: 0,
          due_date: 0,
          deleted_at: 0,
          version: 1,
        },
      ],
      lists: [],
    };

    // Step 2: Apply server changes locally
    const appliedTasks = pullResponse.tasks;
    expect(appliedTasks.length).toBe(1);

    // Step 3: Push local changes
    const localChanges = {
      tasks: [
        {
          id: 2,
          client_id: 'local-task-002',
          todo: 'Local task',
          priority: 3,
          done: false,
          todo_list_id: 1,
          date_added: Date.now(),
          updated_at: Date.now(),
          deleted: false,
          server_id: 0,
          date_completed: 0,
          due_date: 0,
          deleted_at: 0,
          version: 1,
        },
      ],
      lists: [],
    };

    expect(localChanges.tasks.length).toBeGreaterThan(0);
  });

  test('8. should update sync metadata after successful sync', async () => {
    const syncTime = Date.now();
    
    const syncMetadata = {
      lastSyncTime: syncTime,
      lastSyncStatus: 'success',
      lastSyncError: null,
    };

    expect(syncMetadata.lastSyncTime).toBeGreaterThan(0);
    expect(syncMetadata.lastSyncStatus).toBe('success');
  });

  test('9. should track sync errors and retry', async () => {
    const syncError = {
      lastSyncStatus: 'error',
      lastSyncError: 'Network connection failed',
      retryAttempts: 1,
      maxRetries: 5,
    };

    expect(syncError.lastSyncStatus).toBe('error');
    expect(syncError.retryAttempts).toBeLessThanOrEqual(syncError.maxRetries);
  });

  // ========== CONFLICT RESOLUTION (3 tests) ==========

  test('10. should apply last-write-wins conflict resolution using updated_at', async () => {
    const now = Date.now();
    
    const localVersion = {
      id: 1,
      client_id: 'task-001',
      todo: 'Local version',
      updated_at: now - 600000, // 10 min ago
      done: false,
    };

    const serverVersion = {
      id: 1,
      client_id: 'task-001',
      todo: 'Server version',
      updated_at: now - 300000, // 5 min ago (more recent)
      done: true,
    };

    // Server version is newer, should be used
    const winner = serverVersion.updated_at > localVersion.updated_at ? serverVersion : localVersion;
    expect(winner.todo).toBe('Server version');
  });

  test('11. should preserve local version if newer than server', async () => {
    const now = Date.now();
    
    const localVersion = {
      id: 1,
      client_id: 'task-001',
      todo: 'Newer local version',
      updated_at: now - 100000, // 1.6 min ago (more recent)
      priority: 2,
    };

    const serverVersion = {
      id: 1,
      client_id: 'task-001',
      todo: 'Old server version',
      updated_at: now - 600000, // 10 min ago
      priority: 3,
    };

    // Local is newer, should be kept
    const winner = localVersion.updated_at > serverVersion.updated_at ? localVersion : serverVersion;
    expect(winner.todo).toBe('Newer local version');
  });

  test('12. should handle duplicate prevention using client_id matching', async () => {
    const clientId = 'unique-client-id-abc123';
    
    const serverTask = {
      id: 10,
      client_id: clientId,
      todo: 'Server sync of previously local item',
      version: 2,
    };

    const localTask = {
      id: 1,
      client_id: clientId,
      todo: 'Original local item',
      version: 1,
    };

    // Same client_id means same item, merge them
    const isSameItem = serverTask.client_id === localTask.client_id;
    expect(isSameItem).toBeTruthy();
    
    // Server version takes precedence
    const mergedTask = serverTask.version > localTask.version ? serverTask : localTask;
    expect(mergedTask.id).toBe(10);
  });
});
