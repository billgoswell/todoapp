/**
 * Priority Picker Component
 *
 * Allows selection of task priority (High, Medium, Low, None)
 */

import React from 'react';
import {
  View,
  Text,
  Pressable,
  StyleSheet,
} from 'react-native';
import { TASK_PRIORITIES, PRIORITY_LABELS, PRIORITY_COLORS } from '../utils/constants';
import { colors, spacing, typography } from '../theme';

interface PriorityPickerProps {
  value: number;
  onChange: (priority: number) => void;
  label?: string;
}

export const PriorityPicker: React.FC<PriorityPickerProps> = ({
  value,
  onChange,
  label = 'Priority',
}) => {
  const priorities = [
    { value: TASK_PRIORITIES.HIGH, label: PRIORITY_LABELS[1], color: PRIORITY_COLORS[1] },
    { value: TASK_PRIORITIES.MEDIUM, label: PRIORITY_LABELS[2], color: PRIORITY_COLORS[2] },
    { value: TASK_PRIORITIES.LOW, label: PRIORITY_LABELS[3], color: PRIORITY_COLORS[3] },
    { value: TASK_PRIORITIES.NONE, label: PRIORITY_LABELS[4], color: PRIORITY_COLORS[4] },
  ];

  return (
    <View style={styles.container}>
      {label && <Text style={styles.label}>{label}</Text>}

      <View style={styles.priorityGrid}>
        {priorities.map((priority) => (
          <Pressable
            key={priority.value}
            style={[
              styles.priorityButton,
              value === priority.value && styles.priorityButtonSelected,
            ]}
            onPress={() => onChange(priority.value)}
          >
            <View
              style={[
                styles.priorityColor,
                { backgroundColor: priority.color },
              ]}
            />
            <Text
              style={[
                styles.priorityLabel,
                value === priority.value && styles.priorityLabelSelected,
              ]}
            >
              {priority.label}
            </Text>
          </Pressable>
        ))}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
  },
  label: {
    ...typography.styles.label,
    color: colors.text,
    marginBottom: spacing.md,
  },
  priorityGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.md,
  },
  priorityButton: {
    flex: 1,
    minWidth: '45%',
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.md,
    borderRadius: spacing.borderRadius.md,
    borderWidth: 2,
    borderColor: colors.border,
    backgroundColor: colors.background,
  },
  priorityButtonSelected: {
    borderColor: colors.primary,
    backgroundColor: colors.primaryLight,
  },
  priorityColor: {
    width: 12,
    height: 12,
    borderRadius: 6,
    marginRight: spacing.md,
  },
  priorityLabel: {
    ...typography.styles.bodySmall,
    color: colors.textSecondary,
    flex: 1,
  },
  priorityLabelSelected: {
    ...typography.styles.bodySmallMedium,
    color: colors.primary,
  },
});

export default PriorityPicker;
