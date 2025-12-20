package database

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

func InitDatabase(dbPath string) (*sql.DB, error) {
	var db *sql.DB
	if dbPath == "" {
		dbPath = "./database/calendars.db"
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(
		context.Background(),
		`
		CREATE TABLE IF NOT EXISTS calendars (
			url TEXT PRIMARY KEY,
			last_synced TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			calendar_url TEXT NOT NULL REFERENCES calendars(url),
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			location TEXT
		);
		CREATE TABLE IF NOT EXISTS filters (
			calendar_url TEXT NOT NULL REFERENCES calendars(url),
			field TEXT NOT NULL,
			pattern TEXT NOT NULL,
			CHECK (field IN ('name', 'description', 'location')),
			PRIMARY KEY (calendar_url, field, pattern)
		);
		`,
	)
	if err != nil {
		return nil, err
	}
	return db, err
}
