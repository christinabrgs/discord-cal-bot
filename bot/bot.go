// Package bot gives the scaffolding for the discord bot structure
// This includes:
// - Application commands
// - Event management
package bot

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"os/signal"

	c "git.phlcode.club/discord-bot/calendar"
	"git.phlcode.club/discord-bot/store"
	"git.phlcode.club/discord-bot/utils"
	"github.com/bwmarrin/discordgo"
)

var errMissingProperty = errors.New("property is required but missing from event")

type filterField = string

const (
	NameField        filterField = "name"
	DescriptionField filterField = "description"
	LocationField    filterField = "location"
)

var (
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
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cmd c.Commands){
		"subscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate, cmd c.Commands) {
			content := ""
			options := i.ApplicationCommandData().Options
			switch len(options) {
			case 0:
				content = "Input error: missing URL"
			case 1:
				url := options[0].StringValue()
				err := cmd.Subscribe(url, i, nil)
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
				url := options[0].StringValue()
				field := options[1].StringValue()
				pattern := options[2].StringValue()
				filter, err := store.NewFilter(url, field, pattern)
				if err != nil {
					content = "Error subscribing with filter: " + err.Error()
					break
				}
				err = cmd.Subscribe(url, i, filter)
				if err != nil {
					content = "Error subscribing to calendar: " + err.Error()
					break
				}
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
		"unsubscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate, cmd c.Commands) {
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
		"filter": func(s *discordgo.Session, i *discordgo.InteractionCreate, cmd c.Commands) {
			options := i.ApplicationCommandData().Options
			url := options[0].StringValue()
			field := options[1].StringValue()
			pattern := options[2].StringValue()
			err := cmd.Filter(url, field, pattern, i)
			if err != nil {
				slog.Default().Error("error filtering events", slog.String("url", url), slog.String("field", field), slog.String("pattern", pattern), slog.Any("error", err))
			}
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Filtered events",
				},
			})
			if err != nil {
				slog.Default().Error("error sending response to subscribe command", slog.Any("error", err))
			}
		},
	}
)

func Run(db *sql.DB, token string) error {
	store := store.NewSQLiteStore(db)
	e := utils.GetEnv()
	appID := e.DiscordAppID
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return errors.Join(errors.New("invalid bot config"), err)
	}
	logger := slog.Default()
	if os.Getenv("ENV") == "prod" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	// TODO: Replace the default logger with a nicer library
	cmds := c.NewCalendarCommands(*logger, store, discord)
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i, cmds)
		}
	})
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(appID, "", v)
		if err != nil {
			logger.Error("Cannot create command", slog.String("cmdName", v.Name), slog.Any("error", err))
			return err
		}
		registeredCommands[i] = cmd
	}

	err = discord.Open()
	if err != nil {
		logger.Error("unable to open discord socket", slog.Any("error", err))
		return err
	}
	defer func() {
		if err := discord.Close(); err != nil {
			logger.Error("unable to close discod socket", slog.Any("error", err))
		}
	}()

	logger.Info("bot running...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	for _, v := range registeredCommands {
		err := discord.ApplicationCommandDelete(discord.State.User.ID, "", v.ID)
		if err != nil {
			logger.Error("Cannot delete command", slog.String("name", v.Name), slog.Any("error", err))
		}
	}
	return nil
}
