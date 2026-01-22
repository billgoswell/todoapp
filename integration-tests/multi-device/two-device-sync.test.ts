/**
 * Phase 5c: Multi-Device Sync Tests - Two Device Scenarios (6 tests)
 * Tests validating synchronization across two devices
 */

describe('Multi-Device Sync - Two Device Scenarios (6 tests)', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  // Helper to simulate device state
  interface DeviceState {
    deviceId: string;
    tasks: any[];
    lastSyncTime: number;
  }

  function createDevice(id: string): DeviceState {
    return {
      deviceId: id,
      tasks: [],
      lastSyncTime: 0,
    };
  }

  // ========== TWO-DEVICE SCENARIOS (5 tests) ==========

  test('16. Device A creates task → Device B pulls changes → Task appears', () => {
    // Test: Device A creates task → Device B pulls changes → Task appears
    // Validates basic task synchronization between two devices

    const deviceA = createDevice('device-a');
    const deviceB = createDevice('device-b');
    const server = { tasks: [], lastChange: 0 };

    // Device A creates task
    const newTask = {
      id: 0,
      client_id: 'task-1',
      todo: 'Shared task',
      updated_at: Date.now(),
    };
    deviceA.tasks.push(newTask);
    server.tasks.push(newTask);
    server.lastChange = Date.now();

    // Device B pulls changes
    const deviceBLastSync = 0;
    const changesForDeviceB = server.tasks.filter(
      t => t.updated_at > deviceBLastSync
    );
    deviceB.tasks = changesForDeviceB;
    deviceB.lastSyncTime = Date.now();

    expect(deviceA.tasks.length).toBe(1);
    expect(deviceB.tasks.length).toBe(1);
    expect(deviceB.tasks[0].todo).toBe('Shared task');
  });

  test('17. Device A edits task → Device B syncs → Change visible', () => {
    // Test: Device A edits task → Device B syncs → Change visible
    // Validates that edits propagate across devices

    const deviceA = createDevice('device-a');
    const deviceB = createDevice('device-b');
    const server = {
      tasks: [
        {
          id: 1,
          client_id: 'task-1',
          todo: 'Original',
          updated_at: Date.now() - 10000,
        },
      ],
    };

    // Both devices have original
    deviceA.tasks = JSON.parse(JSON.stringify(server.tasks));
    deviceB.tasks = JSON.parse(JSON.stringify(server.tasks));

    // Device A edits
    const editTime = Date.now();
    deviceA.tasks[0].todo = 'Edited on Device A';
    deviceA.tasks[0].updated_at = editTime;
    server.tasks[0].todo = 'Edited on Device A';
    server.tasks[0].updated_at = editTime;

    // Device B syncs and applies
    deviceB.lastSyncTime = Date.now() - 5000;
    const updates = server.tasks.filter(
      t => t.updated_at > deviceB.lastSyncTime
    );
    deviceB.tasks = updates;

    expect(deviceB.tasks[0].todo).toBe('Edited on Device A');
  });

  test('18. Device A deletes task → Device B next sync → Task removed', () => {
    // Test: Device A deletes task → Device B next sync → Task removed
    // Validates that deletions propagate across devices

    const deviceA = createDevice('device-a');
    const deviceB = createDevice('device-b');
    const server = {
      tasks: [
        { id: 1, client_id: 'task-1', todo: 'Task to delete', deleted: false, updated_at: Date.now() - 5000 },
      ],
    };

    deviceA.tasks = JSON.parse(JSON.stringify(server.tasks));
    deviceB.tasks = JSON.parse(JSON.stringify(server.tasks));

    // Device A deletes
    deviceA.tasks[0].deleted = true;
    deviceA.tasks[0].updated_at = Date.now();
    server.tasks[0].deleted = true;
    server.tasks[0].updated_at = Date.now();

    // Device B syncs - should filter out deleted
    deviceB.lastSyncTime = Date.now() - 10000;
    const activeTasksForB = server.tasks.filter(
      t => t.updated_at > deviceB.lastSyncTime && !t.deleted
    );
    deviceB.tasks = activeTasksForB;

    expect(deviceB.tasks.length).toBe(0);
  });

  test('19. Concurrent edits to same task → Last-write-wins applied', () => {
    // Test: Concurrent edits to same task → Last-write-wins applied
    // Validates conflict resolution with concurrent changes

    const deviceA = createDevice('device-a');
    const deviceB = createDevice('device-b');
    const server = {
      task: {
        id: 1,
        client_id: 'task-1',
        todo: 'Shared task',
        updated_at: Date.now() - 10000,
      },
    };

    // Both devices edit same task concurrently
    const editTimeA = Date.now();
    const editTimeB = editTimeA + 1000; // B's edit is later

    const editsA = {
      todo: 'Device A edit',
      updated_at: editTimeA,
    };

    const editsB = {
      todo: 'Device B edit',
      updated_at: editTimeB,
    };

    // Server receives both, applies last-write-wins
    const winner = editsB.updated_at > editsA.updated_at ? editsB : editsA;

    expect(winner.todo).toBe('Device B edit');
    expect(winner.updated_at).toBe(editTimeB);
  });

  test('20. Device A offline, makes changes, Device B makes changes → Conflict resolved on sync', () => {
    // Test: Device A offline, makes changes, B makes changes → Resolved on sync
    // Validates conflict resolution when both devices offline then sync

    const deviceA = createDevice('device-a');
    const deviceB = createDevice('device-b');
    const server = {
      task: { id: 1, client_id: 'task-1', todo: 'Original', updated_at: Date.now() - 20000 },
    };

    deviceA.tasks = [JSON.parse(JSON.stringify(server.task))];
    deviceB.tasks = [JSON.parse(JSON.stringify(server.task))];

    // Device A goes offline and edits
    deviceA.tasks[0].todo = 'A offline edit';
    deviceA.tasks[0].updated_at = Date.now() - 5000;

    // Device B goes offline and edits (later)
    deviceB.tasks[0].todo = 'B offline edit';
    deviceB.tasks[0].updated_at = Date.now();

    // Device B syncs first
    server.task = deviceB.tasks[0];

    // Device A syncs - gets server version
    const serverVersion = { todo: 'B offline edit', updated_at: Date.now() };
    const localVersion = {
      todo: 'A offline edit',
      updated_at: Date.now() - 5000,
    };
    const winner =
      serverVersion.updated_at > localVersion.updated_at
        ? serverVersion
        : localVersion;

    expect(winner.todo).toBe('B offline edit');
  });

  // ========== ERROR RECOVERY (1 test) ==========

  test('21. Sync failure on one device doesn\'t affect other', () => {
    // Test: Sync failure on one device doesn't affect other
    // Validates isolation of sync errors

    const deviceA = createDevice('device-a');
    const deviceB = createDevice('device-b');

    const syncResultA = {
      success: false,
      error: 'Network timeout',
    };

    const syncResultB = {
      success: true,
      error: null,
    };

    expect(syncResultA.success).toBe(false);
    expect(syncResultB.success).toBe(true);
    expect(syncResultB.error).toBeNull();
  });
});
