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

	// Retrieve flags
	schedule, overrides, from, until, retrieveFlagsErr := retrieveFlags()

	if retrieveFlagsErr != nil {
		slog.Error("Error retrieving flags", "error", retrieveFlagsErr)
		os.Exit(1)
	}
	flag.Parse()

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

}
