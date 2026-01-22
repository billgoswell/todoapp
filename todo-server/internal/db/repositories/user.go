package repositories

import (
	"database/sql"
)

// UserRepository handles all database operations for users
type UserRepository struct {
	conn *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(conn *sql.DB) *UserRepository {
	return &UserRepository{conn: conn}
}

// CreateUser creates a new user with an API key
// Returns the user ID
func (r *UserRepository) CreateUser(apiKey string) (int, error) {
	var id int
	err := r.conn.QueryRow(
		"INSERT INTO users (api_key) VALUES ($1) RETURNING id",
		apiKey,
	).Scan(&id)
	return id, err
}

// GetUserByAPIKey retrieves a user ID by their API key
// Returns the user ID and an error if not found
func (r *UserRepository) GetUserByAPIKey(apiKey string) (int, error) {
	var id int
	err := r.conn.QueryRow(
		"SELECT id FROM users WHERE api_key = $1",
		apiKey,
	).Scan(&id)
	return id, err
}
