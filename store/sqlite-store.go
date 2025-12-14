package store

import (
	"database/sql"
	"time"

	"git.phlcode.club/discord-bot/events"
)

type SQLiteStore struct {
	*sql.DB
}

func (s SQLiteStore) InsertURL(url string) (sql.Result, error) {
	result, err := s.Exec(
		`INSERT INTO calendars (url, last_synced) VALUES (?, ?);`,
		url,
		time.Now())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s SQLiteStore) InsertEvent(url string, e events.Event) (sql.Result, error) {
	result, err := s.Exec(
		`INSERT INTO events (calendar_url, name, description, start_time, end_time, location) VALUES (?, ?, ?, ?, ?, ?);`,
		url, e.Name, e.Description, e.StartTime, e.EndTime, e.Location)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s SQLiteStore) DeleteCalendarByURL(url string) (sql.Result, error) {
	result, err := s.Exec(
		`DELETE FROM calendars WHERE url = ?;`,
		url)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s SQLiteStore) DeleteEventsByURL(url string) (sql.Result, error) {
	result, err := s.Exec(
		`DELETE FROM events WHERE calendar_url = ?;`,
		url)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewSQLiteStore(db *sql.DB) Store {
	return SQLiteStore{db}
}
