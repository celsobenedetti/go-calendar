package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calendar struct {
	service *calendar.Service
	calIds  []string
}

var tomorrowFlag = flag.Bool("tom", false, "fetch tomorrow's events")

func main() {
	ctx := context.Background()
	cal := newCalendar(ctx)

	var events *calendar.Events
	if *tomorrowFlag {
		events = cal.tomorrow()
	} else {
		events = cal.today()
	}

	if len(events.Items) == 0 {
		fmt.Print("No upcoming events found.")
		os.Exit(0)
	}

	fmt.Print(toMarkdownList(events))
}

func newCalendar(ctx context.Context) Calendar {
	config := readConfig()
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	cal := Calendar{
		service: srv,
		calIds: []string{
			"primary",
			"celso.benedetti@ocelotbot.com",
		},
	}
	return cal
}

func (c Calendar) getEvents(since, upto time.Time) *calendar.Events {
	result := &calendar.Events{}

	min, max := since.Format(time.RFC3339), upto.Format(time.RFC3339)

	for _, calId := range c.calIds {
		events, err := c.service.Events.List(calId).ShowDeleted(false).
			SingleEvents(true).
			TimeMin(min).
			TimeMax(max).
			MaxResults(10).
			OrderBy("startTime").
			Do()
		if err != nil {
			log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
		}

		result.Items = append(result.Items, events.Items...)
	}

	return result
}

func (c Calendar) today() *calendar.Events {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return c.getEvents(now, midnight)
}

func (c Calendar) tomorrow() *calendar.Events {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	tomMidnight := time.Date(midnight.Year(), midnight.Month(), midnight.Day()+1, 0, 0, 0, 0, midnight.Location())
	return c.getEvents(midnight, tomMidnight)
}

func init() {
	flag.Parse()
}
