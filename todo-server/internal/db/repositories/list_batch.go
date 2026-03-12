package repositories

import (
	"fmt"
	"strings"
	"time"
)

// IDMapping represents a client-to-server ID mapping
type IDMapping struct {
	ClientID string
	ServerID int
}

// UpsertListsBatch inserts or updates multiple lists in a single operation
// Uses PostgreSQL's batch insert/update for better performance
// Returns ID mappings to update client's local database with server IDs
func (r *ListRepository) UpsertListsBatch(userID int, lists []map[string]any) ([]IDMapping, error) {
	if len(lists) == 0 {
		return []IDMapping{}, nil
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
		ON CONFLICT (user_id, client_id) DO UPDATE SET
			display_order = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.display_order ELSE todo_lists.display_order END,
			archived = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.archived ELSE todo_lists.archived END,
			updated_at = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.updated_at ELSE todo_lists.updated_at END,
			version = CASE WHEN EXCLUDED.updated_at > todo_lists.updated_at THEN EXCLUDED.version ELSE todo_lists.version END
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
