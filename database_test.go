package main

import (
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_init.db"
	encryptionKey := "test-key-123"

	// Clean up before test
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	// Test initialization
	err := InitDB(dbPath, encryptionKey)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Verify DB is not nil
	if DB == nil {
		t.Fatal("DB should not be nil after InitDB")
	}

	// Test connection
	err = DB.Ping()
	if err != nil {
		t.Fatalf("DB.Ping() failed: %v", err)
	}

	// Clean up
	CloseDB()
}

// func TestInitDBInvalidKey(t *testing.T) {
// 	dbPath := "test_invalid.db"
//
// 	// Create DB with one key
// 	os.Remove(dbPath)
// 	err := InitDB(dbPath, "key1")
// 	if err != nil {
// 		t.Fatalf("InitDB failed: %v", err)
// 	}
// 	CloseDB()
//
// 	// Try to open with different key - should fail
// 	err = InitDB(dbPath, "key2")
// 	if err == nil {
// 		t.Fatal("Expected error when opening with wrong key")
// 	}
//
// 	os.Remove(dbPath)
// }

func TestCreateSchema(t *testing.T) {
	dbPath := "test_schema.db"
	encryptionKey := "test-key-456"

	os.Remove(dbPath)
	defer os.Remove(dbPath)

	// Initialize DB
	err := InitDB(dbPath, encryptionKey)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	// Create schema
	err = CreateSchema()
	if err != nil {
		t.Fatalf("CreateSchema failed: %v", err)
	}

	// Verify tables exist
	tables := []string{
		"account",
		"tag",
		"tag_association",
		"totp",
		"account_history",
		"tag_association_history",
		"totp_history",
	}

	for _, table := range tables {
		var count int
		err := DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).
			Scan(&count)
		if err != nil {
			t.Fatalf("Error checking table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Table %s should exist", table)
		}
	}
}

func TestCreateSchemaIdempotent(t *testing.T) {
	dbPath := "test_idempotent.db"
	encryptionKey := "test-key-789"

	os.Remove(dbPath)
	defer os.Remove(dbPath)

	err := InitDB(dbPath, encryptionKey)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	// Create schema twice - should not error
	err = CreateSchema()
	if err != nil {
		t.Fatalf("First CreateSchema failed: %v", err)
	}

	err = CreateSchema()
	if err != nil {
		t.Fatalf("Second CreateSchema failed (should be idempotent): %v", err)
	}
}

func TestCloseDB(t *testing.T) {
	dbPath := "test_close.db"
	encryptionKey := "test-key-close"

	os.Remove(dbPath)
	defer os.Remove(dbPath)

	err := InitDB(dbPath, encryptionKey)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Close DB
	err = CloseDB()
	if err != nil {
		t.Fatalf("CloseDB failed: %v", err)
	}

	// Close again should not error
	err = CloseDB()
	if err != nil {
		t.Fatalf("CloseDB on nil DB should not error: %v", err)
	}
}

func TestForeignKeysEnabled(t *testing.T) {
	dbPath := "test_fk.db"
	encryptionKey := "test-key-fk"

	os.Remove(dbPath)
	defer os.Remove(dbPath)

	err := InitDB(dbPath, encryptionKey)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	err = CreateSchema()
	if err != nil {
		t.Fatalf("CreateSchema failed: %v", err)
	}

	// Check if foreign keys are enabled
	var fkEnabled int
	err = DB.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatalf("Error checking foreign keys: %v", err)
	}

	if fkEnabled != 1 {
		t.Error("Foreign keys should be enabled")
	}
}
