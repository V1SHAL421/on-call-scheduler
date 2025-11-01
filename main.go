package main

import (
	"flag"
	"log/slog"
	"os"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Define and parse flags
	schedule := flag.String("schedule", "", "path to schedule JSON file")
	overrides := flag.String("overrides", "", "path to overrides JSON file")
	from := flag.String("from", "", "start time for the schedule")
	until := flag.String("until", "", "end time for the schedule")
	flag.Parse()

	// Validate flags
	if *schedule == "" || *overrides == "" || *from == "" || *until == "" {
		slog.Error("All flags are required")
		os.Exit(1)
	}

	slog.Info("Schedule", "schedule", *schedule)
	slog.Info("Overrides", "overrides", *overrides)
	slog.Info("From", "from", *from)
	slog.Info("Until", "until", *until)

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

	scheduleStruct, scheduleStructErr := parseFieldToStruct(scheduleFile, "schedule")
	if scheduleStructErr != nil {
		slog.Error("Error parsing schedule field", "error", scheduleStructErr)
		os.Exit(1)
	}
	overridesStruct, overridesStructErr := parseFieldToStruct(overridesFile, "overrides")
	if overridesStructErr != nil {
		slog.Error("Error parsing overrides field", "error", overridesStructErr)
		os.Exit(1)
	}

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

	schedulePlan, schedulePlanErr := scheduleStruct.(SchedulePlan)
	if !schedulePlanErr {
		slog.Error("Error casting schedule to SchedulePlan")
		os.Exit(1)
	}

	overridesCasted, overridesCastedErr := overridesStruct.(Overrides)
	if !overridesCastedErr {
		slog.Error("Error casting overrides to OverridesPlan")
		os.Exit(1)
	}

	initialSchedule, initialScheduleErr := createInitialSchedule(schedulePlan, fromTime, untilTime)
	if initialScheduleErr != nil {
		slog.Error("Error creating initial schedule", "error", initialScheduleErr)
		os.Exit(1)
	}

	newSchedule, newScheduleErr := addOverridesToSchedule(initialSchedule, overridesCasted)
	if newScheduleErr != nil {
		slog.Error("Error applying overrides", "error", newScheduleErr)
		os.Exit(1)
	}

	finalSchedule, finalScheduleErr := createFinalSchedule(newSchedule, fromTime, untilTime)
	if finalScheduleErr != nil {
		slog.Error("Error creating final schedule", "error", finalScheduleErr)
		os.Exit(1)
	}
	// Print the final schedule
	slog.Info("Final schedule", "finalSchedule", finalSchedule)

}
