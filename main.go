package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	bot "git.phlcode.club/discord-bot/bot"
	"git.phlcode.club/discord-bot/database"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbPath := os.Getenv("DB_PATH")
	if len(dbPath) == 0 {
		log.Fatal("specify the DB_PATH environment variable")
	}
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		log.Fatal("error initializing DB connection: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("error initializing DB connection: ping error: ", err)
	}
	fmt.Println("database initialized..")

	token := os.Getenv("DISCORD_TOKEN")
	appID := os.Getenv("DISCORD_APP_ID")

	log.Printf("Token length: %d", len(token))
	log.Printf("App ID: %s", appID)

	bot.BotToken = token
	err = bot.Run()
	if err != nil {
		slog.Error("unable to start bot", slog.Any("error", err))
	}
}
