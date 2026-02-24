package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func NewDB() (*sql.DB, error) {
	dbName := "inventory.db"

	// Use DSN parameters to ensure pragmas apply to EVERY connection in the pool
	// _pragma=busy_timeout=5000: Wait up to 5s for locks
	// _pragma=journal_mode=WAL: Better concurrency
	// _pragma=foreign_keys=ON: Enforce FK constraints
	// _pragma=synchronous=NORMAL: Faster writes, safe enough for WAL
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=foreign_keys=ON&_pragma=synchronous=NORMAL", dbName)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// SQLite only supports one writer at a time.
	// To avoid "database is locked" errors during heavy writes, we limit to 1 open connection.
	// With WAL mode, this still allows reasonable performance.
	db.SetMaxOpenConns(1)

	log.Printf("Connected to SQLite database: %s (WAL mode enabled)", dbName)
	return db, nil
}

func InitDB(db *sql.DB) error {
	// Check if initialization is needed
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&name)
	if err == nil {
		// Table exists, skip initial schema execution
		return nil
	}

	log.Println("Initializing database schema...")
	schema, err := os.ReadFile("database/schema/sqlite_schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("Database schema initialized successfully")
	return nil
}
