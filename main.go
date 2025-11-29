package main

import (
	"log"
	"log/slog"
	"os"

	bot "git.phlcode.club/discord-bot/bot"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("DISCORD_TOKEN")

	bot.BotToken = token
	err = bot.Run()
	if err != nil {
		slog.Error("unable to start bot", slog.Any("error", err))
	}
}
