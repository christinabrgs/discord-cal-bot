package main

import (
	"log/slog"
	"os"

	bot "git.phlcode.club/discord-bot/bot"
	"git.phlcode.club/discord-bot/database"
	"git.phlcode.club/discord-bot/utils"
)

func main() {
	utils.InitEnv()
	e := utils.GetEnv()
	db, err := database.InitDatabase(e.DBPath)
	if err != nil {
		slog.Error("error initializing DB connection", slog.Any("error", err))
		os.Exit(66)
	}
	slog.Debug("database initialized...")

	err = bot.Run(db, e.DiscordToken)
	if err != nil {
		slog.Error("unable to start bot", slog.Any("error", err))
		os.Exit(3)
	}
}
