package calendar

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	e "git.phlcode.club/discord-bot/events"
	s "git.phlcode.club/discord-bot/store"
	ics "github.com/arran4/golang-ical"
	"github.com/bwmarrin/discordgo"
)

type Cal struct {
	logger  slog.Logger
	events  map[string][]e.Event
	session *discordgo.Session
	s       s.Store
}

func NewCalendarCommands(logger slog.Logger, s s.Store, session *discordgo.Session) Commands {
	return Cal{
		logger:  logger,
		events:  make(map[string][]e.Event),
		s:       s,
		session: session,
	}
}

// Events implements Commands.
func (c Cal) Events() map[string][]e.Event {
	return c.events
}

func (c Cal) Subscribe(url string, guildID string) error {
	cal, err := ics.ParseCalendarFromUrl(url)
	if err != nil {
		return errors.Join(errors.New("unable to fetch and parse remote ics"), err)
	}

	result, err := c.s.InsertURL(url)
	if err != nil {
		return fmt.Errorf("error inserting calendar into database: %w", err)
	}

	fmt.Println("Inserted calendar with URL:", result)

	events := make([]e.Event, len(cal.Events()))
	for i, event := range cal.Events() {
		var currEvent e.Event
		err := currEvent.ParseFromiCal(event)
		if err != nil {
			c.logger.Error("error parsing ical event", slog.Any("event", event), slog.Any("error", err))
			continue
		}

		result, err := c.s.InsertEvent(url, currEvent)
		if err != nil {
			return fmt.Errorf("error inserting event into database: %w", err)
		}
		events[i] = currEvent

		fmt.Println("Inserted event with ID:", result)

		event, err := c.session.GuildScheduledEventCreate(guildID, &discordgo.GuildScheduledEventParams{
			Name:               currEvent.Name,
			Description:        currEvent.Description,
			ScheduledStartTime: &currEvent.StartTime,
			ScheduledEndTime:   &currEvent.EndTime,
			Status:             discordgo.GuildScheduledEventStatusScheduled,
			EntityType:         discordgo.GuildScheduledEventEntityTypeExternal,
			EntityMetadata: &discordgo.GuildScheduledEventEntityMetadata{
				Location: currEvent.Location,
			},
			PrivacyLevel: discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
		})
		if err != nil {
			return fmt.Errorf("error creating discord guild scheduled event: %w", err)
		}

		// store discord event ID in the database associated with this event
		_ = event
	}

	// c.events[url] = events

	msg := fmt.Sprintf("subscribed to calendar at url %s with %d events...", url, len(events))
	slog.Info(msg, slog.String("url", url), slog.Any("events", c.events))

	return nil
}

func (c Cal) Unsubscribe(url string) error {
	// TODO send message for successful deletion to discord
	_, err := c.s.DeleteCalendarByURL(url)
	if err != nil {
		return fmt.Errorf("error deleting calendar from database: %w", err)
	}
	_, err = c.s.DeleteEventsByURL(url)
	if err != nil {
		return fmt.Errorf("error deleting events from database: %w", err)
	}

	return nil
}

func (c Cal) Filter(url string, field string, pattern regexp.Regexp) error {
	c.logger.Warn("method `Filter` not implemented")
	return nil
}
