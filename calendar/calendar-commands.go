package calendar

import (
	e "git.phlcode.club/discord-bot/events"
	"git.phlcode.club/discord-bot/store"
	"github.com/bwmarrin/discordgo"
)

type Commands interface {
	Subscribe(url string, i *discordgo.InteractionCreate, filter *store.Filter) error
	Unsubscribe(url string, i *discordgo.InteractionCreate) error
	Filter(url, field, pattern string, i *discordgo.InteractionCreate) error
	Events(url string) ([]e.Event, error)
}
