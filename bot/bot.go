package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var BotToken string

func checkNilErr(e error) {
	if e != nil {
		log.Fatal("Error message")
	}
}

func Run() {
	discord, err := discordgo.New("Bot " + BotToken)
	checkNilErr(err)

	discord.AddHandler(newMessage)

	discord.Open()
	defer discord.Close()

	fmt.Println("bot running...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	log.Printf("%+v", message)
	if message.Author.ID == discord.State.User.ID {
		return
	}
	switch {
	case strings.Contains(message.Content, "help!"):
		discord.ChannelMessageSend(message.ChannelID, "Hello World, the Robots are taking over")
	case strings.Contains(message.Content, "bye!"):
		discord.ChannelMessageSend(message.ChannelID, "GoodBye!")

	}
}
