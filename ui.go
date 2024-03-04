package main

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
)

//   - `12:00` My Event Title
//     > Description of event, if any
func toMarkdownList(events *calendar.Events) string {
	var out strings.Builder
	for _, item := range events.Items {
		hour := getHour(item)
		out.WriteString(fmt.Sprintf("- [ ] `%d:%.2d` %v", hour.Hour(), hour.Minute(), item.Summary))

		if len(item.Description) > 0 {
			out.WriteString("\n\t> " + item.Description)
		}
		out.WriteString("\n")
	}
	return out.String()
}

func getHour(item *calendar.Event) time.Time {
	date, err := time.Parse(time.RFC3339, item.Start.DateTime)
	if err != nil {
		date, _ = time.Parse(time.RFC3339, item.Start.Date)
	}
	return date
}
