package main

import "time"

type Override struct {
	User string    `json:"user"`
	From time.Time `json:"start_at"`
	To   time.Time `json:"end_at"`
}

type Overrides []Override

type SchedulePlan struct {
	Users                []string `json:"users"`
	HandoverStartDate    string   `json:"handover_start_at"`
	HandoverIntervalDays int      `json:"handover_interval_days"`
}

type InitialScheduleInterval struct {
	User string    `json:"user"`
	From time.Time `json:"start_at"`
	To   time.Time `json:"end_at"`
}

type InitialScheduleIntervals []InitialScheduleInterval
