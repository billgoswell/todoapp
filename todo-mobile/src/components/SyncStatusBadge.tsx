/**
 * Sync Status Badge
 *
 * Displays current sync status with icon and message
 */

import React, { useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ActivityIndicator,
  Pressable,
  Animated,
} from 'react-native';
import { useSync } from '../state';
import { colors, spacing, typography } from '../theme';
import { pulse } from '../utils/animations';

interface SyncStatusBadgeProps {
  onPress?: () => void;
  showPendingCount?: boolean;
}

export const SyncStatusBadge: React.FC<SyncStatusBadgeProps> = ({
  onPress,
  showPendingCount = true,
}) => {
  const { syncStatus, isSyncing, lastSyncError, unsyncedChanges } = useSync();
  const pulseAnim = React.useRef(pulse(1500, 0.6, 1)).current;

  // Start pulse animation when sync succeeds
  useEffect(() => {
    if (syncStatus === 'success') {
      pulseAnim.startPulse();
    } else {
      pulseAnim.stop();
    }

    return () => {
      pulseAnim.stop();
    };
  }, [syncStatus]);

  const getStatusColor = () => {
    switch (syncStatus) {
      case 'syncing':
        return colors.warning;
      case 'success':
        return colors.success;
      case 'error':
        return colors.error;
      default:
        return colors.textTertiary;
    }
  };

  const getStatusIcon = () => {
    if (isSyncing) return '↻';
    if (syncStatus === 'success') return '✓';
    if (syncStatus === 'error') return '⚠';
    return '◯';
  };

  const getStatusMessage = () => {
    if (isSyncing) return 'Syncing...';
    if (syncStatus === 'success') return 'Synced';
    if (syncStatus === 'error') return lastSyncError || 'Sync failed';
    return 'Not synced';
  };

  const statusColor = getStatusColor();

  return (
    <Pressable
      style={[styles.container, { borderColor: statusColor }]}
      onPress={onPress}
      disabled={!onPress}
    >
      <View style={styles.content}>
        {isSyncing ? (
          <ActivityIndicator
            size="small"
            color={statusColor}
            style={styles.icon}
          />
        ) : (
          <Animated.Text
            style={[
              styles.icon,
              { color: statusColor },
              syncStatus === 'success' && { opacity: pulseAnim.opacity },
            ]}
          >
            {getStatusIcon()}
          </Animated.Text>
        )}

        <View style={styles.textContainer}>
          <Text style={[styles.message, { color: statusColor }]}>
            {getStatusMessage()}
          </Text>

          {showPendingCount && unsyncedChanges > 0 && (
            <Text style={styles.pending}>
              {unsyncedChanges} {unsyncedChanges === 1 ? 'change' : 'changes'} pending
            </Text>
          )}
        </View>
      </View>
    </Pressable>
  );
};

const styles = StyleSheet.create({
  container: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: 1,
    backgroundColor: colors.background,
  },
  content: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
  },
  icon: {
    fontSize: 16,
    fontWeight: '600',
    minWidth: 24,
    textAlign: 'center',
  },
  textContainer: {
    flex: 1,
  },
  message: {
    ...typography.styles.bodySmallMedium,
    marginBottom: spacing.xs,
  },
  pending: {
    ...typography.styles.caption,
    color: colors.textTertiary,
  },
});

export default SyncStatusBadge;
