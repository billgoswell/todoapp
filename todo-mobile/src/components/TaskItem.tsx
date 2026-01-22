/**
 * Task Item Component
 *
 * Displays a single task with priority indicator, completion status,
 * and action buttons.
 */

import React, { useEffect } from 'react';
import {
  View,
  Text,
  Pressable,
  StyleSheet,
  AccessibilityRole,
  Animated,
} from 'react-native';
import { Task } from '../api/types';
import { colors, spacing, typography } from '../theme';
import { formatDate } from '../utils/timestamps';
import { PRIORITY_COLORS, PRIORITY_LABELS } from '../utils/constants';
import { fadeIn, slideUp } from '../utils/animations';

interface TaskItemProps {
  task: Task;
  onPress?: (task: Task) => void;
  onComplete?: (task: Task) => void;
  onDelete?: (task: Task) => void;
  showDueDate?: boolean;
  showPriority?: boolean;
}

export const TaskItem: React.FC<TaskItemProps> = ({
  task,
  onPress,
  onComplete,
  onDelete,
  showDueDate = true,
  showPriority = true,
}) => {
  const isOverdue = task.due_date && task.due_date < Math.floor(Date.now() / 1000) && !task.done;

  // Animation values
  const fadeInAnim = React.useRef(new Animated.Value(0)).current;
  const slideUpAnim = React.useRef(new Animated.Value(30)).current;

  // Animate on mount
  useEffect(() => {
    Animated.parallel([
      Animated.timing(fadeInAnim, {
        toValue: 1,
        duration: 300,
        useNativeDriver: true,
      }),
      Animated.timing(slideUpAnim, {
        toValue: 0,
        duration: 400,
        useNativeDriver: true,
      }),
    ]).start();
  }, []);

  return (
    <Animated.View
      style={{
        opacity: fadeInAnim,
        transform: [{ translateY: slideUpAnim }],
      }}
    >
      <Pressable
      style={[styles.container, task.done && styles.completed]}
      onPress={() => onPress?.(task)}
      accessibilityRole="button"
      accessibilityLabel={`${task.todo}${task.done ? ', completed' : ''}`}
      accessibilityHint={`Priority: ${PRIORITY_LABELS[task.priority as keyof typeof PRIORITY_LABELS] || 'None'}`}
    >
      <View style={styles.content}>
        {/* Priority indicator */}
        {showPriority && (
          <View
            style={[
              styles.priorityIndicator,
              {
                backgroundColor: PRIORITY_COLORS[task.priority as keyof typeof PRIORITY_COLORS],
              },
            ]}
          />
        )}

        {/* Main content */}
        <View style={styles.textContainer}>
          <Text
            style={[styles.taskText, task.done && styles.taskTextCompleted]}
            numberOfLines={2}
          >
            {task.todo}
          </Text>

          {/* Due date and priority label */}
          <View style={styles.metaContainer}>
            {showDueDate && task.due_date && (
              <Text
                style={[
                  styles.metaText,
                  isOverdue && styles.overdueText,
                ]}
              >
                {isOverdue ? '⚠ Overdue' : formatDate(task.due_date, 'MMM dd')}
              </Text>
            )}
            {showPriority && (
              <Text style={styles.metaText}>
                {PRIORITY_LABELS[task.priority as keyof typeof PRIORITY_LABELS] || 'None'}
              </Text>
            )}
          </View>
        </View>
      </View>

      {/* Actions */}
      <View style={styles.actions}>
        {onComplete && (
          <Pressable
            style={[styles.actionButton, styles.completeButton]}
            onPress={() => onComplete(task)}
            accessibilityRole="button"
            accessibilityLabel={task.done ? 'Mark incomplete' : 'Mark complete'}
          >
            <Text style={styles.actionButtonText}>
              {task.done ? '✓' : '○'}
            </Text>
          </Pressable>
        )}

        {onDelete && (
          <Pressable
            style={[styles.actionButton, styles.deleteButton]}
            onPress={() => onDelete(task)}
            accessibilityRole="button"
            accessibilityLabel="Delete task"
          >
            <Text style={styles.deleteButtonText}>×</Text>
          </Pressable>
        )}
      </View>
    </Pressable>
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
    backgroundColor: colors.background,
  },
  completed: {
    backgroundColor: colors.backgroundSecondary,
  },
  content: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'flex-start',
    marginRight: spacing.md,
  },
  priorityIndicator: {
    width: 4,
    height: 40,
    borderRadius: 2,
    marginRight: spacing.md,
  },
  textContainer: {
    flex: 1,
  },
  taskText: {
    ...typography.styles.body,
    color: colors.text,
    marginBottom: spacing.xs,
  },
  taskTextCompleted: {
    ...typography.styles.body,
    color: colors.textTertiary,
    textDecorationLine: 'line-through',
  },
  metaContainer: {
    flexDirection: 'row',
    gap: spacing.md,
  },
  metaText: {
    ...typography.styles.caption,
    color: colors.textTertiary,
  },
  overdueText: {
    color: colors.error,
    fontWeight: '600',
  },
  actions: {
    flexDirection: 'row',
    gap: spacing.sm,
  },
  actionButton: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: spacing.borderRadius.md,
    justifyContent: 'center',
    alignItems: 'center',
  },
  completeButton: {
    backgroundColor: colors.successLight,
  },
  actionButtonText: {
    fontSize: 20,
    fontWeight: '600',
    color: colors.success,
  },
  deleteButton: {
    backgroundColor: colors.errorLight,
  },
  deleteButtonText: {
    fontSize: 22,
    fontWeight: '600',
    color: colors.error,
  },
});

export default TaskItem;
