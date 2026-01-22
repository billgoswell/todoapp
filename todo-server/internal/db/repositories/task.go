package repositories

import (
	"database/sql"
	"time"
)

// TaskRepository handles all database operations for tasks
type TaskRepository struct {
	conn *sql.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(conn *sql.DB) *TaskRepository {
	return &TaskRepository{conn: conn}
}

// GetTasksSince retrieves all tasks modified since the given timestamp for a user
// Returns a slice of task data as maps and an error
func (r *TaskRepository) GetTasksSince(userID int, since int64) ([]map[string]interface{}, error) {
	query := `
		SELECT
			id, user_id, client_id, todo_list_id, todo, priority, done,
			date_added, date_completed, due_date, deleted, deleted_at,
			created_at, updated_at, version
		FROM tasks
		WHERE user_id = $1 AND updated_at >= to_timestamp($2)
		ORDER BY updated_at DESC
	`

	rows, err := r.conn.Query(query, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var (
			id, userID, todoListID, priority, version int
			clientID, todo                              string
			done, deleted                               bool
			dateAdded, createdAt, updatedAt             time.Time
			dateCompleted, dueDate, deletedAt           *time.Time
		)

		if err := rows.Scan(
			&id, &userID, &clientID, &todoListID, &todo, &priority, &done,
			&dateAdded, &dateCompleted, &dueDate, &deleted, &deletedAt,
			&createdAt, &updatedAt, &version,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, map[string]interface{}{
			"id":             id,
			"client_id":      clientID,
			"todo_list_id":   todoListID,
			"todo":           todo,
			"priority":       priority,
			"done":           done,
			"date_added":     dateAdded.Unix(),
			"date_completed": dateCompleted,
			"due_date":       dueDate,
			"deleted":        deleted,
			"deleted_at":     deletedAt,
			"created_at":     createdAt.Unix(),
			"updated_at":     updatedAt.Unix(),
			"version":        version,
		})
	}

	return tasks, rows.Err()
}

// UpsertTask inserts or updates a task using last-write-wins conflict resolution
// Returns the task ID and an error
func (r *TaskRepository) UpsertTask(userID int, task map[string]interface{}) (int, error) {
	query := `
		INSERT INTO tasks (
			user_id, client_id, todo_list_id, todo, priority, done,
			date_added, date_completed, due_date, deleted, deleted_at,
			updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (user_id, client_id) DO UPDATE SET
			todo = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.todo ELSE tasks.todo END,
			priority = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.priority ELSE tasks.priority END,
			done = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.done ELSE tasks.done END,
			date_completed = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.date_completed ELSE tasks.date_completed END,
			due_date = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.due_date ELSE tasks.due_date END,
			deleted = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.deleted ELSE tasks.deleted END,
			deleted_at = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.deleted_at ELSE tasks.deleted_at END,
			updated_at = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.updated_at ELSE tasks.updated_at END,
			version = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.version ELSE tasks.version END
		RETURNING id
	`

	// Helper function to convert timestamp values
	convertTimestamp := func(val interface{}) time.Time {
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

	// Convert all timestamps to time.Time
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

	var id int
	err := r.conn.QueryRow(
		query,
		userID,
		task["client_id"],
		task["todo_list_id"],
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
	).Scan(&id)
	return id, err
}
