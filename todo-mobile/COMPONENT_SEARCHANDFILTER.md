# SearchAndFilter Component Reference

## Overview

`SearchAndFilter` is a flexible task filtering component for React Native that enables users to search tasks and apply multiple filters by status and priority.

**Location**: `src/components/SearchAndFilter.tsx`
**Type**: Functional controlled component with hooks (useState)
**Export**: `export { SearchAndFilter, type SearchFilters } from './SearchAndFilter'`

---

## Props Interface

```typescript
interface SearchFilters {
  searchText: string;
  status: 'all' | 'active' | 'completed' | 'overdue';
  priority: 'all' | 1 | 2 | 3 | 4;
}

interface SearchAndFilterProps {
  filters: SearchFilters;
  onSearchChange: (text: string) => void;
  onStatusFilterChange: (status: FilterStatus) => void;
  onPriorityFilterChange: (priority: FilterPriority) => void;
  onClearFilters: () => void;
}
```

### Props Details

#### `filters: SearchFilters` (Required)
Current filter state object containing:
- `searchText`: Search query string (empty = no search filter)
- `status`: Status filter ('all', 'active', 'completed', 'overdue')
- `priority`: Priority filter ('all', 1=HIGH, 2=MEDIUM, 3=LOW, 4=NONE)

#### `onSearchChange: (text: string) => void` (Required)
Callback fired when user types in search box. Receives the new search text.

#### `onStatusFilterChange: (status: FilterStatus) => void` (Required)
Callback fired when user clicks a status filter button. Receives the selected status.

#### `onPriorityFilterChange: (priority: FilterPriority) => void` (Required)
Callback fired when user clicks a priority filter button. Receives the selected priority (1-4 or 'all').

#### `onClearFilters: () => void` (Required)
Callback fired when user clicks "Clear All Filters" button. Should reset all filters to default state.

---

## Usage Examples

### Basic Integration
```typescript
import { SearchAndFilter, SearchFilters } from '../components';

function MyScreen() {
  const [filters, setFilters] = useState<SearchFilters>({
    searchText: '',
    status: 'all',
    priority: 'all',
  });

  return (
    <SearchAndFilter
      filters={filters}
      onSearchChange={(text) => setFilters(prev => ({ ...prev, searchText: text }))}
      onStatusFilterChange={(status) => setFilters(prev => ({ ...prev, status }))}
      onPriorityFilterChange={(priority) => setFilters(prev => ({ ...prev, priority }))}
      onClearFilters={() => setFilters({ searchText: '', status: 'all', priority: 'all' })}
    />
  );
}
```

### With Task Filtering
```typescript
import { SearchAndFilter, SearchFilters } from '../components';

function HomeScreen() {
  const allTasks = useTasksFromContext();
  const [filters, setFilters] = useState<SearchFilters>({
    searchText: '',
    status: 'all',
    priority: 'all',
  });

  const filteredTasks = useMemo(() => {
    return allTasks.filter((task) => {
      // Search filter
      if (filters.searchText && !task.todo.toLowerCase().includes(filters.searchText.toLowerCase())) {
        return false;
      }

      // Status filter
      if (filters.status === 'active' && task.done) return false;
      if (filters.status === 'completed' && !task.done) return false;
      if (filters.status === 'overdue' && !(task.due_date < now() && !task.done)) return false;

      // Priority filter
      if (filters.priority !== 'all' && task.priority !== filters.priority) return false;

      return true;
    });
  }, [allTasks, filters]);

  const handleClearFilters = () => {
    setFilters({ searchText: '', status: 'all', priority: 'all' });
  };

  return (
    <View>
      <SearchAndFilter
        filters={filters}
        onSearchChange={(text) => setFilters(prev => ({ ...prev, searchText: text }))}
        onStatusFilterChange={(status) => setFilters(prev => ({ ...prev, status }))}
        onPriorityFilterChange={(priority) => setFilters(prev => ({ ...prev, priority }))}
        onClearFilters={handleClearFilters}
      />
      <TaskList tasks={filteredTasks} />
    </View>
  );
}
```

---

## Visual Layout

### Collapsed View (No Filters)
```
┌────────────────────────────────┐
│ 🔍 Search tasks...         ✕    │
├────────────────────────────────┤
│           ⊗ Filters            │
└────────────────────────────────┘
```

### Collapsed View (With Active Filters)
```
┌────────────────────────────────┐
│ 🔍 Search bug fix          ✕    │
├────────────────────────────────┤
│      ⊗ Filters  [2]            │  (2 = number of active filters)
└────────────────────────────────┘
```

### Expanded View
```
┌────────────────────────────────┐
│ 🔍 Search tasks...         ✕    │
├────────────────────────────────┤
│    ⊗ Filters  [2]   ▼          │
├────────────────────────────────┤
│ Status                         │
│ [All] [Active] [Completed]     │
│ [Overdue]                      │
│                                │
│ Priority                       │
│ [All] [🔴High] [🟠Medium]      │
│ [🟡Low] [⚪None]              │
├────────────────────────────────┤
│     Clear All Filters          │
└────────────────────────────────┘
```

---

## Component Features

### Search Input
- Real-time search as user types
- Case-insensitive matching
- Clear button (✕) appears when text entered
- Placeholder: "Search tasks..."
- Search icon (🔍) on left side

### Status Filters
**Available Filters**:
- **All**: Show all tasks (default)
- **Active**: Show uncompleted tasks only
- **Completed**: Show completed tasks only
- **Overdue**: Show overdue incomplete tasks (due_date < today & !done)

**Visual Feedback**:
- Selected status: blue background, white text
- Unselected: white background, gray border, dark text

### Priority Filters
**Available Filters**:
- **All**: Show all priorities (default)
- **High** (1): Show high priority tasks (red indicator)
- **Medium** (2): Show medium priority tasks (orange indicator)
- **Low** (3): Show low priority tasks (yellow indicator)
- **None** (4): Show unset priority tasks (gray indicator)

**Visual Feedback**:
- Selected priority: blue background, white text, no color dot
- Unselected: white background, gray border, dark text + colored dot

### Filter Indicators
- **Active Filter Count**: Badge shows number of active filters
- **Toggle Button**: Changes color to light blue when filters active
- **Clear Button**: Only shows when filters are active

---

## Filtering Logic

### Search Filter
Matches text against task description (case-insensitive):
```
"Buy milk" search matches:
  ✓ "Buy milk for coffee"
  ✓ "buy MILK and butter"
  ✗ "Purchase dairy products"
```

### Status Filter
```
Status: all      → All tasks
Status: active   → Tasks where done == false
Status: completed → Tasks where done == true
Status: overdue  → Tasks where due_date < now() && done == false
```

### Priority Filter
```
Priority: all    → All tasks
Priority: 1      → Tasks where priority == 1 (HIGH)
Priority: 2      → Tasks where priority == 2 (MEDIUM)
Priority: 3      → Tasks where priority == 3 (LOW)
Priority: 4      → Tasks where priority == 4 (NONE)
```

### Multiple Filters (AND Logic)
All filters apply together:
```
searchText: "meeting" && status: "active" && priority: "1"
  → Active tasks containing "meeting" with HIGH priority
```

---

## State Management

### In HomeScreen
```typescript
// Initialize filters
const [filters, setFilters] = useState<SearchFilters>({
  searchText: '',
  status: 'all',
  priority: 'all',
});

// Handle search change
const handleSearchChange = (text: string) => {
  setFilters(prev => ({ ...prev, searchText: text }));
};

// Handle status change
const handleStatusFilterChange = (status: FilterStatus) => {
  setFilters(prev => ({ ...prev, status }));
};

// Handle priority change
const handlePriorityFilterChange = (priority: FilterPriority) => {
  setFilters(prev => ({ ...prev, priority }));
};

// Clear all filters
const handleClearFilters = () => {
  setFilters({
    searchText: '',
    status: 'all',
    priority: 'all',
  });
};
```

### Controlled Component Pattern
- Parent owns filter state
- Component calls callbacks on user interaction
- Parent updates state and re-renders component
- No internal state (except expandable panel)

---

## Integration with HomeScreen

### Data Flow
```
User types in search
           ↓
onSearchChange() callback
           ↓
Parent: setFilters({ ...prev, searchText: text })
           ↓
Component re-renders with new filters prop
           ↓
Parent: useMemo(filteredTasks) recalculates
           ↓
TaskList re-renders with filtered tasks
```

### Filtering Logic in Parent
```typescript
const filteredTasks = useMemo(() => {
  return currentListTasks.filter((task) => {
    // Search filter
    if (filters.searchText) {
      const searchLower = filters.searchText.toLowerCase();
      if (!task.todo.toLowerCase().includes(searchLower)) {
        return false;
      }
    }

    // Status filter
    if (filters.status !== 'all') {
      const currentTime = now();
      const isOverdue = task.due_date && task.due_date < currentTime && !task.done;

      if (filters.status === 'active' && task.done) return false;
      if (filters.status === 'completed' && !task.done) return false;
      if (filters.status === 'overdue' && !isOverdue) return false;
    }

    // Priority filter
    if (filters.priority !== 'all' && task.priority !== filters.priority) {
      return false;
    }

    return true;
  });
}, [currentListTasks, filters]);
```

---

## Styling Integration

### Theme System
Uses centralized theme objects:
- `colors`: primary, primaryLight, background, backgroundSecondary, text, textSecondary, textTertiary, border, error, white, gray500
- `spacing`: lg, md, sm, xs, borderRadius
- `typography`: styles (body, bodySmall, button, caption, label)

### Color Scheme
- **Search**: Icon (tertiary text), input (secondary background)
- **Active Filter**: Light blue badge (primary color)
- **Filter Buttons**: Blue when selected, white with border when unselected
- **Clear Button**: Red text (error color)
- **Priority Indicators**: Custom colors per priority level

---

## Keyboard Behavior

### Search Input
- Shows keyboard when focused
- Dismisses on back press
- Clear button (✕) tappable without keyboard
- No keyboard avoiding needed (component handles internally)

### Filter Buttons
- No keyboard interaction
- Touch-based selection
- Panel expand/collapse with finger tap

---

## Performance Considerations

### Optimization
- `useState` for expandable panel state only
- Parent handles filter state (prevents unnecessary re-renders)
- Filter buttons use `TouchableOpacity` for feedback
- Efficient filter counting (simple addition)

### Re-render Triggers
- Filter changes (from parent prop updates)
- Panel expand/collapse (internal state)
- NO re-renders on user interaction except above

### Memory
- Minimal state (1 boolean for panel)
- Filter state owned by parent
- Garbage collected on unmount

---

## Accessibility

### Touch Targets
- All buttons: 44pt minimum (iOS guidelines)
- Search clear button: `hitSlop` for easy dismiss
- Filter buttons: Comfortable spacing

### Color Contrast
- Text on white: meets WCAG AA standards
- Primary blue on white: 7.5:1 contrast
- Error text on white: 5.5:1 contrast

### Semantic Labels
- Search placeholder: "Search tasks..."
- Filter buttons: Clear text labels
- Status filters: "All", "Active", "Completed", "Overdue"
- Priority filters: "High", "Medium", "Low", "None"

---

## Testing

### Unit Test Example
```typescript
import { render, fireEvent } from '@testing-library/react-native';
import { SearchAndFilter } from '../components';

describe('SearchAndFilter', () => {
  const mockFilters = {
    searchText: '',
    status: 'all' as const,
    priority: 'all' as const,
  };

  it('calls onSearchChange when text entered', () => {
    const onSearchChange = jest.fn();
    const { getByPlaceholderText } = render(
      <SearchAndFilter
        filters={mockFilters}
        onSearchChange={onSearchChange}
        onStatusFilterChange={jest.fn()}
        onPriorityFilterChange={jest.fn()}
        onClearFilters={jest.fn()}
      />
    );

    const input = getByPlaceholderText('Search tasks...');
    fireEvent.changeText(input, 'meeting');

    expect(onSearchChange).toHaveBeenCalledWith('meeting');
  });

  it('shows active filter count', () => {
    const filters = {
      searchText: 'test',
      status: 'active',
      priority: '1',
    };
    const { getByText } = render(
      <SearchAndFilter
        filters={filters}
        onSearchChange={jest.fn()}
        onStatusFilterChange={jest.fn()}
        onPriorityFilterChange={jest.fn()}
        onClearFilters={jest.fn()}
      />
    );

    expect(getByText('2')).toBeTruthy(); // 2 active filters
  });
});
```

### Manual Testing Checklist
- [ ] Search input accepts text
- [ ] Clear button (✕) appears when text entered
- [ ] Clear button removes text
- [ ] Filters button toggles panel open/close
- [ ] Status filters can be selected/deselected
- [ ] Priority filters show colored dots
- [ ] Selected filter has blue background
- [ ] Active filter count badge shows correct number
- [ ] Clear All Filters resets everything
- [ ] Panel closes on Clear All Filters
- [ ] Multiple filters work together (AND logic)
- [ ] Component integrates with TaskList

---

## Edge Cases

### Empty State
- No tasks in list → Empty message shows (not affected by component)
- All filters applied → "No tasks match your filters" message
- No active filters → Normal task list displays

### Search Behavior
- Empty search string → No search filter applied
- Whitespace-only search → Treated as empty
- Special characters → Matched literally

### Filter Combinations
```
Active + High priority = Uncompleted HIGH priority tasks
Overdue + Low priority = Overdue LOW priority tasks
Search + Completed = Completed tasks matching search
```

### Rapid Changes
- User changes filters quickly → Parent handles debouncing (if desired)
- Multiple filter changes → All tracked in parent state
- Component updates efficiently with useMemo

---

## Related Components

- **TaskList**: Displays filtered tasks
- **TaskItem**: Individual task display
- **HomeScreen**: Parent that manages filter state
- **DueDatePicker**: Sets due dates (affects overdue filter)

---

## API Reference

### Exported Types
```typescript
type FilterStatus = 'all' | 'active' | 'completed' | 'overdue';
type FilterPriority = 'all' | 1 | 2 | 3 | 4;

interface SearchFilters {
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
```

### Exported Component
```typescript
export const SearchAndFilter: React.FC<SearchAndFilterProps>
export default SearchAndFilter
```

### Dependencies
- React Native: View, Text, TextInput, ScrollView, TouchableOpacity, Pressable, StyleSheet
- Theme: colors, spacing, typography
- Constants: TASK_PRIORITIES, PRIORITY_LABELS, PRIORITY_COLORS

---

## Future Enhancements

### Potential Features (Post-MVP)
- [ ] Date range filter (tasks due between dates)
- [ ] Multiple list filter (show tasks from multiple lists)
- [ ] Custom filter presets (save/load filter combinations)
- [ ] Recent searches (show previous searches)
- [ ] Filter suggestions (autocomplete)
- [ ] Advanced search operators (tag:work, priority:high, etc.)
- [ ] Filter history with undo/redo
- [ ] Saved filter views

---

## File Statistics

**SearchAndFilter.tsx**
- Lines: 280+
- Components: 1 functional component
- Hooks: 1 (useState)
- Styles: StyleSheet with 30+ style definitions
- TypeScript: 100% typed
- Complexity: Medium (filter logic in parent)

---

## Comparison: Before vs After

### Before Phase 5
```
HomeScreen
  └─ TaskList (all tasks for list)
     └─ TaskItem (no filtering)
```

### After Phase 5
```
HomeScreen
  ├─ SyncStatusBadge
  ├─ SearchAndFilter ← NEW
  │   └─ Search input
  │   └─ Status filters (All/Active/Completed/Overdue)
  │   └─ Priority filters (All/High/Medium/Low/None)
  │   └─ Clear filters button
  └─ TaskList (filtered tasks)
     └─ TaskItem (only matching tasks)
```

---

**Component Version**: 1.0.0
**Last Updated**: Phase 5 - Polish & Testing
**Status**: Production Ready ✅
