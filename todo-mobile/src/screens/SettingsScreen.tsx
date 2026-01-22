/**
 * Settings Screen
 *
 * Screen for configuring API key, server URL, and manual sync control.
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
  ActivityIndicator,
  Switch,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { StackScreenProps } from '@react-navigation/stack';
import { apiClient } from '../api';
import { useSync } from '../state';
import { db } from '../database';
import { colors, spacing, typography } from '../theme';

type RootStackParamList = {
  Home: undefined;
  Settings: undefined;
};

type SettingsScreenProps = StackScreenProps<RootStackParamList, 'Settings'>;

export const SettingsScreen: React.FC<SettingsScreenProps> = ({
  navigation,
}) => {
  const { performSync, testConnection, isOnline, isSyncing } = useSync();

  const [serverUrl, setServerUrl] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [showApiKey, setShowApiKey] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [stats, setStats] = useState<any>(null);

  // Load current settings on mount
  useEffect(() => {
    const loadSettings = async () => {
      const url = apiClient.getServerUrl();
      const key = apiClient.getApiKey();
      setServerUrl(url || '');
      setApiKey(key || '');

      // Load database stats
      const dbStats = await db.getStats();
      setStats(dbStats);
    };

    loadSettings();
  }, []);

  const handleSaveSettings = async () => {
    if (!serverUrl.trim() || !apiKey.trim()) {
      Alert.alert('Error', 'Please fill in server URL and API key');
      return;
    }

    setIsSaving(true);
    try {
      await apiClient.setCredentials(serverUrl.trim(), apiKey.trim());
      Alert.alert('Success', 'Settings saved');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to save settings';
      Alert.alert('Error', errorMessage);
    } finally {
      setIsSaving(false);
    }
  };

  const handleTestConnection = async () => {
    setIsTesting(true);
    try {
      const connected = await testConnection();
      if (connected) {
        Alert.alert('Success', 'Connected to server');
      } else {
        Alert.alert('Error', 'Unable to connect to server');
      }
    } catch (error) {
      Alert.alert('Error', 'Connection test failed');
    } finally {
      setIsTesting(false);
    }
  };

  const handleSync = async () => {
    const success = await performSync();
    if (success) {
      Alert.alert('Success', 'Sync completed');
      // Reload stats
      const dbStats = await db.getStats();
      setStats(dbStats);
    } else {
      Alert.alert('Error', 'Sync failed');
    }
  };

  const handleClearAllData = () => {
    Alert.alert(
      'Clear All Data',
      'This will delete all local tasks and lists. This cannot be undone.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Clear',
          style: 'destructive',
          onPress: async () => {
            try {
              await db.clearAllData();
              Alert.alert('Success', 'All data cleared');
              // Reload stats
              const dbStats = await db.getStats();
              setStats(dbStats);
            } catch (error) {
              Alert.alert('Error', 'Failed to clear data');
            }
          },
        },
      ]
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Pressable
          style={styles.backButton}
          onPress={() => navigation.goBack()}
        >
          <Text style={styles.backButtonText}>‹</Text>
        </Pressable>
        <Text style={styles.headerTitle}>Settings</Text>
        <View style={styles.headerSpacer} />
      </View>

      <ScrollView style={styles.content}>
        {/* Server Configuration */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Server Configuration</Text>

          <View style={styles.fieldContainer}>
            <Text style={styles.fieldLabel}>Server URL</Text>
            <TextInput
              style={styles.input}
              placeholder="http://localhost:8080/api/v1"
              placeholderTextColor={colors.placeholder}
              value={serverUrl}
              onChangeText={setServerUrl}
              editable={!isSaving}
              autoCapitalize="none"
              autoCorrect={false}
            />
          </View>

          <View style={styles.fieldContainer}>
            <View style={styles.fieldLabelContainer}>
              <Text style={styles.fieldLabel}>API Key</Text>
              <Pressable onPress={() => setShowApiKey(!showApiKey)}>
                <Text style={styles.fieldLabelAction}>
                  {showApiKey ? 'Hide' : 'Show'}
                </Text>
              </Pressable>
            </View>
            <TextInput
              style={styles.input}
              placeholder="Your API key"
              placeholderTextColor={colors.placeholder}
              value={apiKey}
              onChangeText={setApiKey}
              editable={!isSaving}
              secureTextEntry={!showApiKey}
              autoCapitalize="none"
              autoCorrect={false}
            />
          </View>

          <View style={styles.buttonContainer}>
            <Pressable
              style={[styles.button, styles.primaryButton, isSaving && styles.buttonDisabled]}
              onPress={handleSaveSettings}
              disabled={isSaving}
            >
              <Text style={styles.primaryButtonText}>
                {isSaving ? 'Saving...' : 'Save Settings'}
              </Text>
            </Pressable>

            <Pressable
              style={[styles.button, styles.secondaryButton, isTesting && styles.buttonDisabled]}
              onPress={handleTestConnection}
              disabled={isTesting || !serverUrl || !apiKey}
            >
              {isTesting ? (
                <ActivityIndicator size="small" color={colors.primary} />
              ) : (
                <Text style={styles.secondaryButtonText}>Test Connection</Text>
              )}
            </Pressable>
          </View>
        </View>

        {/* Sync & Status */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Sync & Status</Text>

          <View style={styles.statusContainer}>
            <View>
              <Text style={styles.statusLabel}>Connection</Text>
              <Text style={[styles.statusValue, { color: isOnline ? colors.success : colors.error }]}>
                {isOnline ? '🟢 Online' : '🔴 Offline'}
              </Text>
            </View>
            <View>
              <Text style={styles.statusLabel}>Syncing</Text>
              <Text style={styles.statusValue}>
                {isSyncing ? '↻ In progress' : '✓ Idle'}
              </Text>
            </View>
          </View>

          <Pressable
            style={[styles.button, styles.primaryButton, (isSyncing || !isOnline) && styles.buttonDisabled]}
            onPress={handleSync}
            disabled={isSyncing || !isOnline}
          >
            {isSyncing ? (
              <ActivityIndicator size="small" color={colors.white} />
            ) : (
              <Text style={styles.primaryButtonText}>Sync Now</Text>
            )}
          </Pressable>
        </View>

        {/* Database Statistics */}
        {stats && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Database</Text>

            <View style={styles.statsContainer}>
              <View style={styles.statItem}>
                <Text style={styles.statLabel}>Lists</Text>
                <Text style={styles.statValue}>{stats.listCount}</Text>
              </View>
              <View style={styles.statItem}>
                <Text style={styles.statLabel}>Tasks</Text>
                <Text style={styles.statValue}>{stats.taskCount}</Text>
              </View>
              <View style={styles.statItem}>
                <Text style={styles.statLabel}>Pending</Text>
                <Text style={styles.statValue}>{stats.unsyncedChanges}</Text>
              </View>
            </View>

            <Pressable
              style={[styles.button, styles.dangerButton]}
              onPress={handleClearAllData}
            >
              <Text style={styles.dangerButtonText}>Clear All Data</Text>
            </Pressable>
          </View>
        )}

        {/* About */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>About</Text>
          <Text style={styles.aboutText}>CommandLineTodo Mobile</Text>
          <Text style={styles.aboutSubText}>Version 1.0.0</Text>
          <Text style={styles.aboutDescription}>
            A simple, offline-first todo app with cloud sync
          </Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    backgroundColor: colors.backgroundSecondary,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  backButton: {
    padding: spacing.md,
    marginLeft: -spacing.md,
  },
  backButtonText: {
    fontSize: 28,
    color: colors.primary,
    fontWeight: '400',
  },
  headerTitle: {
    ...typography.styles.h3,
    color: colors.text,
    flex: 1,
    textAlign: 'center',
  },
  headerSpacer: {
    width: 40,
  },
  content: {
    flex: 1,
  },
  section: {
    paddingVertical: spacing.lg,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  sectionTitle: {
    ...typography.styles.h4,
    color: colors.text,
    marginBottom: spacing.lg,
  },
  fieldContainer: {
    marginBottom: spacing.lg,
  },
  fieldLabel: {
    ...typography.styles.label,
    color: colors.text,
    marginBottom: spacing.md,
  },
  fieldLabelContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  fieldLabelAction: {
    ...typography.styles.label,
    color: colors.primary,
  },
  input: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderRadius: spacing.borderRadius.md,
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.background,
    ...typography.styles.body,
    color: colors.text,
  },
  buttonContainer: {
    gap: spacing.md,
  },
  button: {
    paddingVertical: spacing.lg,
    borderRadius: spacing.borderRadius.md,
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: 44,
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  primaryButton: {
    backgroundColor: colors.primary,
  },
  primaryButtonText: {
    ...typography.styles.button,
    color: colors.white,
  },
  secondaryButton: {
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.background,
  },
  secondaryButtonText: {
    ...typography.styles.button,
    color: colors.primary,
  },
  dangerButton: {
    backgroundColor: colors.errorLight,
    borderWidth: 1,
    borderColor: colors.error,
  },
  dangerButtonText: {
    ...typography.styles.button,
    color: colors.error,
  },
  statusContainer: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    paddingVertical: spacing.lg,
    backgroundColor: colors.backgroundSecondary,
    borderRadius: spacing.borderRadius.md,
    marginBottom: spacing.lg,
  },
  statusLabel: {
    ...typography.styles.caption,
    color: colors.textTertiary,
    marginBottom: spacing.sm,
  },
  statusValue: {
    ...typography.styles.bodySemibold,
    color: colors.text,
  },
  statsContainer: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    paddingVertical: spacing.lg,
    backgroundColor: colors.backgroundSecondary,
    borderRadius: spacing.borderRadius.md,
    marginBottom: spacing.lg,
  },
  statItem: {
    alignItems: 'center',
  },
  statLabel: {
    ...typography.styles.caption,
    color: colors.textTertiary,
    marginBottom: spacing.sm,
  },
  statValue: {
    ...typography.styles.h3,
    color: colors.primary,
  },
  aboutText: {
    ...typography.styles.h4,
    color: colors.text,
    marginBottom: spacing.sm,
  },
  aboutSubText: {
    ...typography.styles.bodySmall,
    color: colors.textTertiary,
    marginBottom: spacing.md,
  },
  aboutDescription: {
    ...typography.styles.body,
    color: colors.textSecondary,
  },
});

export default SettingsScreen;
