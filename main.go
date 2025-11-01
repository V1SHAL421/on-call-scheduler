package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

type Schedule struct {
	Users                []string `json:"users"`
	HandoverStartDate    string   `json:"handover_start_at"`
	HandoverIntervalDays int      `json:"handover_interval_days"`
}

func retrieveFlags() (*string, *string, *string, *string, error) {
	schedule := flag.String("schedule", "", "path to schedule JSON file")
	overrides := flag.String("overrides", "", "path to overrides JSON file")
	from := flag.String("from", "", "start time for the schedule")
	until := flag.String("until", "", "end time for the schedule")

	if *schedule == "" {
		return nil, nil, nil, nil, fmt.Errorf("schedule flag is required")
	}
	if *overrides == "" {
		return nil, nil, nil, nil, fmt.Errorf("overrides flag is required")
	}
	if *from == "" {
		return nil, nil, nil, nil, fmt.Errorf("from flag is required")
	}
	if *until == "" {
		return nil, nil, nil, nil, fmt.Errorf("until flag is required")
	}

	return schedule, overrides, from, until, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Retrieve flags
	schedule, overrides, from, until, err := retrieveFlags()

	if err != nil {
		slog.Error("Error retrieving flags", "error", err)
	}
	flag.Parse()

	slog.Info("Schedule", "schedule", *schedule)
	slog.Info("Overrides", "overrides", *overrides)
	slog.Info("From", "from", *from)
	slog.Info("Until", "until", *until)

}
