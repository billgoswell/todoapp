package repositories

import (
	"database/sql"
	"fmt"
	"time"
)

// ListRepository handles all database operations for todo lists
type ListRepository struct {
	conn *sql.DB
}

// NewListRepository creates a new list repository
func NewListRepository(conn *sql.DB) *ListRepository {
	return &ListRepository{conn: conn}
}

// GetListsSince retrieves all lists modified since the given timestamp for a user
// Returns a slice of list data as maps and an error
func (r *ListRepository) GetListsSince(userID int, since int64) ([]map[string]interface{}, error) {
	query := `
		SELECT
			id, user_id, client_id, name, display_order, archived,
			created_at, updated_at, version
		FROM todo_lists
		WHERE user_id = $1 AND updated_at >= to_timestamp($2)
		ORDER BY updated_at DESC
	`

	rows, err := r.conn.Query(query, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []map[string]interface{}
	for rows.Next() {
		var (
			id, userID, displayOrder, version int
			clientID, name                    string
			archived                          bool
			createdAt, updatedAt              time.Time
		)

		if err := rows.Scan(
			&id, &userID, &clientID, &name, &displayOrder, &archived,
			&createdAt, &updatedAt, &version,
		); err != nil {
			return nil, err
		}

		lists = append(lists, map[string]interface{}{
			"id":              id,
			"client_id":       clientID,
			"name":            name,
			"display_order":   displayOrder,
			"archived":        archived,
			"created_at":      createdAt.Unix(),
			"updated_at":      updatedAt.Unix(),
			"version":         version,
		})
	}

	return lists, rows.Err()
}

// GetIDByClientID returns the server ID for a list given its clientID
// Used to resolve stable UUID references to database IDs
func (r *ListRepository) GetIDByClientID(userID int, clientID string) (int, error) {
	query := `
		SELECT id FROM todo_lists
		WHERE user_id = $1 AND client_id = $2
	`

	var id int
	err := r.conn.QueryRow(query, userID, clientID).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("list not found with client_id: %s", clientID)
	}
	return id, err
}

// UpsertList inserts or updates a list using last-write-wins conflict resolution
// Returns the list ID and an error
func (r *ListRepository) UpsertList(userID int, list map[string]interface{}) (int, error) {
	query := `
		INSERT INTO todo_lists (
			user_id, client_id, name, display_order, archived,
			updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, name) DO UPDATE SET
			display_order = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.display_order ELSE todo_lists.display_order END,
			archived = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.archived ELSE todo_lists.archived END,
			updated_at = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.updated_at ELSE todo_lists.updated_at END,
			version = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.version ELSE todo_lists.version END
		RETURNING id
	`

	// Convert updated_at timestamp (handle both float64 and int64)
	var updatedAtTs int64
	switch v := list["updated_at"].(type) {
	case float64:
		updatedAtTs = int64(v)
	case int64:
		updatedAtTs = v
	default:
		updatedAtTs = time.Now().Unix()
	}

	var id int
	err := r.conn.QueryRow(
		query,
		userID,
		list["client_id"],
		list["name"],
		list["display_order"],
		list["archived"],
		time.Unix(updatedAtTs, 0),
		list["version"],
	).Scan(&id)
	return id, err
}
