package store

import (
	"database/sql"

	"git.phlcode.club/discord-bot/events"
)

type Store interface {
	InsertURL(url string) (sql.Result, error)
	InsertEvent(url string, e events.Event) (sql.Result, error)
	DeleteCalendarByURL(url string) (sql.Result, error)
	DeleteEventsByURL(url string) (sql.Result, error)
}
