/**
 * useLists Hook
 *
 * Custom hook for accessing and managing list state.
 * Provides list operations and state from ListContext.
 */

import { useContext } from 'react';
import { ListContext, ListContextType } from '../context/ListContext';

/**
 * Use list state and operations
 * Must be used within ListContextProvider
 */
export const useLists = (): ListContextType => {
  const context = useContext(ListContext);

  if (context === undefined) {
    throw new Error('useLists must be used within a ListContextProvider');
  }

  return context;
};

/**
 * Get a specific list by ID
 */
export const useListById = (listId: number | null) => {
  const { lists } = useLists();

  if (!listId) {
    return null;
  }

  return lists.find(l => l.id === listId) || null;
};

/**
 * Get the first available list (for default selection)
 */
export const useDefaultList = () => {
  const { lists } = useLists();

  return lists.length > 0 ? lists[0] : null;
};
