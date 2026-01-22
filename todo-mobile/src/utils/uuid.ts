/**
 * UUID generation utilities
 *
 * Used to generate unique client_id for tasks and lists
 * These IDs are used for sync conflict resolution
 */

import uuid from 'react-native-uuid';

/**
 * Generate a unique UUID v4
 * Used as client_id for tasks and lists
 */
export const generateUUID = (): string => {
  return uuid.v4() as string;
};

/**
 * Generate a prefixed UUID (e.g., "task-123e4567-e89b-12d3-a456-426614174000")
 */
export const generatePrefixedUUID = (prefix: string): string => {
  return `${prefix}-${generateUUID()}`;
};

/**
 * Check if a string is a valid UUID
 */
export const isValidUUID = (str: string): boolean => {
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
  return uuidRegex.test(str);
};

/**
 * Generate a UUID v4 for a task
 */
export const generateTaskClientId = (): string => {
  return generateUUID();
};

/**
 * Generate a UUID v4 for a list
 */
export const generateListClientId = (): string => {
  return generateUUID();
};
