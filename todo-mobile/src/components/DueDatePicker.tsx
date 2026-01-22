/**
 * Due Date Picker Component
 *
 * Calendar-based date picker modal for selecting task due dates.
 * Features: calendar view, month navigation, today highlight, date clearing.
 */

import React, { useState } from 'react';
import {
  View,
  Text,
  Modal,
  TouchableOpacity,
  StyleSheet,
  FlatList,
  SafeAreaView,
  ScrollView,
} from 'react-native';
import {
  format,
  startOfMonth,
  endOfMonth,
  eachDayOfInterval,
  isSameDay,
  addMonths,
  subMonths,
  startOfDay,
  isAfter,
  isBefore,
} from 'date-fns';
import { colors, spacing, typography } from '../theme';

interface DueDatePickerProps {
  value: number | null; // Unix timestamp in seconds
  onChange: (timestamp: number | null) => void;
  label?: string;
}

export const DueDatePicker: React.FC<DueDatePickerProps> = ({
  value,
  onChange,
  label = 'Due Date',
}) => {
  const [showPicker, setShowPicker] = useState(false);
  const [currentMonth, setCurrentMonth] = useState<Date>(new Date());

  const selectedDate = value ? new Date(value * 1000) : null;
  const today = new Date();

  // Get all days to display in calendar
  const monthStart = startOfMonth(currentMonth);
  const monthEnd = endOfMonth(currentMonth);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });

  // Get previous month's trailing days for padding
  const dayOfWeek = monthStart.getDay();
  const previousMonthEnd = new Date(monthStart);
  previousMonthEnd.setDate(0);

  const previousDays: Date[] = [];
  for (let i = dayOfWeek - 1; i >= 0; i--) {
    const date = new Date(previousMonthEnd);
    date.setDate(previousMonthEnd.getDate() - i);
    previousDays.push(date);
  }

  // Combine all days
  let allDays = [...previousDays, ...daysInMonth];

  // Pad with next month's days to complete weeks
  while (allDays.length % 7 !== 0) {
    const lastDay = allDays[allDays.length - 1];
    const nextDay = new Date(lastDay);
    nextDay.setDate(lastDay.getDate() + 1);
    allDays.push(nextDay);
  }

  const handleSelectDate = (date: Date) => {
    const midnight = startOfDay(date);
    const timestamp = Math.floor(midnight.getTime() / 1000);
    onChange(timestamp);
    setShowPicker(false);
  };

  const handleClearDate = () => {
    onChange(null);
    setShowPicker(false);
  };

  const handleToday = () => {
    handleSelectDate(today);
  };

  const handlePreviousMonth = () => {
    setCurrentMonth(subMonths(currentMonth, 1));
  };

  const handleNextMonth = () => {
    setCurrentMonth(addMonths(currentMonth, 1));
  };

  const isCurrentMonth = (date: Date) => {
    return (
      date.getMonth() === currentMonth.getMonth() &&
      date.getFullYear() === currentMonth.getFullYear()
    );
  };

  const isToday = (date: Date) => isSameDay(date, today);
  const isSelected = (date: Date) =>
    selectedDate && isSameDay(date, selectedDate);

  const displayText = selectedDate
    ? format(selectedDate, 'MMM d, yyyy')
    : 'No due date';

  // Check if selected date is in the past
  const isPastDue = selectedDate && isBefore(selectedDate, today);

  return (
    <>
      <TouchableOpacity
        style={[
          styles.triggerButton,
          { borderColor: isPastDue ? colors.error : colors.border },
        ]}
        onPress={() => setShowPicker(true)}
      >
        <Text style={[styles.label, { color: colors.textSecondary }]}>
          {label}
        </Text>
        <View style={styles.valueContainer}>
          <Text
            style={[
              styles.value,
              {
                color: isPastDue ? colors.error : colors.text,
              },
            ]}
          >
            {displayText}
          </Text>
          {isPastDue && (
            <Text style={[styles.overdueLabel, { color: colors.error }]}>
              Overdue
            </Text>
          )}
        </View>
      </TouchableOpacity>

      <Modal
        visible={showPicker}
        transparent
        animationType="slide"
        onRequestClose={() => setShowPicker(false)}
      >
        <SafeAreaView
          style={[styles.modal, { backgroundColor: colors.background }]}
        >
          {/* Header with close and done buttons */}
          <View style={[styles.header, { borderBottomColor: colors.border }]}>
            <TouchableOpacity
              onPress={() => setShowPicker(false)}
              hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
            >
              <Text style={[styles.headerButton, { color: colors.primary }]}>
                Cancel
              </Text>
            </TouchableOpacity>

            <Text style={[styles.headerTitle, { color: colors.text }]}>
              {format(currentMonth, 'MMMM yyyy')}
            </Text>

            <TouchableOpacity
              onPress={() => setShowPicker(false)}
              hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
            >
              <Text style={[styles.headerButton, { color: colors.primary }]}>
                Done
              </Text>
            </TouchableOpacity>
          </View>

          {/* Month navigation */}
          <View
            style={[
              styles.monthNavigation,
              { borderBottomColor: colors.border },
            ]}
          >
            <TouchableOpacity
              style={styles.navButton}
              onPress={handlePreviousMonth}
            >
              <Text style={[styles.navButtonText, { color: colors.primary }]}>
                ← Previous
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.todayButton}
              onPress={handleToday}
            >
              <Text style={[styles.todayButtonText, { color: colors.primary }]}>
                Today
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.navButton}
              onPress={handleNextMonth}
            >
              <Text style={[styles.navButtonText, { color: colors.primary }]}>
                Next →
              </Text>
            </TouchableOpacity>
          </View>

          {/* Weekday headers */}
          <View
            style={[
              styles.weekdayHeader,
              { borderBottomColor: colors.border },
            ]}
          >
            {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map((day) => (
              <Text
                key={day}
                style={[styles.weekday, { color: colors.textSecondary }]}
              >
                {day}
              </Text>
            ))}
          </View>

          {/* Calendar grid */}
          <ScrollView style={styles.calendarScroll}>
            <View style={styles.calendar}>
              {allDays.map((date, index) => {
                const dayInCurrentMonth = isCurrentMonth(date);
                const isDateSelected = isSelected(date);
                const isTodayDate = isToday(date);

                return (
                  <TouchableOpacity
                    key={index}
                    style={[
                      styles.dayCell,
                      isDateSelected && {
                        backgroundColor: colors.primary,
                      },
                      isTodayDate &&
                        !isDateSelected && {
                          borderColor: colors.primary,
                          borderWidth: 2,
                        },
                    ]}
                    onPress={() => handleSelectDate(date)}
                  >
                    <Text
                      style={[
                        styles.dayText,
                        {
                          color: isDateSelected
                            ? colors.white
                            : dayInCurrentMonth
                            ? colors.text
                            : colors.textTertiary,
                          fontWeight: isTodayDate ? '600' : '400',
                        },
                      ]}
                    >
                      {date.getDate()}
                    </Text>
                  </TouchableOpacity>
                );
              })}
            </View>
          </ScrollView>

          {/* Clear date button if date is selected */}
          {selectedDate && (
            <TouchableOpacity
              style={[
                styles.clearButton,
                { borderTopColor: colors.border },
              ]}
              onPress={handleClearDate}
            >
              <Text style={[styles.clearButtonText, { color: colors.error }]}>
                Clear Due Date
              </Text>
            </TouchableOpacity>
          )}
        </SafeAreaView>
      </Modal>
    </>
  );
};

const styles = StyleSheet.create({
  // Trigger button
  triggerButton: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.md,
    borderRadius: spacing.borderRadius.md,
    borderWidth: 1,
    backgroundColor: colors.background,
    marginBottom: spacing.md,
  },
  label: {
    ...typography.styles.bodySmall,
    marginBottom: spacing.xs,
  },
  valueContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  value: {
    ...typography.styles.body,
    flex: 1,
  },
  overdueLabel: {
    ...typography.styles.caption,
    marginLeft: spacing.sm,
  },

  // Modal
  modal: {
    flex: 1,
  },

  // Header
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
  },
  headerButton: {
    ...typography.styles.button,
    minWidth: 60,
    textAlign: 'center',
  },
  headerTitle: {
    ...typography.styles.h3,
    flex: 1,
    textAlign: 'center',
  },

  // Month navigation
  monthNavigation: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
  },
  navButton: {
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.md,
  },
  navButtonText: {
    ...typography.styles.body,
    textAlign: 'center',
  },
  todayButton: {
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.lg,
  },
  todayButtonText: {
    ...typography.styles.button,
    textAlign: 'center',
  },

  // Weekday header
  weekdayHeader: {
    flexDirection: 'row',
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
    backgroundColor: colors.backgroundSecondary,
  },
  weekday: {
    flex: 1 / 7,
    textAlign: 'center',
    ...typography.styles.caption,
  },

  // Calendar
  calendarScroll: {
    flex: 1,
  },
  calendar: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.md,
  },
  dayCell: {
    width: '14.2857%', // 1/7
    aspectRatio: 1,
    justifyContent: 'center',
    alignItems: 'center',
    marginVertical: spacing.xs / 2,
    borderRadius: spacing.borderRadius.sm,
  },
  dayText: {
    ...typography.styles.body,
  },

  // Clear button
  clearButton: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderTopWidth: 1,
    backgroundColor: colors.background,
  },
  clearButtonText: {
    ...typography.styles.button,
    textAlign: 'center',
  },
});

export default DueDatePicker;
