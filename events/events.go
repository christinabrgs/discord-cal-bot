package events

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	u "git.phlcode.club/discord-bot/utils"
	ics "github.com/arran4/golang-ical"
)

type Event struct {
	ID          string
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
}

func (e *Event) ParseFromiCal(event *ics.VEvent) error {
	err := u.HandleICSProp(event.GetProperty(ics.ComponentPropertySummary), true, func(val string) error {
		e.Name = val
		return nil
	})
	if err != nil {
		return errors.New("name (summary) is required but missing from event")
	}
	err = u.HandleICSProp(event.GetProperty(ics.ComponentPropertyDtStart), true, func(val string) error {
		startTime, err := u.ParseTime(val)
		if err != nil {
			return fmt.Errorf("unable to parse start time: %s", err.Error())
		}
		e.StartTime = startTime
		return nil
	})
	if err != nil {
		if errors.Is(err, u.ErrMissingProperty) {
			return errors.New("start time is required but missing from event")
		}
		return errors.Join(errors.New("error handling start time: "), err)
	}
	err = u.HandleICSProp(event.GetProperty(ics.ComponentPropertyDtEnd), false, func(val string) error {
		endTime, err := u.ParseTime(val)
		if err != nil {
			return fmt.Errorf("unable to parse end time: %s", err.Error())
		}
		e.EndTime = endTime
		return nil
	})
	if err != nil {
		return errors.Join(errors.New("error handling end time: "), err)
	}
	err = u.HandleICSProp(event.GetProperty(ics.ComponentPropertyDescription), false, func(val string) error {
		e.Description = val
		return nil
	})
	if err != nil {
		slog.Default().Warn("Err was not nil when parsing optional event description", "error", err)
		// This is purposefull empty because we should never get here since this isn't required
	}
	err = u.HandleICSProp(event.GetProperty(ics.ComponentPropertyLocation), false, func(val string) error {
		e.Location = val
		return nil
	})
	if err != nil {
		slog.Default().Warn("Err was not nil when parsing optional event location", "error", err)
		// This is purposefull empty because we should never get here since this isn't required
	}
	return nil
}
