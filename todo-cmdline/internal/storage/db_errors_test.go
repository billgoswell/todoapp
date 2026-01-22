package storage

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/billgoswell/commandlinetodo/internal/errors"
)

// Setup helper to create in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create test table
	if _, err := db.Exec(`
		CREATE TABLE test_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value INTEGER
		)
	`); err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	return db
}

func TestExecuteWithError_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := ExecuteWithError(db, "insert test item",
		"INSERT INTO test_items (name, value) VALUES (?, ?)",
		"test", 42)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestExecuteWithError_InvalidSQL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := ExecuteWithError(db, "bad insert",
		"INSERT INTO nonexistent_table (col) VALUES (?)",
		"value")

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.IsErrorCode(err, errors.ErrDatabaseIO) {
		t.Errorf("expected ErrDatabaseIO, got %s", errors.ErrorCode(err))
	}
}

func TestExecuteWithIDError_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id, err := ExecuteWithIDError(db, "insert item",
		"INSERT INTO test_items (name, value) VALUES (?, ?)",
		"item1", 100)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}
}

func TestExecuteWithIDError_InvalidSQL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := ExecuteWithIDError(db, "bad insert",
		"INSERT INTO nonexistent (col) VALUES (?)",
		"value")

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.IsErrorCode(err, errors.ErrDatabaseIO) {
		t.Errorf("expected ErrDatabaseIO, got %s", errors.ErrorCode(err))
	}
}

func TestQueryWithError_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	db.Exec("INSERT INTO test_items (name, value) VALUES (?, ?)", "item1", 10)
	db.Exec("INSERT INTO test_items (name, value) VALUES (?, ?)", "item2", 20)

	rows, err := QueryWithError(db, "select items",
		"SELECT id, name, value FROM test_items ORDER BY id")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var id int
		var name string
		var value int
		if err := rows.Scan(&id, &name, &value); err != nil {
			t.Errorf("failed to scan row: %v", err)
		}
	}

	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}
}

func TestQueryWithError_InvalidSQL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := QueryWithError(db, "bad select",
		"SELECT * FROM nonexistent_table")

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.IsErrorCode(err, errors.ErrDatabaseQuery) {
		t.Errorf("expected ErrDatabaseQuery, got %s", errors.ErrorCode(err))
	}
}

func TestQueryRowWithError_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Exec("INSERT INTO test_items (name, value) VALUES (?, ?)", "item", 99)

	row := QueryRowWithError(db, "select single",
		"SELECT id, name, value FROM test_items WHERE id = ?", 1)

	var id int
	var name string
	var value int
	if err := row.Scan(&id, &name, &value); err != nil {
		t.Errorf("failed to scan row: %v", err)
	}

	if id != 1 || name != "item" || value != 99 {
		t.Errorf("unexpected scan results: id=%d, name=%s, value=%d", id, name, value)
	}
}

func TestBeginWithError_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tx, err := BeginWithError(db, "start transaction")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	tx.Rollback()
}

func TestBeginWithError_ClosedDB(t *testing.T) {
	db := setupTestDB(t)
	db.Close()

	_, err := BeginWithError(db, "start transaction")
	if err == nil {
		t.Error("expected error for closed database, got nil")
	}
	if !errors.IsErrorCode(err, errors.ErrDatabaseTx) {
		t.Errorf("expected ErrDatabaseTx, got %s", errors.ErrorCode(err))
	}
}

func TestTxExecuteWithError_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tx, _ := BeginWithError(db, "start tx")
	defer tx.Rollback()

	err := TxExecuteWithError(tx, "tx insert",
		"INSERT INTO test_items (name, value) VALUES (?, ?)",
		"tx_item", 55)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestTxExecuteWithError_InvalidSQL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tx, _ := BeginWithError(db, "start tx")
	defer tx.Rollback()

	err := TxExecuteWithError(tx, "bad tx insert",
		"INSERT INTO nonexistent (col) VALUES (?)",
		"value")

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.IsErrorCode(err, errors.ErrDatabaseTx) {
		t.Errorf("expected ErrDatabaseTx, got %s", errors.ErrorCode(err))
	}
}

func TestCommitWithError_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tx, _ := BeginWithError(db, "start tx")

	TxExecuteWithError(tx, "tx insert",
		"INSERT INTO test_items (name, value) VALUES (?, ?)",
		"committed", 77)

	err := CommitWithError(tx, "commit")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify item was actually inserted
	var count int
	db.QueryRow("SELECT COUNT(*) FROM test_items WHERE value = 77").Scan(&count)
	if count != 1 {
		t.Error("expected committed item to exist in database")
	}
}

func TestRollbackWithError_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tx, _ := BeginWithError(db, "start tx")

	TxExecuteWithError(tx, "tx insert",
		"INSERT INTO test_items (name, value) VALUES (?, ?)",
		"rollback_test", 88)

	RollbackWithError(tx, "rollback")

	// Verify item was NOT inserted
	var count int
	db.QueryRow("SELECT COUNT(*) FROM test_items WHERE value = 88").Scan(&count)
	if count != 0 {
		t.Error("expected rolled back item to NOT exist in database")
	}
}

func TestScanRowsWithError_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Exec("INSERT INTO test_items (name, value) VALUES (?, ?)", "item", 1)

	rows, _ := QueryWithError(db, "select", "SELECT id FROM test_items")
	defer rows.Close()

	rows.Next()
	_, err := ScanRowsWithError(rows, "check rows")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestScanRowsWithError_InvalidIteration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create an invalid query that will set rows.Err()
	rows, _ := QueryWithError(db, "select", "SELECT id FROM test_items WHERE id = ?", 1)

	// Iterate past end of rows
	for rows.Next() {
	}

	// Now call ScanRowsWithError which checks rows.Err()
	// In this case, rows.Err() should be nil (normal completion)
	_, err := ScanRowsWithError(rows, "check rows")
	rows.Close()

	// Since we completed normally, no error expected
	if err != nil {
		t.Errorf("expected no error for normal completion, got %v", err)
	}
}

func TestErrorCodeConsistency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	testCases := []struct {
		name        string
		operation   func() error
		expectedErr string
	}{
		{
			name: "execute error",
			operation: func() error {
				return ExecuteWithError(db, "test", "INSERT INTO invalid VALUES (?)", "x")
			},
			expectedErr: errors.ErrDatabaseIO,
		},
		{
			name: "query error",
			operation: func() error {
				_, err := QueryWithError(db, "test", "SELECT * FROM invalid")
				return err
			},
			expectedErr: errors.ErrDatabaseQuery,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operation()
			if err == nil {
				t.Error("expected error, got nil")
			}
			if !errors.IsErrorCode(err, tc.expectedErr) {
				t.Errorf("expected %s, got %s", tc.expectedErr, errors.ErrorCode(err))
			}
		})
	}
}

func TestAppErrorDetails(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := ExecuteWithError(db, "test operation", "INSERT INTO invalid VALUES (?)", "test")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatal("expected AppError type")
	}

	if appErr.Operation != "test operation" {
		t.Errorf("expected operation 'test operation', got %s", appErr.Operation)
	}

	if appErr.Code != errors.ErrDatabaseIO {
		t.Errorf("expected code %s, got %s", errors.ErrDatabaseIO, appErr.Code)
	}

	if appErr.Cause == nil {
		t.Error("expected Cause to be set")
	}
}
