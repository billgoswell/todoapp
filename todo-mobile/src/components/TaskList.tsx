/**
 * Task List Component
 *
 * Scrollable list of tasks with pull-to-refresh and empty state support
 */

import React from 'react';
import {
  View,
  FlatList,
  RefreshControl,
  StyleSheet,
  Text,
  ActivityIndicator,
} from 'react-native';
import { Task } from '../api/types';
import { TaskItem } from './TaskItem';
import { colors, spacing, typography } from '../theme';

interface TaskListProps {
  tasks: Task[];
  loading?: boolean;
  onRefresh?: () => Promise<void>;
  onTaskPress?: (task: Task) => void;
  onTaskComplete?: (task: Task) => void;
  onTaskDelete?: (task: Task) => void;
  showDueDate?: boolean;
  showPriority?: boolean;
  emptyMessage?: string;
}

export const TaskList: React.FC<TaskListProps> = ({
  tasks,
  loading = false,
  onRefresh,
  onTaskPress,
  onTaskComplete,
  onTaskDelete,
  showDueDate = true,
  showPriority = true,
  emptyMessage = 'No tasks yet. Create one to get started!',
}) => {
  const [isRefreshing, setIsRefreshing] = React.useState(false);

  const handleRefresh = React.useCallback(async () => {
    setIsRefreshing(true);
    try {
      await onRefresh?.();
    } finally {
      setIsRefreshing(false);
    }
  }, [onRefresh]);

  if (loading && tasks.length === 0) {
    return (
      <View style={styles.centerContainer}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  if (tasks.length === 0) {
    return (
      <View style={styles.emptyContainer}>
        <Text style={styles.emptyIcon}>📝</Text>
        <Text style={styles.emptyTitle}>No Tasks</Text>
        <Text style={styles.emptyMessage}>{emptyMessage}</Text>
      </View>
    );
  }

  return (
    <FlatList
      data={tasks}
      keyExtractor={(item) => item.client_id || item.id?.toString() || Math.random().toString()}
      renderItem={({ item }) => (
        <TaskItem
          task={item}
          onPress={onTaskPress}
          onComplete={onTaskComplete}
          onDelete={onTaskDelete}
          showDueDate={showDueDate}
          showPriority={showPriority}
        />
      )}
      refreshControl={
        onRefresh ? (
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={handleRefresh}
            colors={[colors.primary]}
            tintColor={colors.primary}
          />
        ) : undefined
      }
      contentContainerStyle={tasks.length === 0 ? styles.emptyListContent : undefined}
      scrollEnabled={tasks.length > 0}
    />
  );
};

const styles = StyleSheet.create({
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: colors.background,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    backgroundColor: colors.background,
  },
  emptyListContent: {
    flexGrow: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyIcon: {
    fontSize: 64,
    marginBottom: spacing.lg,
  },
  emptyTitle: {
    ...typography.styles.h3,
    color: colors.text,
    marginBottom: spacing.md,
    textAlign: 'center',
  },
  emptyMessage: {
    ...typography.styles.body,
    color: colors.textSecondary,
    textAlign: 'center',
  },
});

export default TaskList;
