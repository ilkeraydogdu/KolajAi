package database

import (
	"os"
	"testing"
)

func TestNewSQLiteConnection(t *testing.T) {
	// Test with a temporary database file
	testDB := "test_kolajAi.db"
	defer os.Remove(testDB) // Clean up after test

	db, err := NewSQLiteConnection(testDB)
	if err != nil {
		t.Fatalf("Failed to create SQLite connection: %v", err)
	}
	defer db.Close()

	// Test that we can ping the database
	if err := db.Ping(); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestNewSQLiteConnection_InvalidPath(t *testing.T) {
	// Test with an invalid path (directory that doesn't exist)
	invalidPath := "/nonexistent/directory/test.db"

	db, err := NewSQLiteConnection(invalidPath)
	if err == nil {
		db.Close()
		t.Error("Expected error for invalid database path, but got none")
	}
}

func TestDatabaseExists(t *testing.T) {
	// Test with existing file
	testDB := "test_exists.db"

	// Create a test file
	file, err := os.Create(testDB)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()
	defer os.Remove(testDB)

	if !DatabaseExists(testDB) {
		t.Error("Expected database to exist, but it doesn't")
	}

	// Test with non-existing file
	if DatabaseExists("nonexistent.db") {
		t.Error("Expected database to not exist, but it does")
	}
}
