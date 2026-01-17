// Package models implements the models for this project
package models

import "time"

// EventType represents the type of task event
type EventType string

const (
	// Time tracking events
	EventTypeTimeStart EventType = "time_start"
	EventTypeTimeStop  EventType = "time_stop"

	// Future event types (for reference)
	// EventTypeComment   EventType = "comment"
)

// TaskEvent represents a single event on a task (time tracking, comments, etc.)
type TaskEvent struct {
	ID        int64          `json:"id"`
	TaskID    int64          `json:"task_id"`
	UserID    int64          `json:"user_id"`
	EventType EventType      `json:"event_type"`
	EventData map[string]any `json:"event_data"`
	EventTime time.Time      `json:"event_time"`
	CreatedAt time.Time      `json:"created_at"`
}

// TaskEventView represents a task event for display
type TaskEventView struct {
	ID        int64
	TaskID    int64
	User      UserView
	EventType EventType
	EventData map[string]any
	EventTime time.Time
}

// ToView converts a TaskEvent to a TaskEventView
func (e *TaskEvent) ToView(user *User) *TaskEventView {
	return &TaskEventView{
		ID:        e.ID,
		TaskID:    e.TaskID,
		User:      user.ToViewModel(),
		EventType: e.EventType,
		EventData: e.EventData,
		EventTime: e.EventTime,
	}
}

// IsTimeEvent returns true if this is a time tracking event
func (e *TaskEvent) IsTimeEvent() bool {
	return e.EventType == EventTypeTimeStart || e.EventType == EventTypeTimeStop
}

// FormatDuration formats a duration in seconds to a human-readable string
func FormatDuration(seconds int64) string {
	if seconds < 60 {
		return "< 1m"
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if hours > 0 {
		return formatWithUnits(hours, "h", minutes, "m")
	}
	return formatWithUnits(minutes, "m", 0, "")
}

func formatWithUnits(val1 int64, unit1 string, val2 int64, unit2 string) string {
	if val2 > 0 && unit2 != "" {
		return intToStr(val1) + unit1 + " " + intToStr(val2) + unit2
	}
	return intToStr(val1) + unit1
}

func intToStr(n int64) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
