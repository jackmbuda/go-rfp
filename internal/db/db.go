package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./contracts.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS contracts (
		id TEXT PRIMARY KEY,
		title TEXT,
		description TEXT,
		status TEXT,
		naics TEXT,
		type TEXT,
		posted_date TEXT,
		response_deadline TEXT,
		award_date TEXT,
		contracting_office TEXT,
		agency TEXT,
		updated_at DATETIME
	);
	`
	if _, err := DB.Exec(schema); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}
}
