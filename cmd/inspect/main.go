package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "inventory.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %s", dbPath)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to database!")
	fmt.Println("--------------------------------------------------")

	// List tables
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	fmt.Println("Tables found:")
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Printf("- %s\n", name)
		tables = append(tables, name)
	}
	fmt.Println("--------------------------------------------------")

	// Show users
	fmt.Println("Users Table Content:")
	userRows, err := db.Query("SELECT id, username, role, status FROM users")
	if err != nil {
		log.Printf("Error querying users: %v", err)
	} else {
		defer userRows.Close()
		for userRows.Next() {
			var id int
			var username, role, status string
			userRows.Scan(&id, &username, &role, &status)
			fmt.Printf("ID: %d | Username: %s | Role: %s | Status: %s\n", id, username, role, status)
		}
	}
}
