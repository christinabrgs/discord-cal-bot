package store

import (
	"database/sql"
	"fmt"
	"regexp"
	"time"

	e "git.phlcode.club/discord-bot/events"
)

type SQLiteStore struct {
	*sql.DB
}

func (s SQLiteStore) DeleteEventsByIDs(ids []string) error {
	// WARN: This is really not the right way to do this...
	// We should probably revisit it eventually as this is
	// a potential security risk via SQL injection if we ever use this method without
	var stmt string
	for i, id := range ids {
		if i < len(ids)-1 {
			stmt += id + ", "
		} else {
			stmt += id
		}
	}
	_, err := s.Exec("DELETE FROM events WHERE id in (" + stmt + ");")
	return err
}

func (s SQLiteStore) GetEventsByPattern(filter Filter) ([]string, error) {
	query := "SELECT id, "
	switch filter.Field {
	case FilterFieldName:
		query += "name"
	case FilterFieldDescription:
		query += "description"
	case FilterFieldLocation:
		query += "location"
	default:
		return nil, fmt.Errorf("invalid filter field: %s", filter.Field)
	}
	query += " FROM events WHERE calendar_url = ?;"
	rows, err := s.Query(query, filter.URL)
	if err != nil {
		return nil, fmt.Errorf("unable to get events by pattern: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("error preparing db data for scan: %w", err)
		}

		var id string
		var against string
		err = rows.Scan(&id, &against)
		if rows.Err() != nil {
			return nil, fmt.Errorf("unable to scan data into id and pattern match: %w", err)
		}
		fmt.Println(against, !filter.Pattern.MatchString(against))
		if !filter.Pattern.MatchString(against) {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (s SQLiteStore) GetEventsForURL(url string) ([]e.Event, error) {
	rows, err := s.Query("SELECT id, name, description, start_time, end_time, location FROM events WHERE calendar_url = ?", url)
	if err != nil {
		return nil, fmt.Errorf("unable to get events from db: %w", err)
	}
	defer rows.Close()

	events := make([]e.Event, 0)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("error preparing db data for scan: %w", err)
		}

		var event e.Event
		err = rows.Scan(&event.ID, &event.Name, &event.Description, &event.StartTime, &event.EndTime, &event.Location)
		if rows.Err() != nil {
			return nil, fmt.Errorf("unable to scan data into Event struct: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (s SQLiteStore) CreateFilter(url string, field FilterField, pattern regexp.Regexp) (Filter, error) {
	_, err := s.Exec(
		`INSERT INTO filters (calendar_url, field, pattern) VALUES (?, ?, ?);`,
		url,
		string(field),
		pattern.String(),
	)
	if err != nil {
		return Filter{}, err
	}

	return Filter{URL: url, Field: field, Pattern: pattern}, nil
}

func (s SQLiteStore) DeleteFilter(url string, field FilterField, pattern regexp.Regexp) error {
	panic("unimplemented")
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

func (s SQLiteStore) InsertEvent(url string, e e.Event) (sql.Result, error) {
	result, err := s.Exec(
		`INSERT INTO events (calendar_url, id, name, description, start_time, end_time, location) VALUES (?, ?, ?, ?, ?, ?, ?);`,
		url, e.ID, e.Name, e.Description, e.StartTime, e.EndTime, e.Location)
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

func (s SQLiteStore) DeleteEventsByURL(url string) ([]string, error) {
	rows, err := s.Query(
		`DELETE FROM events WHERE calendar_url = ? RETURNING id;`,
		url)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("error preparing db data for scan: %w", err)
		}

		var id string
		err = rows.Scan(&id)
		if rows.Err() != nil {
			return nil, fmt.Errorf("unable to scan data into Event struct: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func NewSQLiteStore(db *sql.DB) Store {
	return SQLiteStore{db}
}
