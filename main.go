package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calendar struct {
	service *calendar.Service
	calIds  []string
	config  *oauth2.Config
}

func main() {
	ctx := context.Background()
	cal := newCalendar(ctx)

	var events *calendar.Events
	tomorrowFlag := viper.GetBool("tomorrow")
	if tomorrowFlag {
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

	calendarIds := viper.GetStringSlice("calendarIds")
	if len(calendarIds) == 0 {
		log.Fatalf("No calendarIds specified in config")
	}

	cal := Calendar{
		service: srv,
		calIds:  calendarIds,
		config:  config,
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
		retrievErr := &oauth2.RetrieveError{}
		if errors.As(err, &retrievErr) {
			removeToken(tokFile)
			getTokenFromWeb(c.config)
			return c.getEvents(since, upto)
		} else if err != nil {
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
	pflag.BoolP("tomorrow", "t", false, "should fetch  tomorrow's events")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetDefault("calendarIds", []string{"primary"})

	configFile := "config"
	configType := "yaml"
	configPath := "$HOME/.gocal"

	viper.SetConfigName(configFile) // name of config file (without extension)
	viper.SetConfigType(configType) // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(configPath) // call multiple times to add many search paths
	err := viper.ReadInConfig()     // Find and read the config file
	if err != nil {                 // Handle errors reading the config file
		log.Printf("no config file found, creating default at: %s/%s.%s", configPath, configFile, configType)
		err := viper.SafeWriteConfig()
		if err != nil {
			log.Fatalln("failed to write default config file", err)
		}
	}
}
