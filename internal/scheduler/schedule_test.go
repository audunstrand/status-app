package scheduler

import (
	"testing"
	"time"
)

func TestMatchesSchedule(t *testing.T) {
	tests := []struct {
		name         string
		pollSchedule string
		weekday      time.Weekday
		expected     bool
	}{
		{
			name:         "daily schedule always matches",
			pollSchedule: "daily",
			weekday:      time.Monday,
			expected:     true,
		},
		{
			name:         "daily schedule matches any day",
			pollSchedule: "daily",
			weekday:      time.Friday,
			expected:     true,
		},
		{
			name:         "weekly schedule matches Monday",
			pollSchedule: "weekly",
			weekday:      time.Monday,
			expected:     true,
		},
		{
			name:         "weekly schedule doesn't match Tuesday",
			pollSchedule: "weekly",
			weekday:      time.Tuesday,
			expected:     false,
		},
		{
			name:         "monday schedule matches Monday",
			pollSchedule: "monday",
			weekday:      time.Monday,
			expected:     true,
		},
		{
			name:         "monday schedule doesn't match Tuesday",
			pollSchedule: "monday",
			weekday:      time.Tuesday,
			expected:     false,
		},
		{
			name:         "friday schedule matches Friday",
			pollSchedule: "friday",
			weekday:      time.Friday,
			expected:     true,
		},
		{
			name:         "case insensitive - FRIDAY",
			pollSchedule: "FRIDAY",
			weekday:      time.Friday,
			expected:     true,
		},
		{
			name:         "whitespace trimmed",
			pollSchedule: "  monday  ",
			weekday:      time.Monday,
			expected:     true,
		},
		{
			name:         "unknown schedule returns false",
			pollSchedule: "invalid",
			weekday:      time.Monday,
			expected:     false,
		},
		{
			name:         "empty schedule returns false",
			pollSchedule: "",
			weekday:      time.Monday,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a time with the specific weekday
			now := time.Date(2025, 12, getDateForWeekday(tt.weekday), 9, 0, 0, 0, time.UTC)
			result := matchesSchedule(tt.pollSchedule, now)
			if result != tt.expected {
				t.Errorf("matchesSchedule(%q, %s) = %v, want %v",
					tt.pollSchedule, tt.weekday, result, tt.expected)
			}
		})
	}
}

func TestShouldRemind(t *testing.T) {
	monday := time.Date(2025, 12, 8, 9, 0, 0, 0, time.UTC)    // Monday
	tuesday := time.Date(2025, 12, 9, 9, 0, 0, 0, time.UTC)   // Tuesday
	mondayLater := time.Date(2025, 12, 8, 15, 0, 0, 0, time.UTC) // Monday, later time

	tests := []struct {
		name           string
		pollSchedule   string
		lastRemindedAt *time.Time
		now            time.Time
		expected       bool
	}{
		{
			name:           "never reminded - monday schedule on monday",
			pollSchedule:   "monday",
			lastRemindedAt: nil,
			now:            monday,
			expected:       true,
		},
		{
			name:           "never reminded - monday schedule on tuesday",
			pollSchedule:   "monday",
			lastRemindedAt: nil,
			now:            tuesday,
			expected:       false,
		},
		{
			name:           "already reminded today - same day",
			pollSchedule:   "daily",
			lastRemindedAt: &monday,
			now:            mondayLater,
			expected:       false,
		},
		{
			name:           "reminded yesterday - daily schedule",
			pollSchedule:   "daily",
			lastRemindedAt: &monday,
			now:            tuesday,
			expected:       true,
		},
		{
			name:           "reminded yesterday - wrong day for schedule",
			pollSchedule:   "monday",
			lastRemindedAt: &monday,
			now:            tuesday,
			expected:       false,
		},
		{
			name:           "empty schedule never reminds",
			pollSchedule:   "",
			lastRemindedAt: nil,
			now:            monday,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRemind(tt.pollSchedule, tt.lastRemindedAt, tt.now)
			if result != tt.expected {
				t.Errorf("ShouldRemind(%q, %v, %v) = %v, want %v",
					tt.pollSchedule, tt.lastRemindedAt, tt.now.Weekday(), result, tt.expected)
			}
		})
	}
}

func TestIsSameDay(t *testing.T) {
	tests := []struct {
		name     string
		t1       time.Time
		t2       time.Time
		expected bool
	}{
		{
			name:     "same day, same time",
			t1:       time.Date(2025, 12, 8, 9, 0, 0, 0, time.UTC),
			t2:       time.Date(2025, 12, 8, 9, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "same day, different time",
			t1:       time.Date(2025, 12, 8, 9, 0, 0, 0, time.UTC),
			t2:       time.Date(2025, 12, 8, 15, 30, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "different day",
			t1:       time.Date(2025, 12, 8, 9, 0, 0, 0, time.UTC),
			t2:       time.Date(2025, 12, 9, 9, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different month",
			t1:       time.Date(2025, 11, 30, 9, 0, 0, 0, time.UTC),
			t2:       time.Date(2025, 12, 1, 9, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different year",
			t1:       time.Date(2024, 12, 31, 9, 0, 0, 0, time.UTC),
			t2:       time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameDay(tt.t1, tt.t2)
			if result != tt.expected {
				t.Errorf("isSameDay(%v, %v) = %v, want %v",
					tt.t1, tt.t2, result, tt.expected)
			}
		})
	}
}

// getDateForWeekday returns a day number in December 2025 that falls on the given weekday
// December 2025: Mon=1,8,15,22,29 Tue=2,9,16,23,30 Wed=3,10,17,24,31 Thu=4,11,18,25 Fri=5,12,19,26 Sat=6,13,20,27 Sun=7,14,21,28
func getDateForWeekday(weekday time.Weekday) int {
	dates := map[time.Weekday]int{
		time.Monday:    8,
		time.Tuesday:   9,
		time.Wednesday: 10,
		time.Thursday:  11,
		time.Friday:    12,
		time.Saturday:  13,
		time.Sunday:    14,
	}
	return dates[weekday]
}
