package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanizeDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{name: "zero", seconds: 0, expected: "0s"},
		{name: "less than a minute", seconds: 42, expected: "42s"},
		{name: "last second before minute", seconds: 59, expected: "59s"},
		{name: "exactly one minute", seconds: 60, expected: "1m"},
		{name: "truncate leftover seconds", seconds: 119, expected: "1m"},
		{name: "multiple minutes", seconds: 120, expected: "2m"},
		{name: "last second before hour", seconds: 3599, expected: "59m"},
		{name: "exactly one hour", seconds: 3600, expected: "1h"},
		{name: "hours and minutes", seconds: 3660, expected: "1h 1m"},
		{name: "truncate seconds after hour", seconds: 3661, expected: "1h 1m"},
		{name: "exact multiple of hours", seconds: 7200, expected: "2h"},
		{name: "exactly one day", seconds: 86400, expected: "24h"},
		{name: "days and hours", seconds: 90000, expected: "1d 1h"},
		{name: "truncate minutes after days and hours", seconds: 93780, expected: "1d 2h"},
		{name: "truncate minutes after days", seconds: 172980, expected: "2d"},
		{name: "exact multiple of days", seconds: 172800, expected: "2d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HumanizeDuration(tt.seconds))
		})
	}
}
