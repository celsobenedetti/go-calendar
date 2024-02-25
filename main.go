package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calendar struct {
	service *calendar.Service
	calIds  []string
}

func main() {
	ctx := context.Background()
	cal := newCalendar(ctx)

	events := cal.today()
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
		os.Exit(0)
	}

	var sb strings.Builder

	for _, item := range events.Items {
		date, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			date, err = time.Parse(time.RFC3339, item.Start.Date)
			if err != nil {
				log.Fatalf("failed to parse date", err)
			}
		}

		sb.WriteString(fmt.Sprintf("- `%d:%.2d` %v", date.Hour(), date.Minute(), item.Summary))
		if len(item.Description) > 0 {
			sb.WriteString(" - " + item.Description)
		}
		sb.WriteString("\n")
	}

	fmt.Println(sb.String())
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
			"celsobenedetti2@gmail.com",
			"celso.benedetti@ocelotbot.com",
		},
	}
	return cal
}

func (c Calendar) getEvents(since, upto time.Time) *calendar.Events {
	result := &calendar.Events{}

	min, max := since.Format(time.RFC3339), upto.Format(time.RFC3339)
	_ = max

	for _, calId := range c.calIds {
		events, err := c.service.Events.List(calId).ShowDeleted(false).
			SingleEvents(true).
			TimeMin(min).
			// TimeMax(max).
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
