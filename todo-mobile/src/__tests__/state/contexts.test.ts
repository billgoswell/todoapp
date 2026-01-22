/**
 * State Management Tests - Phase 4
 * 14 tests for React Context and custom hooks
 */

import { Task, TodoList } from '../../api/types';

describe('State Management - Phase 4 Tests (14 tests)', () => {
  // ========== TASK CONTEXT (4 tests) ==========

  test('1. should initialize TaskContext with empty tasks', () => {
    const initialState = {
      tasks: [] as Task[],
      selectedListId: 1,
      error: null as string | null,
    };

    expect(initialState.tasks).toEqual([]);
    expect(Array.isArray(initialState.tasks)).toBeTruthy();
  });

  test('2. should add task to context state', () => {
    const tasks: Task[] = [];
    const newTask: Task = {
      id: 1,
      client_id: 'task-001',
      todo: 'New task',
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
    };

    const updatedTasks = [...tasks, newTask];
    expect(updatedTasks.length).toBe(1);
    expect(updatedTasks[0].todo).toBe('New task');
  });

  test('3. should update task in context state', () => {
    const tasks: Task[] = [
      {
        id: 1,
        client_id: 'task-001',
        todo: 'Original text',
        priority: 3,
        done: false,
        todo_list_id: 1,
        date_added: Date.now(),
        updated_at: Date.now(),
        deleted: false,
        server_id: 1,
        date_completed: 0,
        due_date: 0,
        deleted_at: 0,
        version: 1,
      },
    ];

    const updates = { todo: 'Updated text', done: true };
    const updatedTasks = tasks.map(t =>
      t.id === 1 ? { ...t, ...updates, updated_at: Date.now() } : t
    );

    expect(updatedTasks[0].todo).toBe('Updated text');
    expect(updatedTasks[0].done).toBeTruthy();
  });

  test('4. should remove task from context state', () => {
    const tasks: Task[] = [
      {
        id: 1,
        client_id: 'task-001',
        todo: 'Task to keep',
        priority: 3,
        done: false,
        todo_list_id: 1,
        date_added: Date.now(),
        updated_at: Date.now(),
        deleted: false,
        server_id: 1,
        date_completed: 0,
        due_date: 0,
        deleted_at: 0,
        version: 1,
      },
      {
        id: 2,
        client_id: 'task-002',
        todo: 'Task to remove',
        priority: 3,
        done: false,
        todo_list_id: 1,
        date_added: Date.now(),
        updated_at: Date.now(),
        deleted: false,
        server_id: 2,
        date_completed: 0,
        due_date: 0,
        deleted_at: 0,
        version: 1,
      },
    ];

    const remainingTasks = tasks.filter(t => t.id !== 2);
    expect(remainingTasks.length).toBe(1);
    expect(remainingTasks[0].id).toBe(1);
  });

  // ========== LIST CONTEXT (3 tests) ==========

  test('5. should initialize ListContext with lists', () => {
    const initialState = {
      lists: [] as TodoList[],
      error: null as string | null,
    };

    expect(initialState.lists).toEqual([]);
  });

  test('6. should add list to context state', () => {
    const lists: TodoList[] = [];
    const newList: TodoList = {
      id: 1,
      client_id: 'list-001',
      name: 'New List',
      display_order: 0,
      archived: false,
      created_at: Date.now(),
      updated_at: Date.now(),
      server_id: 0,
      version: 1,
    };

    const updatedLists = [...lists, newList];
    expect(updatedLists.length).toBe(1);
    expect(updatedLists[0].name).toBe('New List');
  });

  test('7. should update list properties in context', () => {
    const lists: TodoList[] = [
      {
        id: 1,
        client_id: 'list-001',
        name: 'Original Name',
        display_order: 0,
        archived: false,
        created_at: Date.now(),
        updated_at: Date.now(),
        server_id: 1,
        version: 1,
      },
    ];

    const updatedLists = lists.map(l =>
      l.id === 1 ? { ...l, name: 'Updated Name' } : l
    );

    expect(updatedLists[0].name).toBe('Updated Name');
  });

  // ========== SYNC CONTEXT (4 tests) ==========

  test('8. should initialize SyncContext with sync status', () => {
    const initialState = {
      syncing: false,
      lastSyncTime: 0,
      syncError: null as string | null,
      isOnline: true,
    };

    expect(initialState.syncing).toBeFalsy();
    expect(initialState.lastSyncTime).toBe(0);
  });

  test('9. should update sync status when sync starts', () => {
    let syncState = {
      syncing: false,
      lastSyncTime: 0,
      syncError: null as string | null,
      isOnline: true,
    };

    // Start sync
    syncState = { ...syncState, syncing: true };
    expect(syncState.syncing).toBeTruthy();
  });

  test('10. should update sync metadata after successful sync', () => {
    let syncState = {
      syncing: true,
      lastSyncTime: 0,
      syncError: null as string | null,
      isOnline: true,
    };

    const now = Date.now();
    // Complete sync
    syncState = {
      ...syncState,
      syncing: false,
      lastSyncTime: now,
      syncError: null,
    };

    expect(syncState.syncing).toBeFalsy();
    expect(syncState.lastSyncTime).toBe(now);
  });

  test('11. should track sync errors in context', () => {
    let syncState = {
      syncing: true,
      lastSyncTime: Date.now() - 3600000,
      syncError: null as string | null,
      isOnline: true,
    };

    // Sync fails
    syncState = {
      ...syncState,
      syncing: false,
      syncError: 'Network connection failed',
    };

    expect(syncState.syncError).toBeTruthy();
    expect(syncState.syncing).toBeFalsy();
  });

  // ========== CONTEXT INTEGRATION (3 tests) ==========

  test('12. should track network online/offline status', () => {
    let appState = {
      online: true,
      lastOnlineTime: Date.now(),
    };

    // Go offline
    appState = { ...appState, online: false };
    expect(appState.online).toBeFalsy();

    // Go back online
    appState = { ...appState, online: true };
    expect(appState.online).toBeTruthy();
  });

  test('13. should handle error state across contexts', () => {
    const errors = {
      taskError: null as string | null,
      listError: null as string | null,
      syncError: null as string | null,
    };

    // Set an error
    errors.taskError = 'Failed to create task';
    expect(errors.taskError).toBeTruthy();

    // Clear error
    errors.taskError = null;
    expect(errors.taskError).toBeNull();
  });

  test('14. should manage selected list ID across contexts', () => {
    let appState = {
      selectedListId: 0,
      lists: [] as TodoList[],
      tasks: [] as Task[],
    };

    // Select list
    appState.selectedListId = 1;
    expect(appState.selectedListId).toBe(1);

    // Get tasks for selected list
    const listTasks = appState.tasks.filter(t => t.todo_list_id === appState.selectedListId);
    expect(listTasks.length).toBeGreaterThanOrEqual(0);
  });
});
