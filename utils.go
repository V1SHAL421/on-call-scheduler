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

	result := InitialScheduleIntervals{}
	i, j := 0, 0

	for i < len(initialSchedule) {
		current := initialSchedule[i]
		processed := false

		for j < len(overrides) {
			override := overrides[j]
			if override.To.Before(current.From) || override.To.Equal(current.From) {
				j++
				continue
			}
			if override.From.After(current.To) || override.From.Equal(current.To) {
				break
			}

			if override.From.After(current.From) && !current.From.Equal(override.From) {
				result = append(result, InitialScheduleInterval{
					User: current.User,
					From: current.From,
					To:   override.From,
				})
			}

			if !override.From.Equal(override.To) {
				result = append(result, InitialScheduleInterval{
					User: override.User,
					From: override.From,
					To:   override.To,
				})
			}

			if override.To.Before(current.To) {
				current.From = override.To
				j++
			} else {
				processed = true
				j++
				break
			}
		}

		if !processed {
			result = append(result, current)
		}
		i++
	}

	return result, nil
}

func createFinalSchedule(scheduleWithOverrides InitialScheduleIntervals, fromTime time.Time, untilTime time.Time) (InitialScheduleIntervals, error) {
	if fromTime.After(untilTime) || fromTime.Equal(untilTime) {
		return nil, fmt.Errorf("from time must be before until time")
	}

	fmt.Println(scheduleWithOverrides)

	start, end := -1, -1
	for i, interval := range scheduleWithOverrides {
		if start == -1 && interval.To.After(fromTime) {
			start = i
		}
		if interval.From.Before(untilTime) {
			end = i
		}
	}

	if start == -1 || end == -1 {
		return InitialScheduleIntervals{}, nil
	}

	result := make(InitialScheduleIntervals, end-start+1)
	copy(result, scheduleWithOverrides[start:end+1])

	if result[0].From.Before(fromTime) {
		result[0].From = fromTime
	}
	if result[len(result)-1].To.After(untilTime) {
		result[len(result)-1].To = untilTime
	}

	return result, nil
}
