/**
 * useTasks Hook
 *
 * Custom hook for accessing and managing task state.
 * Provides task operations and state from TaskContext.
 */

import { useContext } from 'react';
import { TaskContext, TaskContextType } from '../context/TaskContext';

/**
 * Use task state and operations
 * Must be used within TaskContextProvider
 */
export const useTasks = (): TaskContextType => {
  const context = useContext(TaskContext);

  if (context === undefined) {
    throw new Error('useTasks must be used within a TaskContextProvider');
  }

  return context;
};

/**
 * Get tasks for a specific list
 */
export const useTasksByList = (listId: number | null) => {
  const { tasksByListId, loading, error } = useTasks();

  if (!listId) {
    return { tasks: [], loading, error };
  }

  return {
    tasks: tasksByListId[listId] || [],
    loading,
    error,
  };
};

/**
 * Get a specific task by ID
 */
export const useTaskById = (taskId: number | null) => {
  const { tasks } = useTasks();

  if (!taskId) {
    return null;
  }

  return tasks.find(t => t.id === taskId) || null;
};
