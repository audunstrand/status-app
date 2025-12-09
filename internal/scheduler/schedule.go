package scheduler

import (
	"strings"
	"time"
)

// ShouldRemind checks if a reminder should be sent for a team based on their schedule
// and when they were last reminded
func ShouldRemind(pollSchedule string, lastRemindedAt *time.Time, now time.Time) bool {
	if pollSchedule == "" {
		return false
	}

	// If never reminded, check if schedule matches today
	if lastRemindedAt == nil {
		return matchesSchedule(pollSchedule, now)
	}

	// Don't send reminder if we already sent one today
	if isSameDay(*lastRemindedAt, now) {
		return false
	}

	// Check if schedule matches today
	return matchesSchedule(pollSchedule, now)
}

// isSameDay checks if two times are on the same calendar day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// matchesSchedule checks if the current day/time matches the poll schedule
func matchesSchedule(pollSchedule string, now time.Time) bool {
	schedule := strings.ToLower(strings.TrimSpace(pollSchedule))
	weekday := now.Weekday()

	switch schedule {
	case "daily":
		return true
	case "weekly":
		// Weekly defaults to Monday
		return weekday == time.Monday
	case "monday":
		return weekday == time.Monday
	case "tuesday":
		return weekday == time.Tuesday
	case "wednesday":
		return weekday == time.Wednesday
	case "thursday":
		return weekday == time.Thursday
	case "friday":
		return weekday == time.Friday
	case "saturday":
		return weekday == time.Saturday
	case "sunday":
		return weekday == time.Sunday
	default:
		// Unknown schedule format, don't send reminder
		return false
	}
}
