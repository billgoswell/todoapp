/**
 * Offline Indicator
 *
 * Banner displayed when device is offline
 */

import React from 'react';
import {
  View,
  Text,
  StyleSheet,
} from 'react-native';
import { useSync } from '../state';
import { colors, spacing, typography } from '../theme';

interface OfflineIndicatorProps {
  message?: string;
}

export const OfflineIndicator: React.FC<OfflineIndicatorProps> = ({
  message = 'You are offline. Changes will sync when connected.',
}) => {
  const { isOnline } = useSync();

  if (isOnline) {
    return null;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.icon}>📡</Text>
      <Text style={styles.message}>{message}</Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    backgroundColor: colors.warningLight,
    borderBottomWidth: 2,
    borderBottomColor: colors.warning,
    gap: spacing.md,
  },
  icon: {
    fontSize: 16,
  },
  message: {
    ...typography.styles.bodySmall,
    color: colors.warning,
    flex: 1,
  },
});

export default OfflineIndicator;
