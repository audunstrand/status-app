package scheduler

import (
	"time"
)

// ShouldRemind checks if a reminder should be sent for a team.
// Standard schedule: every Monday at 9 AM.
// Returns true if we haven't sent a reminder this week yet.
func ShouldRemind(lastRemindedAt *time.Time, now time.Time) bool {
	// If never reminded, send reminder
	if lastRemindedAt == nil {
		return true
	}

	// Don't send reminder if we already sent one this week
	if isSameWeek(*lastRemindedAt, now) {
		return false
	}

	return true
}

// isSameWeek checks if two times are in the same ISO week
func isSameWeek(t1, t2 time.Time) bool {
	y1, w1 := t1.ISOWeek()
	y2, w2 := t2.ISOWeek()
	return y1 == y2 && w1 == w2
}
