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

	if err := migrateLegacySeedData(db); err != nil {
		return nil, fmt.Errorf("migrate legacy seed data: %w", err)
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

	seedData := realSeriesSeedData()

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

func migrateLegacySeedData(db *sql.DB) error {
	replacements := []struct {
		oldTitle string
		newData  struct {
			title       string
			description string
			episodes    int
			image       string
		}
	}{
		{oldTitle: "Nebula High", newData: realSeriesSeedData()[0]},
		{oldTitle: "Midnight Detectives", newData: realSeriesSeedData()[1]},
		{oldTitle: "Pixel Raiders", newData: realSeriesSeedData()[2]},
		{oldTitle: "Cafe Aurora", newData: realSeriesSeedData()[3]},
		{oldTitle: "Shadow Circuit", newData: realSeriesSeedData()[4]},
		{oldTitle: "Ocean Street 9", newData: realSeriesSeedData()[5]},
	}

	for _, item := range replacements {
		if _, err := db.Exec(`
			UPDATE series
			SET title = ?, description = ?, episodes = ?, image = ?
			WHERE title = ?`,
			item.newData.title,
			item.newData.description,
			item.newData.episodes,
			item.newData.image,
			item.oldTitle,
		); err != nil {
			return err
		}
	}

	return nil
}

func realSeriesSeedData() []struct {
	title       string
	description string
	episodes    int
	image       string
} {
	return []struct {
		title       string
		description string
		episodes    int
		image       string
	}{
		{
			title:       "Breaking Bad",
			description: "A chemistry teacher turned meth producer descends into the criminal underworld.",
			episodes:    62,
			image:       "https://placehold.co/600x900?text=Breaking+Bad",
		},
		{
			title:       "Stranger Things",
			description: "A group of kids in Hawkins faces supernatural forces, secret experiments, and the Upside Down.",
			episodes:    34,
			image:       "https://placehold.co/600x900?text=Stranger+Things",
		},
		{
			title:       "Game of Thrones",
			description: "Noble families battle for power in a brutal fantasy world where dragons and danger return.",
			episodes:    73,
			image:       "https://placehold.co/600x900?text=Game+of+Thrones",
		},
		{
			title:       "The Office",
			description: "A mockumentary workplace comedy following the employees of Dunder Mifflin.",
			episodes:    201,
			image:       "https://placehold.co/600x900?text=The+Office",
		},
		{
			title:       "Dark",
			description: "Families in a German town uncover a time travel mystery that spans generations.",
			episodes:    26,
			image:       "https://placehold.co/600x900?text=Dark",
		},
		{
			title:       "The Crown",
			description: "A historical drama chronicling the reign of Queen Elizabeth II across decades.",
			episodes:    60,
			image:       "https://placehold.co/600x900?text=The+Crown",
		},
	}
}
