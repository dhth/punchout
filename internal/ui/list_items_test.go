package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFormatWorklogTimeRange(t *testing.T) {
	vienna, err := time.LoadLocation("Europe/Vienna")
	require.NoError(t, err)
	now := time.Date(2026, time.September, 10, 12, 0, 0, 0, vienna)

	tests := []struct {
		name     string
		now      time.Time
		start    time.Time
		end      time.Time
		expected string
	}{
		{
			name:     "today",
			now:      now,
			start:    time.Date(2026, time.September, 10, 9, 0, 0, 0, vienna),
			end:      time.Date(2026, time.September, 10, 10, 0, 0, 0, vienna),
			expected: "09:00 – 10:00",
		},
		{
			name:     "earlier this week",
			now:      now,
			start:    time.Date(2026, time.September, 7, 9, 0, 0, 0, vienna),
			end:      time.Date(2026, time.September, 7, 10, 0, 0, 0, vienna),
			expected: "Mon, 09:00 – 10:00",
		},
		{
			name:     "later this week",
			now:      now,
			start:    time.Date(2026, time.September, 12, 9, 0, 0, 0, vienna),
			end:      time.Date(2026, time.September, 12, 10, 0, 0, 0, vienna),
			expected: "Sat, 09:00 – 10:00",
		},
		{
			name:     "Sunday is in current week",
			now:      now,
			start:    time.Date(2026, time.September, 13, 9, 0, 0, 0, vienna),
			end:      time.Date(2026, time.September, 13, 10, 0, 0, 0, vienna),
			expected: "Sun, 09:00 – 10:00",
		},
		{
			name:     "next Monday is outside current week",
			now:      now,
			start:    time.Date(2026, time.September, 14, 9, 0, 0, 0, vienna),
			end:      time.Date(2026, time.September, 14, 10, 0, 0, 0, vienna),
			expected: "Sep 14, 09:00 – 10:00",
		},
		{
			name:     "outside current week",
			now:      now,
			start:    time.Date(2026, time.August, 29, 9, 0, 0, 0, vienna),
			end:      time.Date(2026, time.August, 29, 10, 0, 0, 0, vienna),
			expected: "Aug 29, 09:00 – 10:00",
		},
		{
			name:     "spans days in current week",
			now:      now,
			start:    time.Date(2026, time.September, 8, 23, 0, 0, 0, vienna),
			end:      time.Date(2026, time.September, 9, 1, 0, 0, 0, vienna),
			expected: "Tue, 23:00 – Wed, 01:00",
		},
		{
			name:     "spans current week boundary",
			now:      now,
			start:    time.Date(2026, time.September, 6, 23, 0, 0, 0, vienna),
			end:      time.Date(2026, time.September, 7, 1, 0, 0, 0, vienna),
			expected: "Sep 6, 23:00 – Mon, 01:00",
		},
		{
			name:     "spans days outside current week",
			now:      now,
			start:    time.Date(2026, time.August, 29, 23, 0, 0, 0, vienna),
			end:      time.Date(2026, time.August, 30, 1, 0, 0, 0, vienna),
			expected: "Aug 29, 23:00 – Aug 30, 01:00",
		},
		{
			name:     "spans years without displaying year",
			now:      time.Date(2026, time.January, 8, 12, 0, 0, 0, vienna),
			start:    time.Date(2025, time.December, 31, 23, 0, 0, 0, vienna),
			end:      time.Date(2026, time.January, 1, 1, 0, 0, 0, vienna),
			expected: "Dec 31, 23:00 – Jan 1, 01:00",
		},
		{
			name:     "current week spans calendar years",
			now:      time.Date(2026, time.January, 1, 12, 0, 0, 0, vienna),
			start:    time.Date(2025, time.December, 31, 9, 0, 0, 0, vienna),
			end:      time.Date(2025, time.December, 31, 10, 0, 0, 0, vienna),
			expected: "Wed, 09:00 – 10:00",
		},
		{
			name:     "uses now location",
			now:      now,
			start:    time.Date(2026, time.September, 9, 22, 30, 0, 0, time.UTC),
			end:      time.Date(2026, time.September, 9, 23, 30, 0, 0, time.UTC),
			expected: "00:30 – 01:30",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, formatWorklogTimeRange(test.now, test.start, test.end))
		})
	}
}
