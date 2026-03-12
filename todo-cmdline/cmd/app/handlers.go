package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/billgoswell/commandlinetodo/internal/handler"
	"github.com/billgoswell/commandlinetodo/internal/logger"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.viewport.Width = msg.Width
		m.viewport.Height = m.getViewportHeight(msg.Height)
	case tea.KeyMsg:
		logger.LogDebug("Key pressed", "key", msg.String(), "state", m.currentState)
		m.errorMsg = ""
		switch m.currentState {
		case StateListSelector:
			return m.handleListSelector(msg)
		case StateEditTask:
			return m.handleEditMode(msg)
		case StateDeleteConfirm:
			return m.handleDeleteConfirm(msg)
		case StateTaskInput, StatePrioritySelection, StateDueDateInput, StateListNameInput:
			return m.handleInputMode(msg)
		case StateMainBrowse:
			return m.handleMainKeyboard(msg)
		}
	case spinnerTickMsg:
		// Advance spinner frame if currently syncing
		if m.syncEnabled && m.syncStatus.syncing {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(SpinnerFrames)
			return m, tickSpinner() // Schedule next tick
		}
		return m, nil
	case syncCompleteMsg:
		m.syncStatus.syncing = false
		if msg.err != nil {
			m.syncStatus.errorMessage = msg.err.Error()
		} else {
			m.syncStatus.errorMessage = ""
			m.syncStatus.lastSyncTime = msg.syncTime
		}
		// Spinner stops automatically since syncing = false
		return m, nil
	}
	return m, nil
}

func (m *model) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		m.returnToMain()
		return m, nil
	case KeyEnter:
		editedText, err := validateTaskText(m.textInput.Value())
		if err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}
		if m.input.itemIndex >= 0 && m.input.itemIndex < len(m.items) {
			m.items[m.input.itemIndex].todo = editedText
			m.items[m.input.itemIndex].todoListID = m.currentListID
			// Get the list's clientID for stable sync reference
			if currentList, err := m.store.GetTodoListByID(m.currentListID); err == nil {
				m.items[m.input.itemIndex].listClientID = currentList.clientID
			}
			msg := handler.HandleTaskUpdate("update task", func() error {
				return m.store.UpdateItem(m.items[m.input.itemIndex])
			})
			if msg != nil {
				m.errorMsg = msg.Content
			} else {
				logger.LogDebug("Task updated", "id", m.items[m.input.itemIndex].id)
				m.invalidateCache()
			}
		}
		m.returnToMain()
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m *model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyY:
		actualIndex := m.getVisibleItemActualIndex(m.input.deleteIndex)
		if actualIndex >= 0 && actualIndex < len(m.items) {
			itemID := m.items[actualIndex].id
			msg := handler.HandleTaskDelete("delete task", func() error {
				return m.store.DeleteItem(itemID)
			})
			if msg != nil {
				m.errorMsg = msg.Content
			} else {
				logger.LogDebug("Task deleted", "id", itemID)
				m.items = append(m.items[:actualIndex], m.items[actualIndex+1:]...)
				if m.cursor >= m.getVisibleItemCount() && m.cursor > 0 {
					m.cursor--
				}
			}
			m.invalidateCache()
			m.sortItems()
		}
		m.returnToMain()
		return m, nil
	case KeyN, KeyEsc:
		m.returnToMain()
		return m, nil
	}
	return m, nil
}

func (m *model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentState {
	case StateTaskInput, StatePrioritySelection, StateDueDateInput:
		return m.handleTaskCreationFlow(msg)
	}

	key := msg.String()

	if key == KeyEsc {
		m.returnToMain()
		return m, nil
	}

	if key == KeyEnter {
		listName, err := validateListName(m.textInput.Value())
		if err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}

		if m.currentSubState == SubStateListRename {
			if m.input.listIndex < len(m.todoLists) {
				listID := m.todoLists[m.input.listIndex].id
				msg := handler.HandleListUpdate("rename list", func() error {
					return m.store.UpdateTodoListName(listID, listName)
				})
				if msg != nil {
					m.errorMsg = msg.Content
				} else {
					m.todoLists[m.input.listIndex].name = listName
				}
			}
		} else {
			newID, msg := handler.HandleListCreate("create list", func() (int, error) {
				return m.store.CreateTodoList(listName)
			})
			if msg != nil && msg.Level == handler.MessageError {
				m.errorMsg = msg.Content
				m.returnToMain()
				return m, nil
			}
			todolists, msg := handler.SafeDataStoreOperation("load lists", func() ([]todoList, error) {
				return m.store.GetTodoLists()
			})
			if msg != nil && msg.Level == handler.MessageError {
				m.errorMsg = msg.Content
				m.returnToMain()
				return m, nil
			}
			m.todoLists = todolists
			for i, list := range m.todoLists {
				if list.id == newID {
					m.currentListIndex = i
					m.currentListID = newID
					break
				}
			}
			m.cursor = 0
			m.scrollOffset = 0
		}
		m.returnToMain()
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *model) handleTaskCreationFlow(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == KeyEsc {
		if m.taskFlow.step > TaskFlowInputText {
			m.taskFlow.previousStep()
			m.setState(m.getStateForFlowStep(m.taskFlow.step), SubStateNone)
			m.textInput.Reset()
		} else {
			m.taskFlow.reset()
			m.returnToMain()
		}
		return m, nil
	}

	if key == KeyEnter {
		switch m.taskFlow.step {
		case TaskFlowInputText:
			taskText, err := validateTaskText(m.textInput.Value())
			if err != nil {
				m.errorMsg = err.Error()
				return m, nil
			}
			m.taskFlow.text = taskText
			m.taskFlow.priority = DefaultPriority
			m.taskFlow.nextStep()
			m.setState(m.getStateForFlowStep(m.taskFlow.step), SubStateNone)
			m.textInput.Reset()
			return m, nil

		case TaskFlowSelectPriority:
			m.taskFlow.nextStep()
			m.setState(m.getStateForFlowStep(m.taskFlow.step), SubStateNone)
			m.textInput.Reset()
			return m, nil

		case TaskFlowSetDueDate:
			dueDate := parseDueDate(m.textInput.Value())

			if m.currentSubState == SubStateEditDueDate && m.input.itemIndex >= 0 && m.input.itemIndex < len(m.items) {
				m.items[m.input.itemIndex].dueDate = dueDate
				m.items[m.input.itemIndex].todoListID = m.currentListID
				// Get the list's clientID for stable sync reference
				if currentList, err := m.store.GetTodoListByID(m.currentListID); err == nil {
					m.items[m.input.itemIndex].listClientID = currentList.clientID
				}
				msg := handler.HandleTaskUpdate("update task", func() error {
					return m.store.UpdateItem(m.items[m.input.itemIndex])
				})
				if msg != nil {
					m.errorMsg = msg.Content
				} else {
					m.invalidateCache()
				}
			} else {
				// Get the list's clientID for stable sync reference
				listClientID := ""
				if currentList, err := m.store.GetTodoListByID(m.currentListID); err == nil {
					listClientID = currentList.clientID
				}

				newTask := todoItem{
					done:         false,
					todo:         m.taskFlow.text,
					priority:     m.taskFlow.priority,
					dateAdded:    time.Now().Unix(),
					dueDate:      dueDate,
					todoListID:   m.currentListID,
					listClientID: listClientID,
				}
				msg := handler.HandleTaskUpdate("create task", func() error {
					return m.store.SaveItem(newTask)
				})
				if msg != nil {
					m.errorMsg = msg.Content
				} else {
					m.items = append(m.items, newTask)
					m.sortItems()
					m.cursor = 0
				}
			}

			m.taskFlow.reset()
			m.returnToMain()
			return m, nil
		}
		return m, nil
	}

	if m.taskFlow.step == TaskFlowSelectPriority {
		switch key {
		case KeyPriority1:
			m.taskFlow.priority = PriorityHigh
			return m, nil
		case KeyPriority2:
			m.taskFlow.priority = PriorityMedHigh
			return m, nil
		case KeyPriority3:
			m.taskFlow.priority = PriorityMed
			return m, nil
		case KeyPriority4:
			m.taskFlow.priority = PriorityLow
			return m, nil
		case KeyUp, KeyK:
			if m.taskFlow.priority > PriorityHigh {
				m.taskFlow.priority--
			}
			return m, nil
		case KeyDown, KeyJ:
			if m.taskFlow.priority < PriorityLow {
				m.taskFlow.priority++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *model) handleMainKeyboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyCtrlC, KeyQ:
		return m, tea.Quit
	case KeyL:
		m.input.listIndex = m.currentListIndex
		m.setState(StateListSelector, SubStateNone)
		return m, nil
	case KeyS:
		if m.syncEnabled {
			if syncStore, ok := m.store.(*SyncStore); ok {
				logger.LogInfo("Manual sync initiated by user")
				m.syncStatus.syncing = true
				m.syncStatus.errorMessage = ""
				m.spinnerFrame = 0 // Reset spinner

				// Create sync command that returns proper message
				syncCmd := func() tea.Msg {
					err := syncStore.FullSync()
					return syncCompleteMsg{
						err:      err,
						syncTime: time.Now().Unix(),
					}
				}

				return m, tea.Batch(syncCmd, tickSpinner())
			}
		}
		return m, nil
	case KeyUp, KeyK:
		if m.cursor > 0 {
			m.cursor--
		}
	case KeyDown, KeyJ:
		visibleCount := m.getVisibleItemCount()
		if m.cursor < visibleCount-1 {
			m.cursor++
		}
	case KeyEnter:
		if m.cursor < m.getVisibleItemCount() {
			m.toggleTaskDone(m.cursor)
		}
	case KeySpace:
		if m.cursor < m.getVisibleItemCount() {
			m.toggleTaskDone(m.cursor)
		}
	case KeyD:
		if m.cursor < m.getVisibleItemCount() {
			m.input.deleteIndex = m.cursor
			m.setState(StateDeleteConfirm, SubStateNone)
		}
	case KeyA:
		m.textInput.Reset()
		m.textInput.Focus()
		m.setState(StateTaskInput, SubStateNone)
	case KeyE:
		if m.cursor < m.getVisibleItemCount() {
			m.input.itemIndex = m.getVisibleItemActualIndex(m.cursor)
			m.textInput.SetValue(m.items[m.input.itemIndex].todo)
			m.textInput.Focus()
			m.setState(StateEditTask, SubStateNone)
		}
	case KeyT:
		if m.cursor < m.getVisibleItemCount() {
			m.input.itemIndex = m.getVisibleItemActualIndex(m.cursor)
			m.textInput.Reset()
			m.textInput.Focus()
			m.setState(StateDueDateInput, SubStateEditDueDate)
		}
	}
	return m, nil
}

func (m *model) toggleTaskDone(visibleIndex int) {
	actualIndex := m.getVisibleItemActualIndex(visibleIndex)
	if actualIndex < 0 || actualIndex >= len(m.items) {
		return
	}
	m.items[actualIndex].done = !m.items[actualIndex].done
	if m.items[actualIndex].done {
		m.items[actualIndex].dateCompleted = time.Now().Unix()
		logger.LogDebug("Task marked as done", "id", m.items[actualIndex].id)
	} else {
		m.items[actualIndex].dateCompleted = 0
		logger.LogDebug("Task marked as pending", "id", m.items[actualIndex].id)
	}
	msg := handler.HandleTaskUpdate("toggle task done", func() error {
		return m.store.UpdateItem(m.items[actualIndex])
	})
	if msg != nil && msg.Level == handler.MessageError {
		m.errorMsg = msg.Content
	} else {
		m.invalidateCache()
		m.sortItems()
	}
}

func (m *model) handleListSelector(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.currentSubState == SubStateListDeleteConfirm {
		return m.handleListDeleteConfirm(msg)
	}
	if m.currentSubState == SubStateListManage {
		return m.handleListManageMode(msg)
	}
	return m.handleListNavigation(msg)
}

func (m *model) handleListDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyY:
		m.deleteCurrentList()
	case KeyN, KeyEsc:
		m.currentSubState = SubStateListManage
	}
	return m, nil
}

func (m *model) handleListManageMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyR:
		m.enterRenameMode()
	case KeyD:
		m.currentSubState = SubStateListDeleteConfirm
	case KeyA:
		m.toggleListArchive()
	case KeyEsc:
		m.currentSubState = SubStateNone
	}
	return m, nil
}

func (m *model) handleListNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyUp, KeyK:
		if m.input.listIndex > 0 {
			m.input.listIndex--
		}
	case KeyDown, KeyJ:
		if m.input.listIndex < len(m.todoLists)-1 {
			m.input.listIndex++
		}
	case KeyM:
		if m.input.listIndex < len(m.todoLists) {
			m.currentSubState = SubStateListManage
		}
	case KeyEnter:
		m.handleListSelection()
	case KeyN:
		m.startCreateNewList()
	case KeyEsc:
		m.returnToMain()
	}
	return m, nil
}

func (m *model) deleteCurrentList() {
	if m.input.listIndex >= len(m.todoLists) {
		m.currentSubState = SubStateNone
		return
	}

	selectedListID := m.todoLists[m.input.listIndex].id
	msg := handler.HandleListDelete("delete list", func() error {
		return m.store.DeleteTodoList(selectedListID)
	})
	if msg != nil {
		m.errorMsg = msg.Content
		m.currentSubState = SubStateNone
		return
	}

	var remainingItems []todoItem
	for _, item := range m.items {
		if item.todoListID != selectedListID {
			remainingItems = append(remainingItems, item)
		}
	}
	m.items = remainingItems
	m.invalidateCache()

	m.todoLists = append(m.todoLists[:m.input.listIndex], m.todoLists[m.input.listIndex+1:]...)
	if m.input.listIndex >= len(m.todoLists) && m.input.listIndex > 0 {
		m.input.listIndex--
	}

	if selectedListID == m.currentListID && len(m.todoLists) > 0 {
		m.switchToList(0)
	}
	m.returnToMain()
}

func (m *model) enterRenameMode() {
	if m.input.listIndex >= len(m.todoLists) {
		return
	}
	m.textInput.SetValue(m.todoLists[m.input.listIndex].name)
	m.textInput.Focus()
	m.setState(StateListNameInput, SubStateListRename)
}

func (m *model) toggleListArchive() {
	if m.input.listIndex >= len(m.todoLists) {
		return
	}

	selectedList := m.todoLists[m.input.listIndex]
	if selectedList.archived {
		msg := handler.HandleListUpdate("unarchive list", func() error {
			return m.store.UnarchiveTodoList(selectedList.id)
		})
		if msg != nil {
			m.errorMsg = msg.Content
			return
		}
		m.todoLists[m.input.listIndex].archived = false
	} else {
		msg := handler.HandleListUpdate("archive list", func() error {
			return m.store.ArchiveTodoList(selectedList.id)
		})
		if msg != nil {
			m.errorMsg = msg.Content
			return
		}
		m.todoLists[m.input.listIndex].archived = true
		if selectedList.id == m.currentListID {
			m.switchToFirstNonArchivedList()
		}
	}
	m.currentSubState = SubStateNone
}

func (m *model) switchToFirstNonArchivedList() {
	for i, list := range m.todoLists {
		if !list.archived {
			m.switchToList(i)
			return
		}
	}
}

func (m *model) handleListSelection() {
	if m.input.listIndex == len(m.todoLists) {
		m.startCreateNewList()
		return
	}
	if m.input.listIndex < len(m.todoLists) {
		m.switchToList(m.input.listIndex)
		m.returnToMain()
	}
}

func (m *model) switchToList(index int) {
	if index < 0 || index >= len(m.todoLists) {
		return
	}
	m.currentListIndex = index
	m.currentListID = m.todoLists[index].id
	m.cursor = 0
	m.scrollOffset = 0
	m.invalidateCache()
}

func (m *model) startCreateNewList() {
	m.textInput.Reset()
	m.textInput.Focus()
	m.setState(StateListNameInput, SubStateListCreate)
}
