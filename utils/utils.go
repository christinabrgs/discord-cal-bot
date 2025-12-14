package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
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
