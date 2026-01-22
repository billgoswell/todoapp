/**
 * State Management
 *
 * Centralized exports for contexts, providers, and hooks
 */

// Providers
export { AppProvider, default } from './AppProvider';

// Contexts
export {
  TaskContext,
  TaskContextProvider,
  type TaskContextType,
  type TaskContextProviderProps,
} from './context/TaskContext';
export {
  ListContext,
  ListContextProvider,
  type ListContextType,
  type ListContextProviderProps,
} from './context/ListContext';
export {
  SyncContext,
  SyncContextProvider,
  type SyncContextType,
  type SyncContextProviderProps,
  type SyncStatus,
} from './context/SyncContext';

// Hooks
export { useTasks, useTasksByList, useTaskById } from './hooks/useTasks';
export { useLists, useListById, useDefaultList } from './hooks/useLists';
export {
  useSync,
  useSyncAndReload,
  useAutoSync,
  useSyncStatusMessage,
  useSyncOnDemand,
} from './hooks/useSync';
