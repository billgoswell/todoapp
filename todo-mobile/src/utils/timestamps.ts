/**
 * Utility functions for handling Unix timestamps
 *
 * All timestamps in the system are Unix timestamps (seconds since epoch)
 * This ensures compatibility with the server and other clients
 */

import { format, fromUnixTime, isAfter } from 'date-fns';

/**
 * Get current timestamp in seconds since epoch
 */
export const now = (): number => {
  return Math.floor(Date.now() / 1000);
};

/**
 * Convert Unix timestamp to Date object
 */
export const toDate = (timestamp: number): Date => {
  return fromUnixTime(timestamp);
};

/**
 * Format Unix timestamp for display
 * @param timestamp Unix timestamp
 * @param formatStr date-fns format string (default: 'MMM dd, yyyy')
 */
export const formatDate = (timestamp: number, formatStr: string = 'MMM dd, yyyy'): string => {
  if (timestamp === 0 || !timestamp) return 'No date';
  try {
    return format(toDate(timestamp), formatStr);
  } catch (error) {
    console.error('Error formatting date:', error);
    return 'Invalid date';
  }
};

/**
 * Format Unix timestamp with time
 * @param timestamp Unix timestamp
 */
export const formatDateTime = (timestamp: number): string => {
  if (timestamp === 0 || !timestamp) return 'No date';
  try {
    return format(toDate(timestamp), 'MMM dd, yyyy HH:mm');
  } catch (error) {
    console.error('Error formatting datetime:', error);
    return 'Invalid date';
  }
};

/**
 * Check if a due date is overdue
 */
export const isOverdue = (dueDate: number | null | undefined): boolean => {
  if (!dueDate || dueDate === 0) return false;
  return isAfter(now(), dueDate);
};

/**
 * Check if a due date is today
 */
export const isToday = (dueDate: number): boolean => {
  if (!dueDate || dueDate === 0) return false;
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const dueDay = toDate(dueDate);
  dueDay.setHours(0, 0, 0, 0);
  return today.getTime() === dueDay.getTime();
};

/**
 * Check if a due date is tomorrow
 */
export const isTomorrow = (dueDate: number): boolean => {
  if (!dueDate || dueDate === 0) return false;
  const tomorrow = new Date();
  tomorrow.setHours(0, 0, 0, 0);
  tomorrow.setDate(tomorrow.getDate() + 1);
  const dueDay = toDate(dueDate);
  dueDay.setHours(0, 0, 0, 0);
  return tomorrow.getTime() === dueDay.getTime();
};

/**
 * Get relative time string (e.g., "2 hours ago", "in 3 days")
 */
export const getRelativeTime = (timestamp: number): string => {
  if (!timestamp || timestamp === 0) return 'No date';

  const currentTime = now();
  const diffInSeconds = Math.abs(timestamp - currentTime);
  const isPast = timestamp < currentTime;

  if (diffInSeconds < 60) {
    return isPast ? 'just now' : 'soon';
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60);
    return isPast ? `${minutes}m ago` : `in ${minutes}m`;
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600);
    return isPast ? `${hours}h ago` : `in ${hours}h`;
  } else {
    const days = Math.floor(diffInSeconds / 86400);
    return isPast ? `${days}d ago` : `in ${days}d`;
  }
};

/**
 * Add seconds to a timestamp
 */
export const addSeconds = (timestamp: number, seconds: number): number => {
  return timestamp + seconds;
};

/**
 * Add minutes to a timestamp
 */
export const addMinutes = (timestamp: number, minutes: number): number => {
  return timestamp + minutes * 60;
};

/**
 * Add hours to a timestamp
 */
export const addHours = (timestamp: number, hours: number): number => {
  return timestamp + hours * 3600;
};

/**
 * Add days to a timestamp
 */
export const addDays = (timestamp: number, days: number): number => {
  return timestamp + days * 86400;
};

/**
 * Start of day timestamp
 */
export const startOfDay = (timestamp: number = now()): number => {
  const date = toDate(timestamp);
  date.setHours(0, 0, 0, 0);
  return Math.floor(date.getTime() / 1000);
};

/**
 * End of day timestamp
 */
export const endOfDay = (timestamp: number = now()): number => {
  const date = toDate(timestamp);
  date.setHours(23, 59, 59, 999);
  return Math.floor(date.getTime() / 1000);
};
