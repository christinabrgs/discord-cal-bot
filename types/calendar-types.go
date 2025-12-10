package types

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

type filterField = string

type Commands interface {
	Subscribe(url string, guildID string) error
	Unsubscribe(url string) error
	Filter(url string, field filterField, pattern regexp.Regexp) error
	Events() map[string][]Event
}

type Event struct {
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
}

var errMissingProperty = errors.New("property is required but missing from event")

func HandleICSProp(prop *ics.IANAProperty, required bool, handler func(val string) error) error {
	if prop != nil {
		return handler(prop.Value)
	} else if required {
		return errMissingProperty
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

func (e Event) ParseFromiCal(event *ics.VEvent) error {
	err := HandleICSProp(event.GetProperty(ics.ComponentPropertySummary), true, func(val string) error {
		e.Name = val
		return nil
	})
	if err != nil {
		return errors.New("name (summary) is required but missing from event")
	}
	err = HandleICSProp(event.GetProperty(ics.ComponentPropertyDtStart), true, func(val string) error {
		startTime, err := ParseTime(val)
		if err != nil {
			return fmt.Errorf("unable to parse start time: %s", err.Error())
		}
		e.StartTime = startTime
		return nil
	})
	if err != nil {
		if errors.Is(err, errMissingProperty) {
			return errors.New("start time is required but missing from event")
		}
		return errors.Join(errors.New("error handling start time: "), err)
	}
	err = HandleICSProp(event.GetProperty(ics.ComponentPropertyDtEnd), false, func(val string) error {
		endTime, err := ParseTime(val)
		if err != nil {
			return fmt.Errorf("unable to parse end time: %s", err.Error())
		}
		e.EndTime = endTime
		return nil
	})
	if err != nil {
		return errors.Join(errors.New("error handling end time: "), err)
	}
	err = HandleICSProp(event.GetProperty(ics.ComponentPropertyDescription), false, func(val string) error {
		e.Description = val
		return nil
	})
	if err != nil {
		slog.Default().Warn("Err was not nil when parsing optional event description", "error", err)
		// This is purposefull empty because we should never get here since this isn't required
	}
	err = HandleICSProp(event.GetProperty(ics.ComponentPropertyLocation), false, func(val string) error {
		e.Location = val
		return nil
	})
	if err != nil {
		slog.Default().Warn("Err was not nil when parsing optional event location", "error", err)
		// This is purposefull empty because we should never get here since this isn't required
	}
	return nil
}
