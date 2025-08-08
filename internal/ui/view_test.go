package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDurationContext(t *testing.T) {
	testCases := []struct {
		name        string
		beginTS     string
		endTS       string
		expectedCtx string
		expectedOk  bool
	}{
		// success
		{
			name:        "a simple case",
			beginTS:     "2025/08/08 00:40",
			endTS:       "2025/08/08 00:48",
			expectedCtx: "You're recording 8m",
			expectedOk:  true,
		},
		{
			name:        "exact hour",
			beginTS:     "2025/08/08 00:00",
			endTS:       "2025/08/08 01:00",
			expectedCtx: "You're recording 1h",
			expectedOk:  true,
		},
		{
			name:        "hours and minutes",
			beginTS:     "2025/08/08 00:00",
			endTS:       "2025/08/08 02:30",
			expectedCtx: "You're recording 2h 30m",
			expectedOk:  true,
		},
		{
			name:        "across day boundary",
			beginTS:     "2025/08/08 23:30",
			endTS:       "2025/08/09 00:15",
			expectedCtx: "You're recording 45m",
			expectedOk:  true,
		},
		{
			name:        "exactly at 8h threshold",
			beginTS:     "2025/08/08 00:00",
			endTS:       "2025/08/08 08:00",
			expectedCtx: "You're recording 8h",
			expectedOk:  true,
		},
		{
			name:        "> 8h threshold (warning style)",
			beginTS:     "2025/08/08 00:00",
			endTS:       "2025/08/08 08:01",
			expectedCtx: "You're recording 8h 1m",
			expectedOk:  true,
		},
		// failures
		{
			name:        "empty begin",
			beginTS:     "",
			endTS:       "2025/08/08 00:10",
			expectedCtx: "",
			expectedOk:  false,
		},
		{
			name:        "empty end",
			beginTS:     "2025/08/08 00:10",
			endTS:       "",
			expectedCtx: "",
			expectedOk:  false,
		},
		{
			name:        "invalid begin format",
			beginTS:     "2025-08-08 00:10",
			endTS:       "2025/08/08 00:20",
			expectedCtx: "",
			expectedOk:  false,
		},
		{
			name:        "invalid end format",
			beginTS:     "2025/08/08 00:10",
			endTS:       "08-08-2025 00:20",
			expectedCtx: "",
			expectedOk:  false,
		},
		{
			name:        "end before start",
			beginTS:     "2025/08/08 01:00",
			endTS:       "2025/08/08 00:59",
			expectedCtx: "End time is before start time",
			expectedOk:  false,
		},
		{
			name:        "zero duration",
			beginTS:     "2025/08/08 00:00",
			endTS:       "2025/08/08 00:00",
			expectedCtx: "You're recording no time",
			expectedOk:  false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotCtx, gotOk := getDurationContext(tt.beginTS, tt.endTS)

			assert.Equal(t, tt.expectedCtx, gotCtx)
			assert.Equal(t, tt.expectedOk, gotOk)
		})
	}
}
