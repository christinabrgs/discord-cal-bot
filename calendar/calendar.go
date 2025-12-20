package calendar

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	e "git.phlcode.club/discord-bot/events"
	s "git.phlcode.club/discord-bot/store"
	ics "github.com/arran4/golang-ical"
	"github.com/bwmarrin/discordgo"
)

type Cal struct {
	logger  slog.Logger
	session *discordgo.Session
	s       s.Store
}

func NewCalendarCommands(logger slog.Logger, s s.Store, session *discordgo.Session) Commands {
	return Cal{
		logger:  logger,
		s:       s,
		session: session,
	}
}

// Events implements Commands.
func (c Cal) Events(url string) ([]e.Event, error) {
	return c.s.GetEventsForURL(url)
}

func (c Cal) Subscribe(url string, i *discordgo.InteractionCreate, filter *s.Filter) error {
	content := "Subscribing to calendar at: " + url
	err := c.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		slog.Default().Error("error sending response to subscribe command", slog.Any("error", err))
	}
	cal, err := ics.ParseCalendarFromUrl(url)
	if err != nil {
		return errors.Join(errors.New("unable to fetch and parse remote ics"), err)
	}
	content += "\nParsed calendar"
	_, err = c.session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		slog.Default().Error("error editing response to subscribe command", slog.Any("error", err))
	}

	result, err := c.s.InsertURL(url)
	if err != nil {
		return fmt.Errorf("error inserting calendar into database: %w", err)
	}

	fmt.Println("Inserted calendar with URL:", result)

	content += "\nParsing events..."
	_, err = c.session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		slog.Default().Error("error editing response to subscribe command", slog.Any("error", err))
	}
	events := make([]e.Event, 0, len(cal.Events()))
	for _, event := range cal.Events() {
		var currEvent e.Event
		err := currEvent.ParseFromiCal(event)
		if err != nil {
			c.logger.Error("error parsing ical event", slog.Any("event", event), slog.Any("error", err))
			continue
		}

		// Skip creating if it should be filtered
		if filter != nil && !filter.Filter(currEvent) {
			fmt.Printf("filtered: \n%s \n%s \n", filter.Pattern.String(), currEvent.Name)
			continue
		}

		// Skip creating if it is in the past
		if currEvent.StartTime.Before(time.Now()) {
			fmt.Println("Past event")
			continue
		}

		event, err := c.session.GuildScheduledEventCreate(i.GuildID, &discordgo.GuildScheduledEventParams{
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

		currEvent.ID = event.ID

		_, err = c.s.InsertEvent(url, currEvent)
		if err != nil {
			return fmt.Errorf("error inserting event into database: %w", err)
		}

		events = append(events, currEvent)
		content += "\nAdded event " + currEvent.Name
		_, err = c.session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			slog.Default().Error("error editing response to subscribe command", slog.Any("error", err))
		}
	}

	msg := fmt.Sprintf("subscribed to calendar at url %s with %d events...", url, len(events))
	content += "\n" + msg
	_, err = c.session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	slog.Info(msg, slog.String("url", url), slog.Any("events", events))
	return err
}

func (c Cal) Unsubscribe(url string, i *discordgo.InteractionCreate) error {
	// TODO: This should really be a transaction
	ids, err := c.s.DeleteEventsByURL(url)
	if err != nil {
		return fmt.Errorf("error deleting events from database: %w", err)
	}
	eventDeleteErrors := make([]error, 0)
	for _, id := range ids {
		err := c.session.GuildScheduledEventDelete(i.GuildID, id)
		if err != nil {
			eventDeleteErrors = append(eventDeleteErrors, err)
		}
	}
	if len(eventDeleteErrors) > 0 {
		return fmt.Errorf("error deleting events from discord: %+v", eventDeleteErrors)
	}
	_, err = c.s.DeleteCalendarByURL(url)
	if err != nil {
		return fmt.Errorf("error deleting calendar from database: %w", err)
	}
	return nil
}

func (c Cal) Filter(url, field, pattern string, i *discordgo.InteractionCreate) error {
	var f s.FilterField
	switch field {
	case s.FilterFieldName:
		f = s.FilterFieldName
	case s.FilterFieldDescription:
		f = s.FilterFieldDescription
	case s.FilterFieldLocation:
		f = s.FilterFieldLocation
	default:
		return fmt.Errorf("invalid filter field: %s", field)
	}
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %s", err)
	}
	filter, err := c.s.CreateFilter(url, f, *regex)
	if err != nil {
		return fmt.Errorf("unable to store filter: %s", err)
	}
	ids, err := c.s.GetEventsByPattern(filter)
	if err != nil {
		return fmt.Errorf("unable to fetch events from db: %s", err)
	}
	eventDeleteErrors := make([]error, 0)
	for _, id := range ids {
		err := c.session.GuildScheduledEventDelete(i.GuildID, id)
		if err != nil {
			eventDeleteErrors = append(eventDeleteErrors, err)
		}
	}
	if len(eventDeleteErrors) > 0 {
		return fmt.Errorf("discord event delete errors: %+v", eventDeleteErrors)
	}
	err = c.s.DeleteEventsByIDs(ids)
	if err != nil {
		return fmt.Errorf("unable to delete events from db: %s", err)
	}
	return err
}
