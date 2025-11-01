package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"time"
)

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

func parseFieldToStruct(field []byte, name string) (interface{}, error) {
	if len(field) == 0 {
		return nil, fmt.Errorf("field %s is empty", name)
	}
	if field[0] != '{' {
		return nil, fmt.Errorf("field %s is not a JSON object", name)
	}
	if field[len(field)-1] != '}' {
		return nil, fmt.Errorf("field %s is not a JSON object", name)
	}
	switch name {
	case "schedule":
		var schedule SchedulePlan
		err := json.Unmarshal(field, &schedule)
		if err != nil {
			return nil, err
		}
		return schedule, nil

	case "overrides":
		var overrides Overrides
		err := json.Unmarshal(field, &overrides)
		if err != nil {
			return nil, err
		}
		return overrides, nil
	}

	return nil, fmt.Errorf("unknown field name: %s", name)
}

func createInitialSchedule(schedulePlan SchedulePlan, fromTime time.Time, endTime time.Time) (InitialScheduleIntervals, error) {
	if schedulePlan.HandoverIntervalDays <= 0 {
		return nil, fmt.Errorf("handover interval days must be greater than 0")
	}
	if len(schedulePlan.Users) == 0 {
		return nil, fmt.Errorf("users must be defined")
	}
	if fromTime.After(endTime) {
		return nil, fmt.Errorf("from time must be before end time")
	}
	if fromTime.Equal(endTime) {
		return nil, fmt.Errorf("from time must be before end time")
	}

	currTime := fromTime
	currUserIndex := 0
	var initialSchedule InitialScheduleIntervals

	for currTime.Before(endTime) {
		currScheduleInterval := InitialScheduleInterval{
			User: schedulePlan.Users[currUserIndex],
			From: currTime,
			To:   currTime.AddDate(0, 0, schedulePlan.HandoverIntervalDays),
		}
		initialSchedule = append(initialSchedule, currScheduleInterval)
		currTime = currTime.AddDate(0, 0, schedulePlan.HandoverIntervalDays)
		currUserIndex = (currUserIndex + 1) % len(schedulePlan.Users)
	}
	return initialSchedule, nil
}

func addOverridesToSchedule(initialSchedule InitialScheduleIntervals, overrides Overrides) (InitialScheduleIntervals, error) {
	if len(overrides) == 0 {
		return initialSchedule, nil
	}
	if len(initialSchedule) == 0 {
		return nil, fmt.Errorf("initial schedule must be defined")
	}
	sort.Slice(overrides, func(i, j int) bool {
		return overrides[i].From.Before(overrides[j].From)
	})

	return initialSchedule, nil
}
