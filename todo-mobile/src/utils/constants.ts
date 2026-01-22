/**
 * App-wide constants
 */

// Default values
export const DEFAULT_LIST_NAME = 'My Tasks';
export const DEFAULT_PRIORITY = 4; // None
export const DEFAULT_SYNC_INTERVAL = 60; // seconds

// Task priorities
export const TASK_PRIORITIES = {
  HIGH: 1,
  MEDIUM: 2,
  LOW: 3,
  NONE: 4,
} as const;

export const PRIORITY_LABELS = {
  [TASK_PRIORITIES.HIGH]: 'High',
  [TASK_PRIORITIES.MEDIUM]: 'Medium',
  [TASK_PRIORITIES.LOW]: 'Low',
  [TASK_PRIORITIES.NONE]: 'None',
} as const;

export const PRIORITY_COLORS = {
  [TASK_PRIORITIES.HIGH]: '#FF3B30',
  [TASK_PRIORITIES.MEDIUM]: '#FF9500',
  [TASK_PRIORITIES.LOW]: '#FFCC00',
  [TASK_PRIORITIES.NONE]: '#8E8E93',
} as const;

// Sync constants
export const SYNC_STATUS = {
  IDLE: 'idle',
  SYNCING: 'syncing',
  SUCCESS: 'success',
  ERROR: 'error',
} as const;

// Error messages
export const ERROR_MESSAGES = {
  NETWORK_ERROR: 'Network connection failed',
  API_ERROR: 'Server error occurred',
  INVALID_API_KEY: 'Invalid API key',
  SERVER_UNREACHABLE: 'Unable to reach server',
  DATABASE_ERROR: 'Database operation failed',
  SYNC_ERROR: 'Sync failed',
} as const;

// LocalStorage keys
export const STORAGE_KEYS = {
  API_KEY: 'api_key',
  SERVER_URL: 'server_url',
  LAST_SYNC_TIME: 'last_sync_time',
  SYNC_ENABLED: 'sync_enabled',
  AUTO_SYNC: 'auto_sync',
  SYNC_INTERVAL: 'sync_interval',
} as const;

// API Endpoints (relative to base URL)
export const API_ENDPOINTS = {
  HEALTH: '/health',
  SYNC_PULL: '/sync/pull',
  SYNC_PUSH: '/sync/push',
  TASKS: '/tasks',
  LISTS: '/lists',
} as const;

// Network status
export const NETWORK_STATUS = {
  ONLINE: 'online',
  OFFLINE: 'offline',
  UNKNOWN: 'unknown',
} as const;

// Change types for tracking
export const CHANGE_TYPES = {
  CREATE: 'create',
  UPDATE: 'update',
  DELETE: 'delete',
} as const;

// Entity types for change tracking
export const ENTITY_TYPES = {
  TASK: 'task',
  LIST: 'list',
} as const;

// UI Dimensions
export const DIMENSIONS = {
  SCREEN_PADDING: 16,
  BORDER_RADIUS: 8,
  SMALL_ICON_SIZE: 20,
  MEDIUM_ICON_SIZE: 24,
  LARGE_ICON_SIZE: 32,
} as const;

// Animation durations (ms)
export const ANIMATIONS = {
  FAST: 200,
  NORMAL: 300,
  SLOW: 500,
} as const;

// HTTP timeouts (ms)
export const TIMEOUTS = {
  NETWORK: 10000,
  SYNC: 30000,
} as const;

// Retry configuration
export const RETRY = {
  MAX_ATTEMPTS: 3,
  INITIAL_DELAY: 1000, // ms
  MAX_DELAY: 30000, // ms
  BACKOFF_MULTIPLIER: 2,
} as const;
