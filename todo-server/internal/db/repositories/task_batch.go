package repositories

import (
	"fmt"
	"strings"
	"time"
)

// UpsertTasksBatch inserts or updates multiple tasks in a single operation
// Uses PostgreSQL's UNNEST to batch the operations for better performance
// Returns ID mappings to update client's local database with server IDs
func (r *TaskRepository) UpsertTasksBatch(userID int, tasks []map[string]any) ([]IDMapping, error) {
	if len(tasks) == 0 {
		return []IDMapping{}, nil
	}

	// Build the VALUES clause with all tasks
	var valueStrings []string
	var args []any
	argNum := 1

	for _, task := range tasks {
		// Convert timestamps
		dateAdded := convertTimestamp(task["date_added"])
		dateCompleted := convertTimestamp(task["date_completed"])
		dueDate := convertTimestamp(task["due_date"])
		deletedAt := convertTimestamp(task["deleted_at"])
		updatedAt := convertTimestamp(task["updated_at"])

		// Convert pointer timestamps for optional fields
		var dateCompletedPtr, dueDatePtr, deletedAtPtr *time.Time
		if !dateCompleted.IsZero() {
			dateCompletedPtr = &dateCompleted
		}
		if !dueDate.IsZero() {
			dueDatePtr = &dueDate
		}
		if !deletedAt.IsZero() {
			deletedAtPtr = &deletedAt
		}

		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			argNum, argNum+1, argNum+2, argNum+3, argNum+4, argNum+5, argNum+6,
			argNum+7, argNum+8, argNum+9, argNum+10, argNum+11, argNum+12, argNum+13,
		))

		args = append(args,
			userID,
			task["client_id"],
			task["todo_list_id"],
			task["todo_list_client_id"],
			task["todo"],
			task["priority"],
			task["done"],
			dateAdded,
			dateCompletedPtr,
			dueDatePtr,
			task["deleted"],
			deletedAtPtr,
			updatedAt,
			task["version"],
		)
		argNum += 14
	}

	query := fmt.Sprintf(`
		INSERT INTO tasks (
			user_id, client_id, todo_list_id, todo_list_client_id, todo, priority, done,
			date_added, date_completed, due_date, deleted, deleted_at,
			updated_at, version
		) VALUES %s
		ON CONFLICT (user_id, client_id) DO UPDATE SET
			todo_list_client_id = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.todo_list_client_id ELSE tasks.todo_list_client_id END,
			todo = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.todo ELSE tasks.todo END,
			priority = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.priority ELSE tasks.priority END,
			done = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.done ELSE tasks.done END,
			date_completed = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.date_completed ELSE tasks.date_completed END,
			due_date = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.due_date ELSE tasks.due_date END,
			deleted = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.deleted ELSE tasks.deleted END,
			deleted_at = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.deleted_at ELSE tasks.deleted_at END,
			updated_at = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.updated_at ELSE tasks.updated_at END,
			version = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.version ELSE tasks.version END
		RETURNING id, client_id
	`, strings.Join(valueStrings, ", "))

	rows, err := r.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []IDMapping
	for rows.Next() {
		var id int
		var clientID string
		if err := rows.Scan(&id, &clientID); err != nil {
			return nil, err
		}
		mappings = append(mappings, IDMapping{
			ClientID: clientID,
			ServerID: id,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mappings, nil
}

// convertTimestamp is a helper function to convert various timestamp formats to time.Time
func convertTimestamp(val any) time.Time {
	switch v := val.(type) {
	case float64:
		if v == 0 {
			return time.Time{}
		}
		return time.Unix(int64(v), 0)
	case int64:
		if v == 0 {
			return time.Time{}
		}
		return time.Unix(v, 0)
	case nil:
		return time.Time{}
	default:
		return time.Time{}
	}
}
