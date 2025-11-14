package main

import (
	"log"
	"os"

	bot "git.phlcode.club/discord-bot/bot"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("discord_token")

	bot.BotToken = token
	bot.Run()
}
