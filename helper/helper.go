package helper

import (
	"strings"
	"time"
)

func ToPtr[T any](s T) *T { return &s }

type Datetime struct {
	*time.Time
}

type Date struct {
	*time.Time
}

func (d *Date) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), "\"")
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
	if err != nil {
		return err
	}

	*d = Date{&t}
	return nil
}

func (d *Datetime) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), "\"")
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.ParseInLocation(time.DateTime, value, time.UTC)
	if err != nil {
		return err
	}

	*d = Datetime{&t}

	return nil
}
