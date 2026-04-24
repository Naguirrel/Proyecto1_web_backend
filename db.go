package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS series (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			episodes INTEGER NOT NULL CHECK (episodes > 0),
			image TEXT NOT NULL
		);
	`); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	return db, nil
}
