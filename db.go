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

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ratings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			series_id INTEGER NOT NULL,
			reviewer TEXT NOT NULL DEFAULT 'Anonymous',
			score INTEGER NOT NULL CHECK (score >= 1 AND score <= 5),
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(series_id) REFERENCES series(id) ON DELETE CASCADE
		);
	`); err != nil {
		return nil, fmt.Errorf("create ratings table: %w", err)
	}

	if err := seedSeriesIfEmpty(db); err != nil {
		return nil, fmt.Errorf("seed database: %w", err)
	}

	return db, nil
}

func seedSeriesIfEmpty(db *sql.DB) error {
	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM series").Scan(&total); err != nil {
		return err
	}

	if total > 0 {
		return nil
	}

	seedData := []struct {
		title       string
		description string
		episodes    int
		image       string
	}{
		{
			title:       "Nebula High",
			description: "A sci-fi teen drama set inside a floating academy orbiting Saturn.",
			episodes:    24,
			image:       "https://images.unsplash.com/photo-1516849841032-87cbac4d88f7?auto=format&fit=crop&w=900&q=80",
		},
		{
			title:       "Midnight Detectives",
			description: "Two insomniac investigators solve strange city mysteries after dark.",
			episodes:    18,
			image:       "https://images.unsplash.com/photo-1504384308090-c894fdcc538d?auto=format&fit=crop&w=900&q=80",
		},
		{
			title:       "Pixel Raiders",
			description: "Competitive gamers get pulled into the retro arcade world they grew up loving.",
			episodes:    12,
			image:       "https://images.unsplash.com/photo-1511512578047-dfb367046420?auto=format&fit=crop&w=900&q=80",
		},
		{
			title:       "Cafe Aurora",
			description: "A cozy ensemble story about artists, bakers, and friendships in a mountain town.",
			episodes:    30,
			image:       "https://images.unsplash.com/photo-1509042239860-f550ce710b93?auto=format&fit=crop&w=900&q=80",
		},
		{
			title:       "Shadow Circuit",
			description: "An underground robotics league hides secrets that could change the whole country.",
			episodes:    20,
			image:       "https://images.unsplash.com/photo-1485827404703-89b55fcc595e?auto=format&fit=crop&w=900&q=80",
		},
		{
			title:       "Ocean Street 9",
			description: "Neighbors on a beachside block navigate love, family, and surprise second chances.",
			episodes:    16,
			image:       "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=900&q=80",
		},
	}

	for index, item := range seedData {
		if _, err := db.Exec(`
			INSERT INTO series (title, description, episodes, image)
			VALUES (?, ?, ?, ?)`,
			item.title,
			item.description,
			item.episodes,
			item.image,
		); err != nil {
			return err
		}

		if _, err := db.Exec(`
			INSERT INTO ratings (series_id, reviewer, score)
			VALUES (?, ?, ?), (?, ?, ?)`,
			index+1,
			"Pilot Viewer",
			4+(index%2),
			index+1,
			"Weekend Binger",
			3+(index%3),
		); err != nil {
			return err
		}
	}

	return nil
}
