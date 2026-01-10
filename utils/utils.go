package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/joho/godotenv"
)

var ErrMissingProperty = errors.New("property is required but missing from event")

func HandleICSProp(prop *ics.IANAProperty, required bool, handler func(val string) error) error {
	if prop != nil {
		return handler(prop.Value)
	} else if required {
		return ErrMissingProperty
	}
	return nil
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

type Env struct {
	DiscordToken string
	DiscordAppID string
	DBPath       string
}

// check for env variable, to load dot env file
// if no file skip and read env variables
// if file load env and read env variables
// https://github.com/forkd-app/forkd/blob/main/api/util/env.go

var e Env

func InitEnv() {
	Skipenv, exists := os.LookupEnv("SKIP_ENV")
	if !exists || Skipenv != "true" {
		if err := godotenv.Load(); err != nil {
			slog.Error("Error loading .env file", slog.Any("error", err))
			os.Exit(66)
		}
	}

	token, exists := os.LookupEnv("DISCORD_TOKEN")
	if !exists {
		slog.Error("specify the DISCORD_TOKEN environment variable")
		os.Exit(64)
	}

	id, exists := os.LookupEnv("DISCORD_APP_ID")
	if !exists {
		slog.Error("specify the DISCORD_APP_ID environment variable")
		os.Exit(64)
	}

	path, exists := os.LookupEnv("DB_PATH")
	if !exists {
		path = "/calendars.db"
	}

	e = Env{DBPath: path, DiscordToken: token, DiscordAppID: id}
}

func GetEnv() Env {
	return e
}
