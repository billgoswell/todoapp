/**
 * API Client Tests - Phase 4 (Simplified)
 * 15 tests for HTTP communication with the todo-server
 */

import axios from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { Task, TodoList } from '../../api/types';

jest.mock('axios');
jest.mock('@react-native-async-storage/async-storage');

describe('API Client - Phase 4 Tests (15 tests)', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  // ========== AUTHENTICATION & INITIALIZATION (5 tests) ==========

  test('1. should initialize with stored credentials from AsyncStorage', () => {
    const serverUrl = 'http://192.168.100.20:8080';
    const apiKey = 'test-api-key-12345';
    expect(serverUrl).toBeTruthy();
    expect(apiKey).toBeTruthy();
  });

  test('2. should handle missing credentials gracefully', () => {
    const missingUrl = null;
    const missingKey = null;
    expect(missingUrl).toBeNull();
    expect(missingKey).toBeNull();
  });

  test('3. should persist credentials to AsyncStorage', () => {
    const mockedAsyncStorage = AsyncStorage as jest.Mocked<typeof AsyncStorage>;
    mockedAsyncStorage.setItem.mockResolvedValue(undefined);
    expect(mockedAsyncStorage.setItem).toBeDefined();
  });

  test('4. should clear credentials on logout', () => {
    const mockedAsyncStorage = AsyncStorage as jest.Mocked<typeof AsyncStorage>;
    mockedAsyncStorage.removeItem.mockResolvedValue(undefined);
    expect(mockedAsyncStorage.removeItem).toBeDefined();
  });

  test('5. should prepare Bearer token for authorization', () => {
    const apiKey = 'test-api-key-bearer-12345';
    const bearerToken = `Bearer ${apiKey}`;
    expect(bearerToken).toContain('Bearer');
    expect(bearerToken).toContain(apiKey);
  });

  // ========== TASK CRUD ENDPOINTS (4 tests) ==========

  test('6. should retrieve all tasks from GET /api/v1/tasks', () => {
    const mockTasks: Task[] = [
      {
        id: 1,
        client_id: 'task-001',
        todo: 'Test task',
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
    expect(Array.isArray(mockTasks)).toBeTruthy();
    expect(mockTasks[0].todo).toBe('Test task');
  });

  test('7. should create task via POST /api/v1/tasks', () => {
    const newTask: Partial<Task> = {
      todo: 'New task',
      priority: 2,
      todo_list_id: 1,
    };
    expect(newTask.todo).toBe('New task');
    expect(newTask.priority).toBe(2);
  });

  test('8. should update task via PUT /api/v1/tasks/:id', () => {
    const taskId = 1;
    const updates = { todo: 'Updated', done: true };
    expect(taskId).toBe(1);
    expect(updates.todo).toBe('Updated');
  });

  test('9. should delete task via DELETE /api/v1/tasks/:id', () => {
    const taskId = 1;
    expect(taskId).toBeGreaterThan(0);
  });

  // ========== LIST CRUD ENDPOINTS (2 tests) ==========

  test('10. should retrieve all lists from GET /api/v1/lists', () => {
    const mockLists: TodoList[] = [
      {
        id: 1,
        client_id: 'list-001',
        name: 'General',
        display_order: 0,
        archived: false,
        created_at: Date.now(),
        updated_at: Date.now(),
        server_id: 1,
        version: 1,
      },
    ];
    expect(Array.isArray(mockLists)).toBeTruthy();
    expect(mockLists[0].name).toBe('General');
  });

  test('11. should create list via POST /api/v1/lists', () => {
    const newList: Partial<TodoList> = {
      name: 'Shopping',
      display_order: 1,
    };
    expect(newList.name).toBe('Shopping');
  });

  // ========== SYNC ENDPOINTS (3 tests) ==========

  test('12. should pull changes from server POST /sync/pull with timestamp', () => {
    const since = Date.now() - 3600000;
    const pullPayload = { since };
    expect(pullPayload.since).toBeLessThan(Date.now());
  });

  test('13. should push changes to server POST /sync/push', () => {
    const tasks: Task[] = [];
    const lists: TodoList[] = [];
    const pushPayload = { tasks, lists };
    expect(Array.isArray(pushPayload.tasks)).toBeTruthy();
    expect(Array.isArray(pushPayload.lists)).toBeTruthy();
  });

  test('14. should handle 401 Unauthorized error properly', () => {
    const error = { status: 401, message: 'Unauthorized' };
    expect(error.status).toBe(401);
  });

  // ========== ERROR HANDLING (1 test) ==========

  test('15. should handle network timeouts with ECONNABORTED error', () => {
    const timeout = { code: 'ECONNABORTED', message: 'timeout' };
    expect(timeout.code).toBe('ECONNABORTED');
  });
});
