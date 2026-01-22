/**
 * Search and Filter Component
 *
 * Provides search and filtering functionality for tasks.
 * Features: text search, status filtering, priority filtering, filter clearing.
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TextInput,
  ScrollView,
  TouchableOpacity,
  StyleSheet,
  Pressable,
  Animated,
} from 'react-native';
import { colors, spacing, typography } from '../theme';
import { TASK_PRIORITIES, PRIORITY_LABELS, PRIORITY_COLORS } from '../utils/constants';

type FilterStatus = 'all' | 'active' | 'completed' | 'overdue';
type FilterPriority = 'all' | 1 | 2 | 3 | 4;

export interface SearchFilters {
  searchText: string;
  status: FilterStatus;
  priority: FilterPriority;
}

interface SearchAndFilterProps {
  filters: SearchFilters;
  onSearchChange: (text: string) => void;
  onStatusFilterChange: (status: FilterStatus) => void;
  onPriorityFilterChange: (priority: FilterPriority) => void;
  onClearFilters: () => void;
}

export const SearchAndFilter: React.FC<SearchAndFilterProps> = ({
  filters,
  onSearchChange,
  onStatusFilterChange,
  onPriorityFilterChange,
  onClearFilters,
}) => {
  const [showFilters, setShowFilters] = useState(false);
  const filterPanelHeightAnim = React.useRef(new Animated.Value(0)).current;
  const filterPanelOpacityAnim = React.useRef(new Animated.Value(0)).current;

  // Animate panel expand/collapse
  useEffect(() => {
    Animated.parallel([
      Animated.timing(filterPanelHeightAnim, {
        toValue: showFilters ? 1 : 0,
        duration: 300,
        useNativeDriver: false,
      }),
      Animated.timing(filterPanelOpacityAnim, {
        toValue: showFilters ? 1 : 0,
        duration: 200,
        useNativeDriver: true,
      }),
    ]).start();
  }, [showFilters]);

  // Count active filters
  const activeFilterCount =
    (filters.searchText ? 1 : 0) +
    (filters.status !== 'all' ? 1 : 0) +
    (filters.priority !== 'all' ? 1 : 0);

  const hasActiveFilters = activeFilterCount > 0;

  // Status filter buttons
  const statusFilters: { label: string; value: FilterStatus }[] = [
    { label: 'All', value: 'all' },
    { label: 'Active', value: 'active' },
    { label: 'Completed', value: 'completed' },
    { label: 'Overdue', value: 'overdue' },
  ];

  // Priority filter buttons
  const priorityFilters: { label: string; value: FilterPriority; color: string }[] = [
    { label: 'All', value: 'all', color: colors.gray500 },
    { label: 'High', value: TASK_PRIORITIES.HIGH as FilterPriority, color: PRIORITY_COLORS.HIGH },
    { label: 'Medium', value: TASK_PRIORITIES.MEDIUM as FilterPriority, color: PRIORITY_COLORS.MEDIUM },
    { label: 'Low', value: TASK_PRIORITIES.LOW as FilterPriority, color: PRIORITY_COLORS.LOW },
    { label: 'None', value: TASK_PRIORITIES.NONE as FilterPriority, color: PRIORITY_COLORS.NONE },
  ];

  return (
    <View style={styles.container}>
      {/* Search Input */}
      <View style={styles.searchContainer}>
        <Text style={[styles.searchIcon, { color: colors.textTertiary }]}>🔍</Text>
        <TextInput
          style={[styles.searchInput, { color: colors.text }]}
          placeholder="Search tasks..."
          placeholderTextColor={colors.placeholder}
          value={filters.searchText}
          onChangeText={onSearchChange}
        />
        {filters.searchText ? (
          <Pressable
            onPress={() => onSearchChange('')}
            hitSlop={{ top: 8, bottom: 8, left: 8, right: 8 }}
          >
            <Text style={[styles.clearIcon, { color: colors.textTertiary }]}>✕</Text>
          </Pressable>
        ) : null}
      </View>

      {/* Filter Toggle Button */}
      <TouchableOpacity
        style={[
          styles.filterToggle,
          hasActiveFilters && styles.filterToggleActive,
        ]}
        onPress={() => setShowFilters(!showFilters)}
      >
        <Text style={[styles.filterToggleText, {color: hasActiveFilters ? colors.primary : colors.textSecondary}]}>
          ⊗ Filters
        </Text>
        {activeFilterCount > 0 && (
          <View
            style={[
              styles.filterBadge,
              { backgroundColor: colors.primary },
            ]}
          >
            <Text style={styles.filterBadgeText}>{activeFilterCount}</Text>
          </View>
        )}
      </TouchableOpacity>

      {/* Expandable Filter Panel */}
      {showFilters && (
        <Animated.View
          style={[
            styles.filterPanel,
            {
              borderTopColor: colors.border,
              backgroundColor: colors.backgroundSecondary,
              opacity: filterPanelOpacityAnim,
            },
          ]}
        >
          {/* Status Filters */}
          <View style={styles.filterSection}>
            <Text style={[styles.filterSectionTitle, { color: colors.text }]}>
              Status
            </Text>
            <View style={styles.filterButtonRow}>
              {statusFilters.map((filter) => (
                <TouchableOpacity
                  key={filter.value}
                  style={[
                    styles.filterButton,
                    filters.status === filter.value && {
                      backgroundColor: colors.primary,
                    },
                    filters.status !== filter.value && {
                      backgroundColor: colors.background,
                      borderWidth: 1,
                      borderColor: colors.border,
                    },
                  ]}
                  onPress={() => onStatusFilterChange(filter.value)}
                >
                  <Text
                    style={[
                      styles.filterButtonText,
                      {
                        color:
                          filters.status === filter.value
                            ? colors.white
                            : colors.text,
                      },
                    ]}
                  >
                    {filter.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>

          {/* Priority Filters */}
          <View style={styles.filterSection}>
            <Text style={[styles.filterSectionTitle, { color: colors.text }]}>
              Priority
            </Text>
            <View style={styles.filterButtonRow}>
              {priorityFilters.map((filter) => (
                <TouchableOpacity
                  key={filter.value}
                  style={[
                    styles.filterButton,
                    filters.priority === filter.value && {
                      backgroundColor: colors.primary,
                      borderWidth: 0,
                    },
                    filters.priority !== filter.value && {
                      backgroundColor: colors.background,
                      borderWidth: 1,
                      borderColor: colors.border,
                    },
                  ]}
                  onPress={() => onPriorityFilterChange(filter.value)}
                >
                  {filter.value !== 'all' && (
                    <View
                      style={[
                        styles.priorityIndicator,
                        { backgroundColor: filter.color },
                      ]}
                    />
                  )}
                  <Text
                    style={[
                      styles.filterButtonText,
                      {
                        color:
                          filters.priority === filter.value
                            ? colors.white
                            : colors.text,
                      },
                    ]}
                  >
                    {filter.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>

          {/* Clear Filters Button */}
          {hasActiveFilters && (
            <TouchableOpacity
              style={[styles.clearButton, { borderTopColor: colors.border }]}
              onPress={() => {
                onClearFilters();
                setShowFilters(false);
              }}
            >
              <Text style={[styles.clearButtonText, { color: colors.error }]}>
                Clear All Filters
              </Text>
            </TouchableOpacity>
          )}
        </Animated.View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: colors.background,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },

  // Search
  searchContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    backgroundColor: colors.background,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  searchIcon: {
    fontSize: 16,
    marginRight: spacing.md,
  },
  searchInput: {
    flex: 1,
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.md,
    borderRadius: spacing.borderRadius.sm,
    backgroundColor: colors.backgroundSecondary,
    ...typography.styles.body,
  },
  clearIcon: {
    fontSize: 18,
    marginLeft: spacing.md,
    fontWeight: '600',
  },

  // Filter Toggle
  filterToggle: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.sm,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    backgroundColor: colors.background,
  },
  filterToggleActive: {
    backgroundColor: colors.primaryLight,
  },
  filterToggleText: {
    ...typography.styles.body,
    fontWeight: '600',
  },
  filterBadge: {
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs / 2,
    borderRadius: spacing.borderRadius.sm,
    minWidth: 24,
  },
  filterBadgeText: {
    ...typography.styles.caption,
    color: colors.white,
    fontWeight: '600',
    textAlign: 'center',
  },

  // Filter Panel
  filterPanel: {
    borderTopWidth: 1,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  filterSection: {
    marginBottom: spacing.lg,
  },
  filterSectionTitle: {
    ...typography.styles.label,
    marginBottom: spacing.md,
  },
  filterButtonRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
  },
  filterButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.xs,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: spacing.borderRadius.md,
  },
  filterButtonText: {
    ...typography.styles.bodySmall,
    fontWeight: '500',
  },
  priorityIndicator: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },

  // Clear Button
  clearButton: {
    paddingVertical: spacing.md,
    borderTopWidth: 1,
    marginHorizontal: spacing.md * -1,
    marginBottom: spacing.md * -1,
    paddingHorizontal: spacing.md,
  },
  clearButtonText: {
    ...typography.styles.button,
    textAlign: 'center',
  },
});

export default SearchAndFilter;
