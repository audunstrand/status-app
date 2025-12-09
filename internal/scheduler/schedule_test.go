package scheduler

import (
	"testing"
	"time"
)

func TestShouldRemind(t *testing.T) {
	// Week 1: Monday Dec 8, 2025
	week1Monday := time.Date(2025, 12, 8, 9, 0, 0, 0, time.UTC)
	// Same week: Tuesday Dec 9, 2025
	week1Tuesday := time.Date(2025, 12, 9, 9, 0, 0, 0, time.UTC)
	// Next week: Monday Dec 15, 2025
	week2Monday := time.Date(2025, 12, 15, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		lastRemindedAt *time.Time
		now            time.Time
		expected       bool
	}{
		{
			name:           "never reminded - should send",
			lastRemindedAt: nil,
			now:            week1Monday,
			expected:       true,
		},
		{
			name:           "reminded earlier this week - should not send",
			lastRemindedAt: &week1Monday,
			now:            week1Tuesday,
			expected:       false,
		},
		{
			name:           "reminded last week - should send",
			lastRemindedAt: &week1Monday,
			now:            week2Monday,
			expected:       true,
		},
		{
			name:           "reminded same day - should not send",
			lastRemindedAt: &week1Monday,
			now:            week1Monday,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRemind(tt.lastRemindedAt, tt.now)
			if result != tt.expected {
				t.Errorf("ShouldRemind(%v, %v) = %v, want %v",
					tt.lastRemindedAt, tt.now, result, tt.expected)
			}
		})
	}
}

func TestIsSameWeek(t *testing.T) {
	tests := []struct {
		name     string
		t1       time.Time
		t2       time.Time
		expected bool
	}{
		{
			name:     "same week - Monday and Tuesday",
			t1:       time.Date(2025, 12, 8, 9, 0, 0, 0, time.UTC),
			t2:       time.Date(2025, 12, 9, 15, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "same week - same day different time",
			t1:       time.Date(2025, 12, 8, 9, 0, 0, 0, time.UTC),
			t2:       time.Date(2025, 12, 8, 15, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "different week",
			t1:       time.Date(2025, 12, 8, 9, 0, 0, 0, time.UTC),
			t2:       time.Date(2025, 12, 15, 9, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different year",
			t1:       time.Date(2024, 12, 30, 9, 0, 0, 0, time.UTC),
			t2:       time.Date(2025, 1, 6, 9, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameWeek(tt.t1, tt.t2)
			if result != tt.expected {
				t.Errorf("isSameWeek(%v, %v) = %v, want %v",
					tt.t1, tt.t2, result, tt.expected)
			}
		})
	}
}
