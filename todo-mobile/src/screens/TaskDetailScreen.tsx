/**
 * Task Detail Screen
 *
 * Screen for creating a new task or editing an existing task.
 * Allows setting description, priority, and due date.
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TextInput,
  Pressable,
  StyleSheet,
  Alert,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { StackScreenProps } from '@react-navigation/stack';
import { useTasks, useTaskById } from '../state';
import { PriorityPicker } from '../components/PriorityPicker';
import { DueDatePicker } from '../components/DueDatePicker';
import { colors, spacing, typography } from '../theme';
import { TASK_PRIORITIES } from '../utils/constants';
import { now } from '../utils/timestamps';

type RootStackParamList = {
  Home: undefined;
  TaskDetail: { taskId?: number; listId: number };
};

type TaskDetailScreenProps = StackScreenProps<RootStackParamList, 'TaskDetail'>;

export const TaskDetailScreen: React.FC<TaskDetailScreenProps> = ({
  route,
  navigation,
}) => {
  const { taskId, listId } = route.params;
  const isEditing = !!taskId;

  const { createTask, updateTask } = useTasks();
  const existingTask = useTaskById(taskId || null);

  const [text, setText] = useState('');
  const [priority, setPriority] = useState(TASK_PRIORITIES.NONE);
  const [dueDate, setDueDate] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Load task if editing
  useEffect(() => {
    if (isEditing && existingTask) {
      setText(existingTask.todo);
      setPriority(existingTask.priority);
      setDueDate(existingTask.due_date || null);
    }
  }, [isEditing, existingTask]);

  const handleSave = async () => {
    if (!text.trim()) {
      Alert.alert('Error', 'Please enter a task description');
      return;
    }

    setIsLoading(true);
    try {
      if (isEditing && existingTask) {
        // Update existing task
        await updateTask(existingTask.id!, {
          todo: text.trim(),
          priority,
          due_date: dueDate,
        });
        Alert.alert('Success', 'Task updated');
      } else {
        // Create new task
        await createTask(listId, text.trim(), priority, dueDate || undefined);
        Alert.alert('Success', 'Task created');
      }

      navigation.goBack();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'An error occurred';
      Alert.alert('Error', errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={styles.flex}
      >
      <ScrollView style={styles.scrollView}>
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.headerTitle}>
            {isEditing ? 'Edit Task' : 'New Task'}
          </Text>
        </View>

        {/* Task Input */}
        <View style={styles.section}>
          <Text style={styles.sectionLabel}>Task Description</Text>
          <TextInput
            style={styles.input}
            placeholder="What needs to be done?"
            placeholderTextColor={colors.placeholder}
            value={text}
            onChangeText={setText}
            multiline={true}
            numberOfLines={4}
            editable={!isLoading}
          />
        </View>

        {/* Priority Picker */}
        <View style={styles.sectionContainer}>
          <PriorityPicker
            value={priority}
            onChange={setPriority}
            label="Priority"
          />
        </View>

        {/* Due Date Picker */}
        <View style={styles.section}>
          <DueDatePicker
            value={dueDate}
            onChange={setDueDate}
            label="Due Date"
          />
        </View>

        {/* Info */}
        <View style={styles.section}>
          <Text style={styles.infoText}>
            {isEditing ? 'Changes will be synced automatically' : 'Your task will be saved locally and synced when connected'}
          </Text>
        </View>

        {/* Buttons */}
        <View style={styles.buttonContainer}>
          <Pressable
            style={[styles.button, styles.cancelButton]}
            onPress={() => navigation.goBack()}
            disabled={isLoading}
          >
            <Text style={styles.cancelButtonText}>Cancel</Text>
          </Pressable>

          <Pressable
            style={[styles.button, styles.saveButton, isLoading && styles.saveButtonDisabled]}
            onPress={handleSave}
            disabled={isLoading}
          >
            <Text style={styles.saveButtonText}>
              {isLoading ? 'Saving...' : isEditing ? 'Update' : 'Create'}
            </Text>
          </Pressable>
        </View>
      </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  flex: {
    flex: 1,
  },
  scrollView: {
    flex: 1,
  },
  header: {
    paddingVertical: spacing.lg,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
    backgroundColor: colors.backgroundSecondary,
  },
  headerTitle: {
    ...typography.styles.h2,
    color: colors.text,
  },
  section: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  sectionContainer: {
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  sectionLabel: {
    ...typography.styles.label,
    color: colors.text,
    marginBottom: spacing.md,
  },
  input: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.md,
    borderRadius: spacing.borderRadius.md,
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.background,
    ...typography.styles.body,
    color: colors.text,
    minHeight: 100,
  },
  infoText: {
    ...typography.styles.bodySmall,
    color: colors.textTertiary,
    fontStyle: 'italic',
  },
  buttonContainer: {
    flexDirection: 'row',
    gap: spacing.md,
    paddingVertical: spacing.lg,
    paddingHorizontal: spacing.lg,
  },
  button: {
    flex: 1,
    paddingVertical: spacing.lg,
    borderRadius: spacing.borderRadius.md,
    justifyContent: 'center',
    alignItems: 'center',
  },
  cancelButton: {
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.background,
  },
  cancelButtonText: {
    ...typography.styles.button,
    color: colors.text,
  },
  saveButton: {
    backgroundColor: colors.primary,
  },
  saveButtonDisabled: {
    opacity: 0.6,
  },
  saveButtonText: {
    ...typography.styles.button,
    color: colors.white,
  },
});

export default TaskDetailScreen;
