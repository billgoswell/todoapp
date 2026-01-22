/**
 * List Context
 *
 * Manages all list-related state and provides methods for list operations.
 * Synchronizes with SQLite database and tracks changes for sync.
 */

import React, { createContext, useCallback, useState } from 'react';
import { TodoList } from '../../api/types';
import { db } from '../../database';
import { generateListClientId } from '../../utils/uuid';

export interface ListContextType {
  // State
  lists: TodoList[];
  loading: boolean;
  error: string | null;

  // Operations
  loadLists: (includeArchived?: boolean) => Promise<void>;
  createList: (name: string) => Promise<TodoList>;
  updateList: (id: number, updates: Partial<TodoList>) => Promise<TodoList | null>;
  deleteList: (id: number) => Promise<void>;
  archiveList: (id: number) => Promise<void>;
  unarchiveList: (id: number) => Promise<void>;
  reorderLists: (lists: TodoList[]) => Promise<void>;
  clearError: () => void;
}

export const ListContext = createContext<ListContextType | undefined>(undefined);

export interface ListContextProviderProps {
  children: React.ReactNode;
}

export const ListContextProvider: React.FC<ListContextProviderProps> = ({ children }) => {
  const [lists, setLists] = useState<TodoList[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * Load lists
   */
  const loadLists = useCallback(async (includeArchived: boolean = false) => {
    setLoading(true);
    setError(null);
    try {
      let loadedLists: TodoList[];

      if (includeArchived) {
        loadedLists = await db.lists.findAllIncludingArchived();
      } else {
        loadedLists = await db.lists.findAll();
      }

      setLists(loadedLists);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load lists';
      setError(errorMessage);
      console.error('Error loading lists:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * Create a new list
   */
  const createList = useCallback(async (name: string) => {
    setError(null);
    try {
      const now = Math.floor(Date.now() / 1000);

      const newList: TodoList = {
        client_id: generateListClientId(),
        name,
        display_order: lists.length,
        archived: false,
        created_at: now,
        updated_at: now,
        version: 1,
      };

      // Create in database
      const createdList = await db.lists.create(newList);

      // Record change for sync
      await db.changes.recordListCreate(createdList.id!);

      // Update local state
      setLists(prev => [...prev, createdList]);

      return createdList;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create list';
      setError(errorMessage);
      console.error('Error creating list:', err);
      throw err;
    }
  }, [lists.length]);

  /**
   * Update an existing list
   */
  const updateList = useCallback(async (id: number, updates: Partial<TodoList>) => {
    setError(null);
    try {
      const currentList = lists.find(l => l.id === id);
      if (!currentList) {
        throw new Error(`List with id ${id} not found`);
      }

      const updatedList: TodoList = {
        ...currentList,
        ...updates,
        updated_at: Math.floor(Date.now() / 1000),
      };

      // Update in database
      await db.lists.update(updatedList);

      // Record change for sync
      await db.changes.recordListUpdate(id);

      // Update local state
      setLists(prev =>
        prev.map(l => (l.id === id ? updatedList : l))
      );

      return updatedList;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update list';
      setError(errorMessage);
      console.error('Error updating list:', err);
      return null;
    }
  }, [lists]);

  /**
   * Delete a list (archive it)
   */
  const deleteList = useCallback(async (id: number) => {
    setError(null);
    try {
      // Archive the list
      await db.lists.archive(id);

      // Record change for sync
      await db.changes.recordListDelete(id);

      // Remove from local state
      setLists(prev => prev.filter(l => l.id !== id));
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to delete list';
      setError(errorMessage);
      console.error('Error deleting list:', err);
      throw err;
    }
  }, []);

  /**
   * Archive a list
   */
  const archiveList = useCallback(async (id: number) => {
    setError(null);
    try {
      await db.lists.archive(id);
      await db.changes.recordListUpdate(id);

      // Remove from list view
      setLists(prev => prev.filter(l => l.id !== id));
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to archive list';
      setError(errorMessage);
      console.error('Error archiving list:', err);
      throw err;
    }
  }, []);

  /**
   * Unarchive a list
   */
  const unarchiveList = useCallback(async (id: number) => {
    setError(null);
    try {
      await db.lists.unarchive(id);
      await db.changes.recordListUpdate(id);

      // Reload lists to get the unarchived one
      await loadLists(false);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to unarchive list';
      setError(errorMessage);
      console.error('Error unarchiving list:', err);
      throw err;
    }
  }, [loadLists]);

  /**
   * Reorder lists
   */
  const reorderLists = useCallback(async (reorderedLists: TodoList[]) => {
    setError(null);
    try {
      await db.lists.reorder(reorderedLists);

      // Record updates for each list
      for (const list of reorderedLists) {
        await db.changes.recordListUpdate(list.id!);
      }

      // Update local state
      setLists(reorderedLists);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to reorder lists';
      setError(errorMessage);
      console.error('Error reordering lists:', err);
      throw err;
    }
  }, []);

  /**
   * Clear error state
   */
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const value: ListContextType = {
    // State
    lists,
    loading,
    error,

    // Operations
    loadLists,
    createList,
    updateList,
    deleteList,
    archiveList,
    unarchiveList,
    reorderLists,
    clearError,
  };

  return <ListContext.Provider value={value}>{children}</ListContext.Provider>;
};
