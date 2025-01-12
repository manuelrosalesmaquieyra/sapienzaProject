package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB crea una base de datos en memoria para testing
func setupTestDB(t *testing.T) AppDatabase {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("error opening database: %v", err)
	}

	appDB, err := New(db)
	if err != nil {
		t.Fatalf("error creating app database: %v", err)
	}

	return appDB
}

// TestPing verifica la conexión a la base de datos
func TestPing(t *testing.T) {
	db := setupTestDB(t)
	if err := db.Ping(); err != nil {
		t.Errorf("ping failed: %v", err)
	}
}
