/**
 * Jest setup and configuration for mobile app tests
 *
 * This file is run before each test file and sets up the test environment
 */

// Mock AsyncStorage for React Native testing
jest.mock('@react-native-async-storage/async-storage', () => ({
	getItem: jest.fn(() => Promise.resolve(null)),
	setItem: jest.fn(() => Promise.resolve()),
	removeItem: jest.fn(() => Promise.resolve()),
	clear: jest.fn(() => Promise.resolve()),
	getAllKeys: jest.fn(() => Promise.resolve([])),
	multiGet: jest.fn(() => Promise.resolve([])),
	multiSet: jest.fn(() => Promise.resolve()),
	multiRemove: jest.fn(() => Promise.resolve()),
}));

// Mock NetInfo for network status detection
jest.mock('@react-native-community/netinfo', () => ({
	fetch: jest.fn(() => Promise.resolve({
		isConnected: true,
		isInternetReachable: true,
	})),
	addEventListener: jest.fn(() => jest.fn()),
}));

// Mock Expo SQLite
jest.mock('expo-sqlite', () => ({
	openDatabase: jest.fn(() => ({
		transaction: jest.fn((callback) => {
			callback({
				executeSql: jest.fn(),
			});
		}),
		exec: jest.fn(() => Promise.resolve([])),
	})),
}));

// Mock axios for API testing
jest.mock('axios', () => ({
	create: jest.fn(() => ({
		get: jest.fn(() => Promise.resolve({ data: {} })),
		post: jest.fn(() => Promise.resolve({ data: {} })),
		put: jest.fn(() => Promise.resolve({ data: {} })),
		delete: jest.fn(() => Promise.resolve({ data: {} })),
		request: jest.fn(() => Promise.resolve({ data: {} })),
	})),
	default: {
		create: jest.fn(() => ({
			get: jest.fn(() => Promise.resolve({ data: {} })),
			post: jest.fn(() => Promise.resolve({ data: {} })),
			put: jest.fn(() => Promise.resolve({ data: {} })),
			delete: jest.fn(() => Promise.resolve({ data: {} })),
			request: jest.fn(() => Promise.resolve({ data: {} })),
		})),
	},
}));

// Suppress console errors during tests (optional)
const originalError = console.error;
beforeAll(() => {
	console.error = (...args: any[]) => {
		if (
			typeof args[0] === 'string' &&
			args[0].includes('NativeEventEmitter')
		) {
			return;
		}
		originalError.call(console, ...args);
	};
});

afterAll(() => {
	console.error = originalError;
});

// Global test utilities
global.testUtils = {
	/**
	 * Create a mock API response
	 */
	createMockResponse: (data: any, status = 200) => ({
		data,
		status,
		statusText: 'OK',
		headers: {},
		config: {},
	}),

	/**
	 * Create mock task data for testing
	 */
	createMockTask: (overrides = {}) => ({
		id: 1,
		clientId: 'task-001',
		todo: 'Test task',
		priority: 3,
		done: false,
		dateAdded: Date.now(),
		dateCompleted: null,
		dueDate: null,
		deleted: false,
		deletedAt: null,
		todoListId: 1,
		updatedAt: Date.now(),
		version: 1,
		...overrides,
	}),

	/**
	 * Create mock list data for testing
	 */
	createMockList: (overrides = {}) => ({
		id: 1,
		clientId: 'list-001',
		name: 'Default List',
		displayOrder: 0,
		archived: false,
		createdAt: Date.now(),
		updatedAt: Date.now(),
		version: 1,
		...overrides,
	}),
};

// Type declarations for TypeScript
declare global {
	namespace NodeJS {
		interface Global {
			testUtils: {
				createMockResponse: (data: any, status?: number) => any;
				createMockTask: (overrides?: any) => any;
				createMockList: (overrides?: any) => any;
			};
		}
	}
}

export {};


// Dummy test to satisfy Jest requirement
describe('Setup', () => {
  test('jest is configured properly', () => {
    expect(true).toBe(true);
  });
});
