package main

import (
	"log/slog"
	"os"

	bot "git.phlcode.club/discord-bot/bot"
	"git.phlcode.club/discord-bot/database"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", slog.Any("error", err))
		os.Exit(66)
	}

	dbPath, ok := os.LookupEnv("DB_PATH")
	if !ok {
		slog.Error("specify the DB_PATH environment variable")
		os.Exit(64)
	}
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		slog.Error("error initializing DB connection", slog.Any("error", err))
		os.Exit(66)
	}
	slog.Debug("database initialized...")

	token, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
		slog.Error("specify the DISCORD_TOKEN environment variable")
		os.Exit(64)
	}

	err = bot.Run(db, token)
	if err != nil {
		slog.Error("unable to start bot", slog.Any("error", err))
		os.Exit(3)
	}
}
