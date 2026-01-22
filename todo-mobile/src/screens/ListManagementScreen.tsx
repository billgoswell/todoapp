/**
 * List Management Screen
 *
 * Screen for creating and managing todo lists.
 * Allows create, rename, and delete lists.
 */

import React, { useState } from 'react';
import {
  View,
  Text,
  Pressable,
  StyleSheet,
  Alert,
  ScrollView,
  TextInput,
  FlatList,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { StackScreenProps } from '@react-navigation/stack';
import { useLists, useTasks } from '../state';
import { colors, spacing, typography } from '../theme';

type RootStackParamList = {
  Home: undefined;
  ListManagement: undefined;
};

type ListManagementScreenProps = StackScreenProps<RootStackParamList, 'ListManagement'>;

export const ListManagementScreen: React.FC<ListManagementScreenProps> = ({
  navigation,
}) => {
  const { lists, createList, updateList, deleteList, loading, error } = useLists();
  const { selectedListId, setSelectedListId } = useTasks();
  const [newListName, setNewListName] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  const handleSelectList = (listId: number) => {
    setSelectedListId(listId);
    navigation.goBack();
  };

  const handleCreateList = async () => {
    if (!newListName.trim()) {
      Alert.alert('Error', 'Please enter a list name');
      return;
    }

    setIsCreating(true);
    try {
      await createList(newListName.trim());
      setNewListName('');
      Alert.alert('Success', 'List created');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create list';
      Alert.alert('Error', errorMessage);
    } finally {
      setIsCreating(false);
    }
  };

  const handleRenameList = (listId: number, currentName: string) => {
    Alert.prompt(
      'Rename List',
      'Enter new name:',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Rename',
          onPress: async (newName) => {
            if (newName && newName.trim() && newName !== currentName) {
              try {
                await updateList(listId, { name: newName.trim() });
                Alert.alert('Success', 'List renamed');
              } catch (err) {
                Alert.alert('Error', 'Failed to rename list');
              }
            }
          },
        },
      ],
      'plain-text',
      currentName
    );
  };

  const handleDeleteList = (listId: number, listName: string) => {
    Alert.alert(
      'Delete List',
      `Are you sure you want to delete "${listName}"? All tasks in this list will be deleted.`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            try {
              await deleteList(listId);
              Alert.alert('Success', 'List deleted');
            } catch (err) {
              Alert.alert('Error', 'Failed to delete list');
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
        <Text style={styles.headerTitle}>Lists</Text>
        <View style={styles.headerSpacer} />
      </View>

      {/* Error Alert */}
      {error && (
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
        </View>
      )}

      <ScrollView style={styles.content}>
        {/* Create List Section */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Create New List</Text>
          <View style={styles.createListContainer}>
            <TextInput
              style={styles.input}
              placeholder="List name..."
              placeholderTextColor={colors.placeholder}
              value={newListName}
              onChangeText={setNewListName}
              editable={!isCreating}
            />
            <Pressable
              style={[styles.createButton, isCreating && styles.createButtonDisabled]}
              onPress={handleCreateList}
              disabled={isCreating}
            >
              <Text style={styles.createButtonText}>+</Text>
            </Pressable>
          </View>
        </View>

        {/* Lists Section */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Your Lists</Text>
          {lists.length === 0 ? (
            <View style={styles.emptyStateContainer}>
              <Text style={styles.emptyIcon}>📋</Text>
              <Text style={styles.emptyText}>No lists yet</Text>
            </View>
          ) : (
            <View>
              {lists.map((list, index) => (
                <Pressable
                  key={list.client_id || `list-${index}`}
                  style={[
                    styles.listItem,
                    index !== lists.length - 1 && styles.listItemBorder,
                    selectedListId === list.id && styles.listItemSelected,
                  ]}
                  onPress={() => handleSelectList(list.id || 0)}
                >
                  <View style={styles.listItemContent}>
                    <Text style={[
                      styles.listItemName,
                      selectedListId === list.id && styles.listItemNameSelected,
                    ]}>{list.name}</Text>
                    {selectedListId === list.id && (
                      <Text style={styles.selectedIndicator}>✓ Selected</Text>
                    )}
                  </View>

                  <View style={styles.listItemActions}>
                    <Pressable
                      style={styles.actionButton}
                      onPress={() => handleRenameList(list.id || 0, list.name)}
                    >
                      <Text style={styles.actionButtonText}>✏️</Text>
                    </Pressable>
                    <Pressable
                      style={styles.actionButton}
                      onPress={() => handleDeleteList(list.id || 0, list.name)}
                    >
                      <Text style={styles.deleteButtonText}>🗑️</Text>
                    </Pressable>
                  </View>
                </Pressable>
              ))}
            </View>
          )}
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
  errorContainer: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    backgroundColor: colors.errorLight,
    borderBottomWidth: 1,
    borderBottomColor: colors.error,
  },
  errorText: {
    ...typography.styles.bodySmall,
    color: colors.error,
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
  createListContainer: {
    flexDirection: 'row',
    gap: spacing.md,
    alignItems: 'center',
  },
  input: {
    flex: 1,
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderRadius: spacing.borderRadius.md,
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.background,
    ...typography.styles.body,
    color: colors.text,
  },
  createButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: colors.primary,
    justifyContent: 'center',
    alignItems: 'center',
  },
  createButtonDisabled: {
    opacity: 0.6,
  },
  createButtonText: {
    fontSize: 24,
    color: colors.white,
    fontWeight: '600',
  },
  emptyStateContainer: {
    alignItems: 'center',
    paddingVertical: spacing.xl,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: spacing.md,
  },
  emptyText: {
    ...typography.styles.body,
    color: colors.textSecondary,
  },
  listItem: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.lg,
  },
  listItemBorder: {
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  listItemContent: {
    flex: 1,
  },
  listItemName: {
    ...typography.styles.body,
    color: colors.text,
    marginBottom: spacing.xs,
  },
  listItemNameSelected: {
    color: colors.primary,
    fontWeight: '600',
  },
  listItemSelected: {
    backgroundColor: colors.primaryLight,
  },
  selectedIndicator: {
    ...typography.styles.caption,
    color: colors.primary,
  },
  listItemActions: {
    flexDirection: 'row',
    gap: spacing.md,
  },
  actionButton: {
    padding: spacing.md,
    borderRadius: spacing.borderRadius.md,
    backgroundColor: colors.backgroundSecondary,
  },
  actionButtonText: {
    fontSize: 16,
  },
  deleteButtonText: {
    fontSize: 16,
  },
});

export default ListManagementScreen;
