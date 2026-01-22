/**
 * List Selector Component
 *
 * Dropdown-like component for selecting a list
 */

import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  Pressable,
  FlatList,
  Modal,
} from 'react-native';
import { TodoList } from '../api/types';
import { colors, spacing, typography } from '../theme';

interface ListSelectorProps {
  lists: TodoList[];
  selectedListId: number | null;
  onSelectList: (list: TodoList) => void;
  label?: string;
}

export const ListSelector: React.FC<ListSelectorProps> = ({
  lists,
  selectedListId,
  onSelectList,
  label = 'Select List',
}) => {
  const [showModal, setShowModal] = React.useState(false);
  const selectedList = lists.find(l => l.id === selectedListId);

  return (
    <>
      <View style={styles.container}>
        {label && <Text style={styles.label}>{label}</Text>}

        <Pressable
          style={styles.trigger}
          onPress={() => setShowModal(true)}
        >
          <Text style={styles.triggerText}>
            {selectedList?.name || 'Choose a list...'}
          </Text>
          <Text style={styles.chevron}>›</Text>
        </Pressable>
      </View>

      <Modal
        visible={showModal}
        transparent={true}
        animationType="slide"
        onRequestClose={() => setShowModal(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.modalContent}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>Select a List</Text>
              <Pressable
                onPress={() => setShowModal(false)}
                style={styles.closeButton}
              >
                <Text style={styles.closeButtonText}>×</Text>
              </Pressable>
            </View>

            <FlatList
              data={lists}
              keyExtractor={item => item.id?.toString() || item.client_id}
              renderItem={({ item }) => (
                <Pressable
                  style={[
                    styles.listItem,
                    selectedListId === item.id && styles.listItemSelected,
                  ]}
                  onPress={() => {
                    onSelectList(item);
                    setShowModal(false);
                  }}
                >
                  <Text
                    style={[
                      styles.listItemText,
                      selectedListId === item.id && styles.listItemTextSelected,
                    ]}
                  >
                    {item.name}
                  </Text>
                  {selectedListId === item.id && (
                    <Text style={styles.checkmark}>✓</Text>
                  )}
                </Pressable>
              )}
            />
          </View>
        </View>
      </Modal>
    </>
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
  trigger: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderRadius: spacing.borderRadius.md,
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.background,
  },
  triggerText: {
    ...typography.styles.body,
    color: colors.text,
    flex: 1,
  },
  chevron: {
    fontSize: 24,
    color: colors.textTertiary,
  },
  modalOverlay: {
    flex: 1,
    justifyContent: 'flex-end',
    backgroundColor: 'rgba(0, 0, 0, 0.4)',
  },
  modalContent: {
    maxHeight: '80%',
    backgroundColor: colors.background,
    borderTopLeftRadius: spacing.borderRadius.lg,
    borderTopRightRadius: spacing.borderRadius.lg,
    paddingBottom: spacing.lg,
  },
  modalHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.lg,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  modalTitle: {
    ...typography.styles.h3,
    color: colors.text,
  },
  closeButton: {
    padding: spacing.md,
    marginRight: -spacing.md,
  },
  closeButtonText: {
    fontSize: 28,
    color: colors.textSecondary,
    fontWeight: '300',
  },
  listItem: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  listItemSelected: {
    backgroundColor: colors.primaryLight,
  },
  listItemText: {
    ...typography.styles.body,
    color: colors.text,
    flex: 1,
  },
  listItemTextSelected: {
    ...typography.styles.bodySemibold,
    color: colors.primary,
  },
  checkmark: {
    fontSize: 18,
    color: colors.success,
    fontWeight: '600',
  },
});

export default ListSelector;
