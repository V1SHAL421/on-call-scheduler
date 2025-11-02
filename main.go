// On-Call Scheduler generates on-call schedules with override support
package main

import (
	"flag"
	"log/slog"
	"on-call-scheduler/src"
	"os"
	"time"
)

// main processes command line arguments and generates an on-call schedule
// by combining a base schedule plan with override intervals
func main() {
	// Initialize structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Define command line flags
	schedule := flag.String("schedule", "", "path to schedule JSON file")
	overrides := flag.String("overrides", "", "path to overrides JSON file")
	from := flag.String("from", "", "start time for the schedule (RFC3339 format)")
	until := flag.String("until", "", "end time for the schedule (RFC3339 format)")
	flag.Parse()

	// Validate all required flags are provided
	if *schedule == "" || *overrides == "" || *from == "" || *until == "" {
		slog.Error("All flags are required")
		os.Exit(1)
	}

	// Log input parameters
	slog.Info("Schedule", "schedule", *schedule)
	slog.Info("Overrides", "overrides", *overrides)
	slog.Info("From", "from", *from)
	slog.Info("Until", "until", *until)

	// Read JSON files
	scheduleFile, scheduleFileErr := os.ReadFile(*schedule)
	if scheduleFileErr != nil {
		slog.Error("Error reading schedule file", "error", scheduleFileErr)
		os.Exit(1)
	}
	overridesFile, overridesFileErr := os.ReadFile(*overrides)
	if overridesFileErr != nil {
		slog.Error("Error reading overrides file", "error", overridesFileErr)
		os.Exit(1)
	}

	// Parse JSON into structs
	scheduleStruct, scheduleStructErr := src.ParseFieldToStruct(scheduleFile, "schedule")
	if scheduleStructErr != nil {
		slog.Error("Error parsing schedule field", "error", scheduleStructErr)
		os.Exit(1)
	}
	overridesStruct, overridesStructErr := src.ParseFieldToStruct(overridesFile, "overrides")
	if overridesStructErr != nil {
		slog.Error("Error parsing overrides field", "error", overridesStructErr)
		os.Exit(1)
	}

	// Parse time strings into time.Time objects
	fromTime, fromTimeErr := time.Parse(time.RFC3339, *from)
	if fromTimeErr != nil {
		slog.Error("Invalid from time format", "error", fromTimeErr)
		os.Exit(1)
	}

	untilTime, untilTimeErr := time.Parse(time.RFC3339, *until)
	if untilTimeErr != nil {
		slog.Error("Invalid until time format", "error", untilTimeErr)
		os.Exit(1)
	}

	// Type cast parsed structs to expected types
	schedulePlan, schedulePlanErr := scheduleStruct.(src.SchedulePlan)
	if !schedulePlanErr {
		slog.Error("Error casting schedule to SchedulePlan")
		os.Exit(1)
	}

	overridesCasted, overridesCastedErr := overridesStruct.(src.Overrides)
	if !overridesCastedErr {
		slog.Error("Error casting overrides to OverridesPlan")
		os.Exit(1)
	}

	// Generate initial schedule based on rotation plan
	initialSchedule, initialScheduleErr := src.CreateInitialSchedule(schedulePlan, fromTime, untilTime)
	if initialScheduleErr != nil {
		slog.Error("Error creating initial schedule", "error", initialScheduleErr)
		os.Exit(1)
	}

	// Apply overrides to the initial schedule
	newSchedule, newScheduleErr := src.AddOverridesToSchedule(initialSchedule, overridesCasted)
	if newScheduleErr != nil {
		slog.Error("Error applying overrides", "error", newScheduleErr)
		os.Exit(1)
	}

	// Output the final schedule
	slog.Info("Final schedule", "finalSchedule", newSchedule)
}
