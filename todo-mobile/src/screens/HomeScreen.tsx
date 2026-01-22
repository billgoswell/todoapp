/**
 * Home Screen
 *
 * Main screen displaying list of tasks for selected list.
 * Provides create, edit, complete, and delete functionality.
 */

import React, { useEffect, useCallback, useState, useMemo } from 'react';
import {
  View,
  Text,
  Pressable,
  StyleSheet,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { StackScreenProps } from '@react-navigation/stack';
import { useTasks, useLists, useSync, useSyncOnDemand } from '../state';
import { TaskList } from '../components/TaskList';
import { SearchAndFilter, SearchFilters } from '../components/SearchAndFilter';
import { SyncStatusBadge } from '../components/SyncStatusBadge';
import { OfflineIndicator } from '../components/OfflineIndicator';
import { colors, spacing, typography } from '../theme';
import { now } from '../utils/timestamps';

type RootStackParamList = {
  Home: undefined;
  TaskDetail: { taskId?: number; listId: number };
  ListManagement: undefined;
  Settings: undefined;
};

type HomeScreenProps = StackScreenProps<RootStackParamList, 'Home'>;

export const HomeScreen: React.FC<HomeScreenProps> = ({ navigation }) => {
  const { tasks, loadTasks, deleteTask, completeTask, selectedListId, setSelectedListId, error, clearError } = useTasks();
  const { lists, loadLists } = useLists();
  const { isOnline } = useSync();
  const { sync } = useSyncOnDemand();

  // Filter state
  const [filters, setFilters] = useState<SearchFilters>({
    searchText: '',
    status: 'all',
    priority: 'all',
  });

  // Load lists on mount
  useEffect(() => {
    loadLists();
  }, []);

  // Set first list as default on mount
  useEffect(() => {
    if (lists.length > 0 && selectedListId === null) {
      setSelectedListId(lists[0].id || null);
    }
  }, [lists]);

  // Load tasks when list changes
  useEffect(() => {
    if (selectedListId) {
      loadTasks(selectedListId);
    }
  }, [selectedListId]);

  // Get tasks for current list
  const currentListTasks = selectedListId
    ? tasks.filter(t => t.todo_list_id === selectedListId)
    : [];

  // Apply filters to tasks
  const filteredTasks = useMemo(() => {
    return currentListTasks.filter((task) => {
      // Search filter
      if (filters.searchText) {
        const searchLower = filters.searchText.toLowerCase();
        if (!task.todo.toLowerCase().includes(searchLower)) {
          return false;
        }
      }

      // Status filter
      if (filters.status !== 'all') {
        const currentTime = now();
        const isOverdue = task.due_date && task.due_date < currentTime && !task.done;

        if (filters.status === 'active' && task.done) {
          return false;
        }
        if (filters.status === 'completed' && !task.done) {
          return false;
        }
        if (filters.status === 'overdue' && !isOverdue) {
          return false;
        }
      }

      // Priority filter
      if (filters.priority !== 'all' && task.priority !== filters.priority) {
        return false;
      }

      return true;
    });
  }, [currentListTasks, filters]);

  const currentList = lists.find(l => l.id === selectedListId);

  const handleTaskPress = useCallback((taskId: number) => {
    navigation.navigate('TaskDetail', {
      taskId,
      listId: selectedListId || 0,
    });
  }, [selectedListId]);

  const handleCreateTask = useCallback(() => {
    if (!selectedListId) {
      Alert.alert('Error', 'Please select a list first');
      return;
    }
    navigation.navigate('TaskDetail', {
      listId: selectedListId,
    });
  }, [selectedListId]);

  const handleCompleteTask = useCallback(async (taskId: number) => {
    try {
      const task = tasks.find(t => t.id === taskId);
      if (task) {
        await completeTask(taskId);
      }
    } catch (err) {
      Alert.alert('Error', 'Failed to complete task');
    }
  }, [tasks, completeTask]);

  const handleDeleteTask = useCallback((taskId: number) => {
    Alert.alert(
      'Delete Task',
      'Are you sure you want to delete this task?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            try {
              await deleteTask(taskId);
            } catch (err) {
              Alert.alert('Error', 'Failed to delete task');
            }
          },
        },
      ]
    );
  }, [deleteTask]);

  const handleRefresh = useCallback(async () => {
    await sync();
  }, [sync]);

  const handleSyncPress = useCallback(async () => {
    await sync();
  }, [sync]);

  const handleSearchChange = useCallback((text: string) => {
    setFilters((prev) => ({
      ...prev,
      searchText: text,
    }));
  }, []);

  const handleStatusFilterChange = useCallback((status: 'all' | 'active' | 'completed' | 'overdue') => {
    setFilters((prev) => ({
      ...prev,
      status,
    }));
  }, []);

  const handlePriorityFilterChange = useCallback((priority: 'all' | 1 | 2 | 3 | 4) => {
    setFilters((prev) => ({
      ...prev,
      priority,
    }));
  }, []);

  const handleClearFilters = useCallback(() => {
    setFilters({
      searchText: '',
      status: 'all',
      priority: 'all',
    });
  }, []);

  return (
    <SafeAreaView style={styles.container}>
      {/* Offline Indicator */}
      <OfflineIndicator />

      {/* List selector header */}
      <View style={styles.headerContainer}>
        <Text style={styles.headerLabel}>List:</Text>
        <Pressable
          style={styles.listSelectorButton}
          onPress={() => navigation.navigate('ListManagement')}
        >
          <Text style={styles.listSelectorButtonText}>
            {currentList?.name || 'Select List'}
          </Text>
          <Text style={styles.chevron}>›</Text>
        </Pressable>
      </View>

      {/* Sync Status Badge */}
      <SyncStatusBadge onPress={handleSyncPress} />

      {/* Search and Filter */}
      {selectedListId && (
        <SearchAndFilter
          filters={filters}
          onSearchChange={handleSearchChange}
          onStatusFilterChange={handleStatusFilterChange}
          onPriorityFilterChange={handlePriorityFilterChange}
          onClearFilters={handleClearFilters}
        />
      )}

      {/* Error Alert */}
      {error && (
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
          <Pressable onPress={clearError} style={styles.errorDismiss}>
            <Text style={styles.errorDismissText}>Dismiss</Text>
          </Pressable>
        </View>
      )}

      {/* Task List */}
      {selectedListId ? (
        <TaskList
          tasks={filteredTasks}
          loading={false}
          onRefresh={handleRefresh}
          onTaskPress={() => {}}
          onTaskComplete={(task) => handleCompleteTask(task.id!)}
          onTaskDelete={(task) => handleDeleteTask(task.id!)}
          emptyMessage={
            filteredTasks.length === 0 && currentListTasks.length > 0
              ? 'No tasks match your filters'
              : `No tasks in ${currentList?.name || 'this list'}`
          }
        />
      ) : (
        <View style={styles.noListContainer}>
          <Text style={styles.noListEmoji}>📋</Text>
          <Text style={styles.noListTitle}>No Lists</Text>
          <Text style={styles.noListMessage}>Create a list to get started</Text>
          <Pressable
            style={styles.createListButton}
            onPress={() => navigation.navigate('ListManagement')}
          >
            <Text style={styles.createListButtonText}>Create List</Text>
          </Pressable>
        </View>
      )}

      {/* FAB - Create Task */}
      <Pressable
        style={styles.fab}
        onPress={handleCreateTask}
      >
        <Text style={styles.fabText}>+</Text>
      </Pressable>

      {/* Right header buttons */}
      <View style={styles.headerRight}>
        <Pressable
          onPress={() => navigation.navigate('ListManagement')}
          style={({ pressed }) => [
            styles.headerButton,
            pressed && styles.headerButtonPressed,
          ]}
        >
          <Text style={styles.headerButtonText}>📋</Text>
        </Pressable>
        <Pressable
          onPress={() => navigation.navigate('Settings')}
          style={({ pressed }) => [
            styles.headerButton,
            pressed && styles.headerButtonPressed,
          ]}
        >
          <Text style={styles.headerButtonText}>⚙️</Text>
        </Pressable>
      </View>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  headerContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    backgroundColor: colors.background,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  headerLabel: {
    ...typography.styles.label,
    color: colors.textSecondary,
    marginRight: spacing.md,
  },
  listSelectorButton: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderRadius: spacing.borderRadius.md,
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.backgroundSecondary,
  },
  listSelectorButtonText: {
    ...typography.styles.body,
    color: colors.text,
    flex: 1,
  },
  chevron: {
    fontSize: 20,
    color: colors.textTertiary,
  },
  errorContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    backgroundColor: colors.errorLight,
    borderBottomWidth: 1,
    borderBottomColor: colors.error,
  },
  errorText: {
    ...typography.styles.bodySmall,
    color: colors.error,
    flex: 1,
  },
  errorDismiss: {
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.md,
  },
  errorDismissText: {
    ...typography.styles.label,
    color: colors.error,
  },
  noListContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
  },
  noListEmoji: {
    fontSize: 64,
    marginBottom: spacing.lg,
  },
  noListTitle: {
    ...typography.styles.h3,
    color: colors.text,
    marginBottom: spacing.md,
  },
  noListMessage: {
    ...typography.styles.body,
    color: colors.textSecondary,
    textAlign: 'center',
    marginBottom: spacing.lg,
  },
  createListButton: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.xl,
    borderRadius: spacing.borderRadius.md,
    backgroundColor: colors.primary,
  },
  createListButtonText: {
    ...typography.styles.button,
    color: colors.white,
  },
  fab: {
    position: 'absolute',
    bottom: spacing.lg,
    right: spacing.lg,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.primary,
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 5,
  },
  fabText: {
    fontSize: 32,
    color: colors.white,
    fontWeight: '600',
    lineHeight: 32,
  },
  headerRight: {
    position: 'absolute',
    top: 0,
    right: spacing.lg,
    flexDirection: 'row',
    gap: spacing.md,
    paddingVertical: spacing.md,
  },
  headerButton: {
    padding: spacing.md,
    borderRadius: spacing.borderRadius.md,
  },
  headerButtonPressed: {
    backgroundColor: colors.backgroundSecondary,
  },
  headerButtonText: {
    fontSize: 20,
  },
});

export default HomeScreen;
