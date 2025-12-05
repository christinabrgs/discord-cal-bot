// Package bot gives the scaffolding for the discord bot structure
// This includes:
// - Application commands
// - Event management
package bot

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"git.phlcode.club/discord-bot/database"
	ics "github.com/arran4/golang-ical"
	"github.com/bwmarrin/discordgo"
)

var (
	errMissingProperty = errors.New("property is required but missing from event")
	db, _              = database.InitDatabase(os.Getenv("DB_PATH"))
)

type filterField = string

const (
	NameField        filterField = "name"
	DescriptionField filterField = "description"
	LocationField    filterField = "location"
)

func handleICSProp(prop *ics.IANAProperty, required bool, handler func(val string) error) error {
	if prop != nil {
		return handler(prop.Value)
	} else if required {
		return errMissingProperty
	}
	return nil
}

type Commands interface {
	Subscribe(url string) error
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

func ParseTime(value string) (time.Time, error) {
	if strings.HasSuffix(value, "Z") {
		time, err := time.Parse(`20060102T150405Z`, value)
		if err != nil {
			return time, fmt.Errorf("unable to parse time: %s", err.Error())
		}
		return time, nil
	}
	time, err := time.Parse(`20060102T150405`, value)
	if err != nil {
		return time, fmt.Errorf("unable to parse time: %s", err.Error())
	}
	return time, nil
}

func (e *Event) ParseFromiCal(event *ics.VEvent) error {
	err := handleICSProp(event.GetProperty(ics.ComponentPropertySummary), true, func(val string) error {
		e.Name = val
		return nil
	})
	if err != nil {
		return errors.New("name (summary) is required but missing from event")
	}
	err = handleICSProp(event.GetProperty(ics.ComponentPropertyDtStart), true, func(val string) error {
		startTime, err := ParseTime(val)
		if err != nil {
			return fmt.Errorf("unable to parse start time: %s", err.Error())
		}
		e.StartTime = startTime
		return nil
	})
	if err != nil {
		if errors.Is(err, errMissingProperty) {
			return errors.New("start time is required but missing from event")
		}
		return errors.Join(errors.New("error handling start time: "), err)
	}
	err = handleICSProp(event.GetProperty(ics.ComponentPropertyDtEnd), false, func(val string) error {
		endTime, err := ParseTime(val)
		if err != nil {
			return fmt.Errorf("unable to parse end time: %s", err.Error())
		}
		e.EndTime = endTime
		return nil
	})
	if err != nil {
		return errors.Join(errors.New("error handling end time: "), err)
	}
	err = handleICSProp(event.GetProperty(ics.ComponentPropertyDescription), false, func(val string) error {
		e.Description = val
		return nil
	})
	if err != nil {
		slog.Default().Warn("Err was not nil when parsing optional event description", "error", err)
		// This is purposefull empty because we should never get here since this isn't required
	}
	err = handleICSProp(event.GetProperty(ics.ComponentPropertyLocation), false, func(val string) error {
		e.Location = val
		return nil
	})
	if err != nil {
		slog.Default().Warn("Err was not nil when parsing optional event location", "error", err)
		// This is purposefull empty because we should never get here since this isn't required
	}
	return nil
}

type Cal struct {
	logger slog.Logger
	// TODO: This should be updated to handle like an SQLite db or some sort of persisted KV store
	events map[string][]Event
}

// Events implements Commands.
func (c Cal) Events() map[string][]Event {
	return c.events
}

func (c Cal) Subscribe(url string) error {
	cal, err := ics.ParseCalendarFromUrl(url)
	if err != nil {
		return errors.Join(errors.New("unable to fetch and parse remote ics"), err)
	}

	calStatement := `INSERT INTO calendars (url, last_synced) VALUES (?, ?);`

	result, err := db.Exec(calStatement, url, time.Now())
	if err != nil {
		return fmt.Errorf("error inserting calendar into database: %w", err)
	}

	fmt.Println("Inserted calendar with URL:", result)

	events := make([]Event, len(cal.Events()))
	for i, event := range cal.Events() {
		var e Event
		err := e.ParseFromiCal(event)
		if err != nil {
			c.logger.Error("error parsing ical event", slog.Any("event", event), slog.Any("error", err))
			continue
		}

		eventStatement := `INSERT INTO events (calendar_url, name, description, start_time, end_time, location) VALUES (?, ?, ?, ?, ?, ?);`

		result, err := db.Exec(eventStatement, url, e.Name, e.Description, e.StartTime, e.EndTime, e.Location)
		if err != nil {
			return fmt.Errorf("error inserting event into database: %w", err)
		}
		events[i] = e

		fmt.Println("Inserted event with ID:", result)
	}

	c.events[url] = events

	msg := fmt.Sprintf("subscribed to calendar at url %s with %d events...", url, len(events))
	slog.Info(msg, slog.String("url", url), slog.Any("events", c.events))

	return nil
}

func (c Cal) Unsubscribe(url string) error {
	c.logger.Warn("method `Unsubscribe` not implemented")
	return nil
}

func (c Cal) Filter(url string, field filterField, pattern regexp.Regexp) error {
	c.logger.Warn("method `Filter` not implemented")
	return nil
}

var (
	BotToken  string
	eventPerm int64 = discordgo.PermissionManageEvents
	urlOpt          = discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "url",
		Description: "URL for remote calendar",
		Required:    true,
	}
	commands = []*discordgo.ApplicationCommand{
		{
			ID:                       "phl-code-club-cal-bot-subscribe",
			Name:                     "subscribe",
			Description:              "Subscribe CalendarBot to a remote calendar",
			DefaultMemberPermissions: &eventPerm,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			IntegrationTypes:         &[]discordgo.ApplicationIntegrationType{discordgo.ApplicationIntegrationGuildInstall},
			Options: []*discordgo.ApplicationCommandOption{
				&urlOpt,
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "field",
					Description: "field to filter on",
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "name",
							Value: "name",
						},
						{
							Name:  "summary",
							Value: "summary",
						},
						{
							Name:  "location",
							Value: "location",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "pattern",
					Description: "filter pattern",
				},
			},
		},
		{
			ID:                       "phl-code-club-cal-bot-unsubscribe",
			Name:                     "unsubscribe",
			Description:              "Unsubscribe CalendarBot from a remote calendar",
			DefaultMemberPermissions: &eventPerm,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			IntegrationTypes:         &[]discordgo.ApplicationIntegrationType{discordgo.ApplicationIntegrationGuildInstall},
			Options: []*discordgo.ApplicationCommandOption{
				&urlOpt,
			},
		},
		{
			ID:                       "phl-code-club-cal-bot-filter",
			Name:                     "filter",
			Description:              "Add filter to existing calendar and reprocess events",
			DefaultMemberPermissions: &eventPerm,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			IntegrationTypes:         &[]discordgo.ApplicationIntegrationType{discordgo.ApplicationIntegrationGuildInstall},
			Options: []*discordgo.ApplicationCommandOption{
				&urlOpt,
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "field",
					Required:    true,
					Description: "field to filter on",
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "name",
							Value: "name",
						},
						{
							Name:  "summary",
							Value: "summary",
						},
						{
							Name:  "location",
							Value: "location",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "pattern",
					Description: "filter pattern",
					Required:    true,
				},
			},
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cmd Commands){
		"subscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate, cmd Commands) {
			content := ""
			options := i.ApplicationCommandData().Options
			switch len(options) {
			case 0:
				content = "Input error: missing URL"
			case 1:
				url := options[0].StringValue()
				err := cmd.Subscribe(url)
				if err != nil {
					content = "Error subscribing to calendar: " + err.Error()
					break
				}
				content = "URL: " + url
			case 2:
				switch options[1].Name {
				case "field":
					content = "Input error: missing filter option `pattern`"
				case "pattern":
					content = "Input error: missing filter option `field`"
				default:
					content = "Input error: invalid input options, missing additional optional field"
				}
			case 3:
				// TODO: Add implementation for subscribe AND filter
				content = "SUBSCRIBE WITH FILTER"
			default:
				content = "Input error: invalid input options"
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
			if err != nil {
				slog.Default().Error("error sending response to subscribe command", slog.Any("error", err))
			}
		},
		"unsubscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate, cmd Commands) {
			content := ""
			options := i.ApplicationCommandData().Options
			switch len(options) {
			case 0:
				content = "Input error: missing URL"
			case 1:
				// TODO: Implement unsubscribe
				content = "UNSUBSCRIBE"
			default:
				content = "Input error: invalid input options"
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
			if err != nil {
				slog.Default().Error("error sending response to subscribe command", slog.Any("error", err))
			}
		},
		"filter": func(s *discordgo.Session, i *discordgo.InteractionCreate, cmd Commands) {
		},
	}
)

func Run() error {
	appID := os.Getenv("DISCORD_APP_ID")
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		return errors.Join(errors.New("invalid bot config"), err)
	}
	// TODO: Replace the default logger with a nicer library
	cmds := Cal{events: make(map[string][]Event), logger: *slog.Default()}
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i, cmds)
		}
	})
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(appID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	err = discord.Open()
	if err != nil {
		slog.Default().Error("unable to open discord socket", slog.Any("error", err))
	}
	defer func() {
		if err := discord.Close(); err != nil {
			slog.Default().Error("unable to close discod socket", slog.Any("error", err))
		}
	}()

	slog.Default().Info("bot running...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	for _, v := range registeredCommands {
		err := discord.ApplicationCommandDelete(discord.State.User.ID, "", v.ID)
		if err != nil {
			slog.Error("Cannot delete command", slog.String("name", v.Name), slog.Any("error", err))
		}
	}
	return nil
}
