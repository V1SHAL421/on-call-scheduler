// Package test contains comprehensive test suite for the on-call scheduler
// Tests cover JSON parsing, schedule generation, override application, and time window filtering
package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"on-call-scheduler/src"
)

// TestParseFieldToStruct tests JSON parsing functionality for schedule plans and overrides
func TestParseFieldToStruct(t *testing.T) {
	// Test parsing valid schedule JSON with users and handover interval
	t.Run("parse schedule", func(t *testing.T) {
		json := []byte(`{"users":["alice","bob"],"handover_interval_days":7}`)
		result, err := src.ParseFieldToStruct(json, "schedule")

		require.NoError(t, err)
		schedule, ok := result.(src.SchedulePlan)
		require.True(t, ok)
		assert.Equal(t, []string{"alice", "bob"}, schedule.Users)
		assert.Equal(t, 7, schedule.HandoverIntervalDays)
	})

	// Test parsing valid overrides JSON array with user and time range
	t.Run("parse overrides", func(t *testing.T) {
		json := []byte(`[{"user":"charlie","start_at":"2025-11-10T17:00:00Z","end_at":"2025-11-10T22:00:00Z"}]`)
		result, err := src.ParseFieldToStruct(json, "overrides")

		require.NoError(t, err)
		overrides, ok := result.(src.Overrides)
		require.True(t, ok)
		assert.Len(t, overrides, 1)
		assert.Equal(t, "charlie", overrides[0].User)
	})

	// Test error handling for empty JSON input
	t.Run("empty field", func(t *testing.T) {
		_, err := src.ParseFieldToStruct([]byte{}, "schedule")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	// Test error handling for unsupported field types
	t.Run("unknown field", func(t *testing.T) {
		json := []byte(`{}`)
		_, err := src.ParseFieldToStruct(json, "unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown field name")
	})
}

// TestCreateInitialSchedule tests rotating schedule generation based on user list and handover intervals
func TestCreateInitialSchedule(t *testing.T) {
	// Setup: 3 users with 7-day handover intervals over 14-day period
	plan := src.SchedulePlan{
		Users:                []string{"alice", "bob", "charlie"},
		HandoverIntervalDays: 7,
	}
	from := time.Date(2025, 11, 7, 17, 0, 0, 0, time.UTC)
	until := time.Date(2025, 11, 21, 17, 0, 0, 0, time.UTC)

	// Test valid schedule generation: should create 2 intervals (alice: days 1-7, bob: days 8-14)
	t.Run("valid schedule", func(t *testing.T) {
		schedule, err := src.CreateInitialSchedule(plan, from, until)

		require.NoError(t, err)
		assert.Len(t, schedule, 2) // 14 days / 7 days per interval = 2 intervals
		assert.Equal(t, "alice", schedule[0].User)
		assert.Equal(t, "bob", schedule[1].User)
		assert.Equal(t, from, schedule[0].From)
		assert.Equal(t, from.AddDate(0, 0, 7), schedule[0].To)
	})

	// Test validation: handover interval must be positive
	t.Run("invalid handover interval", func(t *testing.T) {
		invalidPlan := plan
		invalidPlan.HandoverIntervalDays = 0
		_, err := src.CreateInitialSchedule(invalidPlan, from, until)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handover interval days must be greater than 0")
	})

	// Test validation: user list cannot be empty
	t.Run("empty users", func(t *testing.T) {
		invalidPlan := plan
		invalidPlan.Users = []string{}
		_, err := src.CreateInitialSchedule(invalidPlan, from, until)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "users must be defined")
	})

	// Test validation: from time must be before until time
	t.Run("invalid time range", func(t *testing.T) {
		_, err := src.CreateInitialSchedule(plan, until, from)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "from time must be before end time")
	})
}

// TestAddOverridesToSchedule tests override application logic and edge cases
func TestAddOverridesToSchedule(t *testing.T) {
	baseTime := time.Date(2025, 11, 7, 17, 0, 0, 0, time.UTC)
	// Setup: alice (days 1-7), bob (days 8-14)
	initialSchedule := src.InitialScheduleIntervals{
		{User: "alice", From: baseTime, To: baseTime.AddDate(0, 0, 7)},
		{User: "bob", From: baseTime.AddDate(0, 0, 7), To: baseTime.AddDate(0, 0, 14)},
	}

	// Test no overrides: should return original schedule unchanged
	t.Run("no overrides", func(t *testing.T) {
		result, err := src.AddOverridesToSchedule(initialSchedule, src.Overrides{})
		require.NoError(t, err)
		assert.Equal(t, initialSchedule, result)
	})

	// Test simple override: charlie takes over alice's shift for days 2-3
	// Expected result: alice (day 1), charlie (days 2-3), alice (days 4-7), bob (days 8-14)
	t.Run("simple override", func(t *testing.T) {
		overrides := src.Overrides{
			{User: "charlie", From: baseTime.Add(24 * time.Hour), To: baseTime.Add(48 * time.Hour)},
		}
		result, err := src.AddOverridesToSchedule(initialSchedule, overrides)

		require.NoError(t, err)
		assert.Len(t, result, 4) // Alice split into 2 parts + charlie + bob unchanged
		assert.Equal(t, "alice", result[0].User)   // Pre-override alice
		assert.Equal(t, "charlie", result[1].User) // Override period
		assert.Equal(t, "alice", result[2].User)   // Post-override alice
		assert.Equal(t, "bob", result[3].User)     // Unchanged bob
	})

	// Test validation: cannot apply overrides to empty schedule
	t.Run("empty initial schedule", func(t *testing.T) {
		overrides := src.Overrides{{User: "charlie", From: baseTime, To: baseTime.Add(time.Hour)}}
		_, err := src.AddOverridesToSchedule(src.InitialScheduleIntervals{}, overrides)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initial schedule must be defined")
	})

	// Test zero-duration override filtering: overrides with same start/end time should be ignored
	t.Run("zero duration override filtered", func(t *testing.T) {
		overrides := src.Overrides{
			{User: "charlie", From: baseTime.Add(24 * time.Hour), To: baseTime.Add(24 * time.Hour)},
		}
		result, err := src.AddOverridesToSchedule(initialSchedule, overrides)

		require.NoError(t, err)
		assert.Equal(t, initialSchedule, result) // Should be unchanged
	})
}

// TestCreateFinalSchedule tests time window filtering and boundary adjustment
func TestCreateFinalSchedule(t *testing.T) {
	baseTime := time.Date(2025, 11, 7, 17, 0, 0, 0, time.UTC)
	// Setup: schedule that extends beyond requested time window
	schedule := src.InitialScheduleIntervals{
		{User: "alice", From: baseTime.AddDate(0, 0, -1), To: baseTime.AddDate(0, 0, 3)},   // Starts before window
		{User: "bob", From: baseTime.AddDate(0, 0, 3), To: baseTime.AddDate(0, 0, 10)},    // Within window
		{User: "charlie", From: baseTime.AddDate(0, 0, 10), To: baseTime.AddDate(0, 0, 17)}, // Extends beyond window
	}

	// Test time window trimming: intervals should be clipped to requested time range
	t.Run("trim to window", func(t *testing.T) {
		from := baseTime
		until := baseTime.AddDate(0, 0, 14)

		result, err := src.CreateFinalSchedule(schedule, from, until)

		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, from, result[0].From) // Alice's start trimmed to window start
		assert.Equal(t, until, result[2].To)  // Charlie's end trimmed to window end
	})

	// Test no overlap: time window completely outside schedule should return empty result
	t.Run("no overlap", func(t *testing.T) {
		from := baseTime.AddDate(0, 0, 20) // Window starts after schedule ends
		until := baseTime.AddDate(0, 0, 25)

		result, err := src.CreateFinalSchedule(schedule, from, until)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	// Test validation: from time must be before until time
	t.Run("invalid time range", func(t *testing.T) {
		from := baseTime.AddDate(0, 0, 10)
		until := baseTime // Until is before from

		_, err := src.CreateFinalSchedule(schedule, from, until)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "from time must be before until time")
	})

	// Test zero-duration filtering: intervals that become zero-duration after trimming should be removed
	t.Run("filter zero duration after trim", func(t *testing.T) {
		// Alice's interval ends exactly at window start, should be filtered out
		scheduleWithEdgeCase := src.InitialScheduleIntervals{
			{User: "alice", From: baseTime.AddDate(0, 0, -1), To: baseTime},
			{User: "bob", From: baseTime, To: baseTime.AddDate(0, 0, 7)},
		}

		result, err := src.CreateFinalSchedule(scheduleWithEdgeCase, baseTime, baseTime.AddDate(0, 0, 7))

		require.NoError(t, err)
		assert.Len(t, result, 1) // Alice filtered out, only bob remains
		assert.Equal(t, "bob", result[0].User)
	})
}
