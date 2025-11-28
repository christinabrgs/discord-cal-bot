package bot

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/bwmarrin/discordgo"
)

var (
	BotToken string
	s        *discordgo.Session
)

type Commands interface {
	Subscribe(url string) ([]Event, error)
	Unsubscribe(url string)
	Filter(word string)
}

type Event struct {
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
}

type Cal struct {
	Events []Event
}

func (c Cal) Subscribe(url string) ([]Event, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cal, err := ics.ParseCalendar(resp.Body)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	fmt.Println(cal.Events())
	return nil, nil
}

func (c Cal) Unsubscribe(url string) {
	fmt.Print("not implemented")
}

func (c Cal) Filter(url string) {
	fmt.Print("not implemented")
}

func checkNilErr(e error) {
	if e != nil {
		log.Fatal("Error message")
	}
}

func Run() {
	discord, err := discordgo.New("Bot " + BotToken)
	checkNilErr(err)

	d := Cal{}
	d.Subscribe("https://api2.luma.com/ics/get?entity=calendar&id=cal-s0hIkjIqVvTXRBG")
	// d.Unsubscribe("string")
	// d.Filter("string")

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
