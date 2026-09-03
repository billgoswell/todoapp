/**
 * Task Context
 *
 * Manages all task-related state and provides methods for task operations.
 * Synchronizes with SQLite database and tracks changes for sync.
 */

import React, { createContext, useCallback, useState } from 'react';
import { Task } from '../../api/types';
import { db } from '../../database';
import { generateTaskClientId } from '../../utils/uuid';

export interface TaskContextType {
  // State
  tasks: Task[];
  tasksByListId: Record<number, Task[]>;
  selectedListId: number | null;
  loading: boolean;
  error: string | null;

  // Operations
  loadTasks: (listId?: number) => Promise<void>;
  createTask: (listId: number, todo: string, priority?: number, dueDate?: number) => Promise<Task>;
  updateTask: (id: number, updates: Partial<Task>) => Promise<Task | null>;
  deleteTask: (id: number) => Promise<void>;
  completeTask: (id: number) => Promise<void>;
  setSelectedListId: (listId: number | null) => void;
  clearError: () => void;
}

export const TaskContext = createContext<TaskContextType | undefined>(undefined);

export interface TaskContextProviderProps {
  children: React.ReactNode;
}

export const TaskContextProvider: React.FC<TaskContextProviderProps> = ({ children }) => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [tasksByListId, setTasksByListId] = useState<Record<number, Task[]>>({});
  const [selectedListId, setSelectedListId] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * Load tasks - optionally filtered by list ID
   */
  const loadTasks = useCallback(async (listId?: number) => {
    setLoading(true);
    setError(null);
    try {
      let loadedTasks: Task[];

      if (listId !== undefined) {
        loadedTasks = await db.tasks.findByListId(listId);
        setTasksByListId(prev => ({
          ...prev,
          [listId]: loadedTasks,
        }));
      } else {
        loadedTasks = await db.tasks.findAll();
        // Organize by list
        const organized: Record<number, Task[]> = {};
        loadedTasks.forEach(task => {
          if (!organized[task.todo_list_id]) {
            organized[task.todo_list_id] = [];
          }
          organized[task.todo_list_id].push(task);
        });
        setTasksByListId(organized);
      }

      setTasks(loadedTasks);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load tasks';
      setError(errorMessage);
      console.error('Error loading tasks:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * Create a new task
   */
  const createTask = useCallback(
    async (listId: number, todo: string, priority: number = 4, dueDate?: number) => {
      setError(null);
      try {
        const now = Math.floor(Date.now() / 1000);

        const newTask: Task = {
          client_id: generateTaskClientId(),
          todo_list_id: listId,
          todo,
          priority,
          done: false,
          date_added: now,
          due_date: dueDate || null,
          deleted: false,
          deleted_at: null,
          created_at: now,
          updated_at: now,
          version: 1,
        };

        // Create in database
        const createdTask = await db.tasks.create(newTask);

        // Record change for sync
        await db.changes.recordTaskCreate(createdTask.id!);

        // Update local state
        setTasks(prev => [createdTask, ...prev]);
        setTasksByListId(prev => ({
          ...prev,
          [listId]: [createdTask, ...(prev[listId] || [])],
        }));

        return createdTask;
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Failed to create task';
        setError(errorMessage);
        console.error('Error creating task:', err);
        throw err;
      }
    },
    []
  );

  /**
   * Update an existing task
   */
  const updateTask = useCallback(async (id: number, updates: Partial<Task>) => {
    setError(null);
    try {
      const currentTask = tasks.find(t => t.id === id);
      if (!currentTask) {
        throw new Error(`Task with id ${id} not found`);
      }

      const updatedTask: Task = {
        ...currentTask,
        ...updates,
        updated_at: Math.floor(Date.now() / 1000),
      };

      // Update in database
      await db.tasks.update(updatedTask);

      // Record change for sync
      await db.changes.recordTaskUpdate(id);

      // Update local state
      setTasks(prev =>
        prev.map(t => (t.id === id ? updatedTask : t))
      );

      setTasksByListId(prev => {
        const updated = { ...prev };
        Object.keys(updated).forEach(listId => {
          updated[parseInt(listId, 10)] = updated[parseInt(listId, 10)].map(t =>
            t.id === id ? updatedTask : t
          );
        });
        return updated;
      });

      return updatedTask;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update task';
      setError(errorMessage);
      console.error('Error updating task:', err);
      return null;
    }
  }, [tasks]);

  /**
   * Delete a task (soft delete)
   */
  const deleteTask = useCallback(async (id: number) => {
    setError(null);
    try {
      // Soft delete in database
      await db.tasks.delete(id);

      // Record change for sync
      await db.changes.recordTaskDelete(id);

      // Remove from local state
      setTasks(prev => prev.filter(t => t.id !== id));

      setTasksByListId(prev => {
        const updated = { ...prev };
        Object.keys(updated).forEach(listId => {
          updated[parseInt(listId, 10)] = updated[parseInt(listId, 10)].filter(
            t => t.id !== id
          );
        });
        return updated;
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to delete task';
      setError(errorMessage);
      console.error('Error deleting task:', err);
      throw err;
    }
  }, []);

  /**
   * Mark a task as complete
   */
  const completeTask = useCallback(
    async (id: number) => {
      const now = Math.floor(Date.now() / 1000);
      await updateTask(id, {
        done: true,
        date_completed: now,
      });
    },
    [updateTask]
  );

  /**
   * Clear error state
   */
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const value: TaskContextType = {
    // State
    tasks,
    tasksByListId,
    selectedListId,
    loading,
    error,

    // Operations
    loadTasks,
    createTask,
    updateTask,
    deleteTask,
    completeTask,
    setSelectedListId,
    clearError,
  };

  return <TaskContext.Provider value={value}>{children}</TaskContext.Provider>;
};
