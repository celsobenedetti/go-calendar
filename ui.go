package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
)

//   - `12:00` My Event Title
//     > Description of event, if any
func toMarkdownList(events *calendar.Events) string {
	var out strings.Builder
	for _, item := range events.Items {
		writeListItem(item, out)
	}
	return out.String()
}

func writeListItem(item *calendar.Event, out strings.Builder) {
	date, err := time.Parse(time.RFC3339, item.Start.DateTime)
	if err != nil {
		date, err = time.Parse(time.RFC3339, item.Start.Date)
		if err != nil {
			log.Fatalf("failed to parse date %v", err)
		}
	}

	out.WriteString(fmt.Sprintf("- `%d:%.2d` %v", date.Hour(), date.Minute(), item.Summary))
	if len(item.Description) > 0 {
		out.WriteString("\n\t> " + item.Description)
	}
	out.WriteString("\n")
}
