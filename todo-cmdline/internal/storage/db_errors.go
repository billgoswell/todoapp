package storage

import (
	"database/sql"

	"github.com/billgoswell/commandlinetodo/internal/errors"
	"github.com/billgoswell/commandlinetodo/internal/logger"
)

// ExecuteWithError wraps database execution with structured error handling
func ExecuteWithError(db *sql.DB, operation string, query string, args ...any) error {
	if _, err := db.Exec(query, args...); err != nil {
		appErr := errors.DatabaseError(operation, err)
		appErr.Log()
		return appErr
	}
	return nil
}

// ExecuteWithIDError wraps database execution with ID retrieval and error handling
func ExecuteWithIDError(db *sql.DB, operation string, query string, args ...any) (int, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		appErr := errors.DatabaseError(operation, err)
		appErr.Log()
		return 0, appErr
	}

	id, err := result.LastInsertId()
	if err != nil {
		appErr := errors.DatabaseError(operation, err)
		appErr.Log()
		return 0, appErr
	}

	return int(id), nil
}

// QueryWithError wraps database queries with error handling
func QueryWithError(db *sql.DB, operation string, query string, args ...any) (*sql.Rows, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		appErr := errors.QueryError(operation, err)
		appErr.Log()
		return nil, appErr
	}
	return rows, nil
}

// QueryRowWithError wraps single row queries with error handling
func QueryRowWithError(db *sql.DB, operation string, query string, args ...any) *sql.Row {
	return db.QueryRow(query, args...)
}

// BeginWithError wraps transaction begin with error handling
func BeginWithError(db *sql.DB, operation string) (*sql.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		appErr := errors.TransactionError(operation, err)
		appErr.Log()
		return nil, appErr
	}
	return tx, nil
}

// TxExecuteWithError wraps transaction execution with error handling
func TxExecuteWithError(tx *sql.Tx, operation string, query string, args ...any) error {
	_, err := tx.Exec(query, args...)
	if err != nil {
		appErr := errors.TransactionError(operation, err)
		appErr.Log()
		return appErr
	}
	return nil
}

// CommitWithError wraps transaction commit with error handling
func CommitWithError(tx *sql.Tx, operation string) error {
	if err := tx.Commit(); err != nil {
		appErr := errors.TransactionError(operation, err)
		appErr.Log()
		return appErr
	}
	return nil
}

// RollbackWithError wraps transaction rollback with error handling
func RollbackWithError(tx *sql.Tx, operation string) {
	if err := tx.Rollback(); err != nil {
		logger.LogError(operation, "error", err)
	}
}

// ScanRowsWithError helps scan multiple rows with error handling
func ScanRowsWithError(rows *sql.Rows, operation string) (*sql.Rows, error) {
	if err := rows.Err(); err != nil {
		rows.Close()
		appErr := errors.QueryError(operation, err)
		appErr.Log()
		return nil, appErr
	}
	return rows, nil
}
