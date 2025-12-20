package store

import (
	"database/sql"
	"fmt"
	"regexp"

	"git.phlcode.club/discord-bot/events"
)

type FilterField = string

const (
	FilterFieldName        = "name"
	FilterFieldDescription = "description"
	FilterFieldLocation    = "location"
)

func NewFilter(url, field, pattern string) (*Filter, error) {
	var filter Filter
	filter.URL = url
	switch field {
	case FilterFieldName:
	case FilterFieldDescription:
	case FilterFieldLocation:
	default:
		return nil, fmt.Errorf("unexpected filter field value: %s", field)
	}
	filter.Field = FilterField(field)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid filter pattern: %s", pattern)
	}
	filter.Pattern = *regex
	return &filter, nil
}

type Filter struct {
	URL     string
	Field   FilterField
	Pattern regexp.Regexp
}

func (f Filter) Filter(event events.Event) bool {
	var against string
	switch f.Field {
	case FilterFieldName:
		against = event.Name
	case FilterFieldDescription:
		against = event.Description
	case FilterFieldLocation:
		against = event.Location
	}
	return f.Pattern.MatchString(against)
}

type Store interface {
	InsertURL(url string) (sql.Result, error)
	InsertEvent(url string, e events.Event) (sql.Result, error)
	DeleteCalendarByURL(url string) (sql.Result, error)
	DeleteEventsByURL(url string) ([]string, error)
	DeleteEventsByIDs(ids []string) error
	GetEventsByPattern(filter Filter) ([]string, error)
	GetEventsForURL(url string) ([]events.Event, error)
	CreateFilter(url string, field FilterField, pattern regexp.Regexp) (Filter, error)
	DeleteFilter(url string, field FilterField, pattern regexp.Regexp) error
}
