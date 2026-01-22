package repositories

import (
	"fmt"
	"strings"
	"time"
)

// UpsertListsBatch inserts or updates multiple lists in a single operation
// Uses PostgreSQL's batch insert/update for better performance
func (r *ListRepository) UpsertListsBatch(userID int, lists []map[string]any) error {
	if len(lists) == 0 {
		return nil
	}

	// Build the VALUES clause with all lists
	var valueStrings []string
	var args []any
	argNum := 1

	for _, list := range lists {
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

		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			argNum, argNum+1, argNum+2, argNum+3, argNum+4, argNum+5, argNum+6,
		))

		args = append(args,
			userID,
			list["client_id"],
			list["name"],
			list["display_order"],
			list["archived"],
			time.Unix(updatedAtTs, 0),
			list["version"],
		)
		argNum += 7
	}

	query := fmt.Sprintf(`
		INSERT INTO todo_lists (
			user_id, client_id, name, display_order, archived,
			updated_at, version
		) VALUES %s
		ON CONFLICT (user_id, name) DO UPDATE SET
			display_order = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.display_order ELSE todo_lists.display_order END,
			archived = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.archived ELSE todo_lists.archived END,
			updated_at = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.updated_at ELSE todo_lists.updated_at END,
			version = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.version ELSE todo_lists.version END
	`, strings.Join(valueStrings, ", "))

	_, err := r.conn.Exec(query, args...)
	return err
}
