package types

import (
	"regexp"
	"time"
)

type filterField = string

type Commands interface {
	Subscribe(url string, guildID string) error
	Unsubscribe(url string) error
	Filter(url string, field filterField, pattern regexp.Regexp) error
	Events() map[string][]Event
}

type Event struct {
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
}
