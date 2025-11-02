package src

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

func ParseFieldToStruct(field []byte, name string) (interface{}, error) {
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

func CreateInitialSchedule(schedulePlan SchedulePlan, fromTime time.Time, endTime time.Time) (InitialScheduleIntervals, error) {
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

func AddOverridesToSchedule(initialSchedule InitialScheduleIntervals, overrides Overrides) (InitialScheduleIntervals, error) {
	if len(overrides) == 0 {
		return initialSchedule, nil
	}
	if len(initialSchedule) == 0 {
		return nil, fmt.Errorf("initial schedule must be defined")
	}

	// Filter out zero-duration overrides
	validOverrides := Overrides{}
	for _, override := range overrides {
		if !override.From.Equal(override.To) {
			validOverrides = append(validOverrides, override)
		}
	}

	// If no valid overrides, return original schedule
	if len(validOverrides) == 0 {
		return initialSchedule, nil
	}

	// Sort overrides by start time
	sort.Slice(validOverrides, func(i, j int) bool {
		return validOverrides[i].From.Before(validOverrides[j].From)
	})

	// Start with initial schedule and apply overrides one by one
	result := make(InitialScheduleIntervals, len(initialSchedule))
	copy(result, initialSchedule)

	for _, override := range validOverrides {
		newResult := InitialScheduleIntervals{}

		for _, interval := range result {
			// No overlap - keep interval as is
			if override.To.Before(interval.From) || override.To.Equal(interval.From) ||
				override.From.After(interval.To) || override.From.Equal(interval.To) {
				newResult = append(newResult, interval)
				continue
			}

			// There is overlap - split the interval
			// Add pre-override part
			if override.From.After(interval.From) {
				newResult = append(newResult, InitialScheduleInterval{
					User: interval.User,
					From: interval.From,
					To:   override.From,
				})
			}

			// Add override part (clipped to interval bounds)
			overrideStart := override.From
			overrideEnd := override.To
			if overrideStart.Before(interval.From) {
				overrideStart = interval.From
			}
			if overrideEnd.After(interval.To) {
				overrideEnd = interval.To
			}
			newResult = append(newResult, InitialScheduleInterval{
				User: override.User,
				From: overrideStart,
				To:   overrideEnd,
			})

			// Add post-override part
			if override.To.Before(interval.To) {
				newResult = append(newResult, InitialScheduleInterval{
					User: interval.User,
					From: override.To,
					To:   interval.To,
				})
			}
		}

		result = newResult
	}

	return result, nil
}

func CreateFinalSchedule(scheduleWithOverrides InitialScheduleIntervals, fromTime time.Time, untilTime time.Time) (InitialScheduleIntervals, error) {
	if fromTime.After(untilTime) || fromTime.Equal(untilTime) {
		return nil, fmt.Errorf("from time must be before until time")
	}

	fmt.Println(scheduleWithOverrides)

	// Find first and last intervals that overlap with requested time window
	start, end := -1, -1
	for i, interval := range scheduleWithOverrides {
		if start == -1 && interval.To.After(fromTime) {
			start = i
		}
		if interval.From.Before(untilTime) {
			end = i
		}
	}

	// Return empty schedule if no intervals overlap with time window
	if start == -1 || end == -1 {
		return InitialScheduleIntervals{}, nil
	}

	// Extract relevant intervals and adjust boundaries
	result := make(InitialScheduleIntervals, end-start+1)
	copy(result, scheduleWithOverrides[start:end+1])

	// Trim first interval start time to requested window
	if result[0].From.Before(fromTime) {
		result[0].From = fromTime
	}
	// Trim last interval end time to requested window
	if result[len(result)-1].To.After(untilTime) {
		result[len(result)-1].To = untilTime
	}

	return result, nil
}
