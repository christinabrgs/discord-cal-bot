package dbFunctions

import (
	"database/sql"
	"time"

	"git.phlcode.club/discord-bot/types"
)

func InsertURL(db *sql.DB, url string) (sql.Result, error) {
	result, err := db.Exec(
		`INSERT INTO calendars (url, last_synced) VALUES (?, ?);`,
		url,
		time.Now())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func InsertEvent(db *sql.DB, url string, e types.Event) (sql.Result, error) {
	result, err := db.Exec(
		`INSERT INTO events (calendar_url, name, description, start_time, end_time, location) VALUES (?, ?, ?, ?, ?, ?);`,
		url, e.Name, e.Description, e.StartTime, e.EndTime, e.Location)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func DeleteCalendarByURL(db *sql.DB, url string) (sql.Result, error) {
	result, err := db.Exec(
		`DELETE FROM calendars WHERE url = ?;`,
		url)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func DeleteEventsByURL(db *sql.DB, url string) (sql.Result, error) {
	result, err := db.Exec(
		`DELETE FROM events WHERE calendar_url = ?;`,
		url)
	if err != nil {
		return nil, err
	}
	return result, nil
}
