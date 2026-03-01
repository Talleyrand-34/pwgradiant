package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// DB is the global database connection
var DB *sql.DB

// InitDB initializes the encrypted database connection
func InitDB(dbPath, encryptionKey string) error {
	dsn := fmt.Sprintf("file:%s?_pragma_key=%s&_pragma_cipher_page_size=4096", dbPath, encryptionKey)
	
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	DB = db
	return nil
}

// CreateSchema creates the database schema if it doesn't exist
func CreateSchema() error {
	sqlScript := `
PRAGMA foreign_keys = ON;

-- Main table: account
CREATE TABLE IF NOT EXISTS account (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  email     TEXT DEFAULT NULL,
  user      TEXT NOT NULL,
  password  TEXT NOT NULL,
  url       TEXT DEFAULT NULL,
  notes     TEXT DEFAULT NULL,
  expire    TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Main table: tag
CREATE TABLE IF NOT EXISTS tag (
  id    INTEGER PRIMARY KEY AUTOINCREMENT,
  name  TEXT UNIQUE NOT NULL
);

-- Relation table: tag_association
CREATE TABLE IF NOT EXISTS tag_association (
  account_id INTEGER NOT NULL,
  tag_id     INTEGER NOT NULL,
  PRIMARY KEY (account_id, tag_id),
  FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
  FOREIGN KEY (tag_id)     REFERENCES tag(id) ON DELETE CASCADE
);

-- Main table: totp (linked to account)
CREATE TABLE IF NOT EXISTS totp (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  account_id      INTEGER NOT NULL UNIQUE,
  totp_seed       TEXT NOT NULL,
  c_totp_seed     TEXT DEFAULT NULL,
  paillier_n      TEXT DEFAULT NULL,
  use_homomorphic INTEGER DEFAULT 0,
  FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE
);

----------------------------------------------------
-- History / versioning tables
----------------------------------------------------

CREATE TABLE IF NOT EXISTS account_history (
  history_id    INTEGER PRIMARY KEY AUTOINCREMENT,
  account_id    INTEGER DEFAULT NULL,
  email         TEXT DEFAULT NULL,
  user          TEXT DEFAULT NULL,
  password      TEXT DEFAULT NULL,
  url           TEXT DEFAULT NULL,
  notes         TEXT DEFAULT NULL,
  expire        TEXT DEFAULT NULL,
  valid_from    TEXT NOT NULL DEFAULT (datetime('now')),
  valid_to      TEXT NOT NULL DEFAULT (datetime('now')),
  change_reason TEXT DEFAULT NULL,
  FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS tag_association_history (
  account_id INTEGER NOT NULL,
  tag_id     INTEGER NOT NULL,
  valid_from TEXT NOT NULL DEFAULT (datetime('now')),
  valid_to   TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (account_id, tag_id, valid_from),
  FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
  FOREIGN KEY (tag_id)     REFERENCES tag(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS totp_history (
  history_id      INTEGER PRIMARY KEY AUTOINCREMENT,
  totp_id         INTEGER NOT NULL,
  totp_seed       TEXT NOT NULL,
  c_totp_seed     TEXT DEFAULT NULL,
  paillier_n      TEXT DEFAULT NULL,
  use_homomorphic INTEGER DEFAULT 0,
  valid_from      TEXT NOT NULL DEFAULT (datetime('now')),
  valid_to        TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (totp_id) REFERENCES totp(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_account_email ON account(email);
CREATE INDEX IF NOT EXISTS idx_account_user ON account(user);
CREATE INDEX IF NOT EXISTS idx_account_expire ON account(expire);
CREATE INDEX IF NOT EXISTS idx_tag_association_account ON tag_association(account_id);
CREATE INDEX IF NOT EXISTS idx_tag_association_tag ON tag_association(tag_id);
CREATE INDEX IF NOT EXISTS idx_totp_account ON totp(account_id);
CREATE INDEX IF NOT EXISTS idx_account_history_account ON account_history(account_id);
CREATE INDEX IF NOT EXISTS idx_account_history_valid ON account_history(valid_from, valid_to);
CREATE INDEX IF NOT EXISTS idx_tag_association_history_dates ON tag_association_history(valid_from, valid_to);
CREATE INDEX IF NOT EXISTS idx_totp_history_totp ON totp_history(totp_id);
`

	_, err := DB.Exec(sqlScript)
	if err != nil {
		return fmt.Errorf("error creating schema: %w", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
