package calendar

import (
	"regexp"

	e "git.phlcode.club/discord-bot/events"
)

type Commands interface {
	Subscribe(url string, guildID string) error
	Unsubscribe(url string) error
	Filter(url string, field string, pattern regexp.Regexp) error
	Events() map[string][]e.Event
}
