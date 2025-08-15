package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/stretchr/testify/assert"
)

func TestGetDurationValidityContext(t *testing.T) {
	testCases := []struct {
		name             string
		beginTS          string
		endTS            string
		expectedCtx      string
		expectedValidity wlFormValidity
	}{
		// success
		{
			name:             "a simple case",
			beginTS:          "2025/08/08 00:40",
			endTS:            "2025/08/08 00:48",
			expectedCtx:      "You're recording 8m",
			expectedValidity: wlSubmitOk,
		},
		{
			name:             "exact hour",
			beginTS:          "2025/08/08 00:00",
			endTS:            "2025/08/08 01:00",
			expectedCtx:      "You're recording 1h",
			expectedValidity: wlSubmitOk,
		},
		{
			name:             "hours and minutes",
			beginTS:          "2025/08/08 00:00",
			endTS:            "2025/08/08 02:30",
			expectedCtx:      "You're recording 2h 30m",
			expectedValidity: wlSubmitOk,
		},
		{
			name:             "across day boundary",
			beginTS:          "2025/08/08 23:30",
			endTS:            "2025/08/09 00:15",
			expectedCtx:      "You're recording 45m",
			expectedValidity: wlSubmitOk,
		},
		{
			name:             "exactly at 8h threshold",
			beginTS:          "2025/08/08 00:00",
			endTS:            "2025/08/08 08:00",
			expectedCtx:      "You're recording 8h",
			expectedValidity: wlSubmitOk,
		},
		{
			name:             "> 8h threshold",
			beginTS:          "2025/08/08 00:00",
			endTS:            "2025/08/08 08:01",
			expectedCtx:      "You're recording 8h 1m",
			expectedValidity: wlSubmitWarn,
		},
		// failures
		{
			name:             "empty begin",
			beginTS:          "",
			endTS:            "2025/08/08 00:10",
			expectedCtx:      "Begin time is empty",
			expectedValidity: wlSubmitErr,
		},
		{
			name:             "empty end",
			beginTS:          "2025/08/08 00:10",
			endTS:            "",
			expectedCtx:      "End time is empty",
			expectedValidity: wlSubmitErr,
		},
		{
			name:             "invalid begin ts",
			beginTS:          "2025-08-08 00:10",
			endTS:            "2025/08/08 00:20",
			expectedCtx:      "Begin time is invalid",
			expectedValidity: wlSubmitErr,
		},
		{
			name:             "invalid end format",
			beginTS:          "2025/08/08 00:10",
			endTS:            "08-08-2025 00:20",
			expectedCtx:      "End time is invalid",
			expectedValidity: wlSubmitErr,
		},
		{
			name:             "end before start",
			beginTS:          "2025/08/08 01:00",
			endTS:            "2025/08/08 00:59",
			expectedCtx:      "End time is before start time",
			expectedValidity: wlSubmitErr,
		},
		{
			name:             "zero duration",
			beginTS:          "2025/08/08 00:00",
			endTS:            "2025/08/08 00:00",
			expectedCtx:      "You're recording no time, change begin and/or end time",
			expectedValidity: wlSubmitErr,
		},
		{
			name:             "exactly one minute",
			beginTS:          "2025/08/08 00:00",
			endTS:            "2025/08/08 00:01",
			expectedCtx:      "You're recording 1m",
			expectedValidity: wlSubmitOk,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotCtx, gotOk := getDurationValidityContext(tt.beginTS, tt.endTS)

			assert.Equal(t, tt.expectedCtx, gotCtx)
			assert.Equal(t, tt.expectedValidity, gotOk)
		})
	}
}

func TestGetCmdToSaveActiveWLWithoutComment(t *testing.T) {
	tests := []struct {
		name            string
		fallbackComment *string
		beginTS         string
		endTS           string
		expectedMessage string
		shouldReturnCmd bool
	}{
		{
			name:            "no fallback comment configured",
			fallbackComment: nil,
			beginTS:         "2025/08/08 00:00",
			endTS:           "2025/08/08 01:00",
			expectedMessage: "Cannot save without comment: no fallback comment configured",
			shouldReturnCmd: false,
		},
		{
			name:            "fallback comment configured, valid duration",
			fallbackComment: stringPtr("Fallback comment"),
			beginTS:         "2025/08/08 00:00",
			endTS:           "2025/08/08 01:00",
			expectedMessage: "",
			shouldReturnCmd: true,
		},
		{
			name:            "fallback comment configured, invalid duration",
			fallbackComment: stringPtr("Fallback comment"),
			beginTS:         "2025/08/08 00:00",
			endTS:           "2025/08/08 00:00",
			expectedMessage: "Cannot save: duration is less than 1 minute",
			shouldReturnCmd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{
				fallbackComment: tt.fallbackComment,
				trackingInputs:  make([]textinput.Model, 3),
			}

			// Initialize text inputs
			for i := range m.trackingInputs {
				m.trackingInputs[i] = textinput.New()
			}

			m.trackingInputs[entryBeginTS].SetValue(tt.beginTS)
			m.trackingInputs[entryEndTS].SetValue(tt.endTS)

			cmd := m.getCmdToSaveActiveWLWithoutComment()

			if tt.shouldReturnCmd {
				assert.NotNil(t, cmd, "Expected command to be returned")
			} else {
				assert.Nil(t, cmd, "Expected no command to be returned")
			}

			if tt.expectedMessage != "" {
				assert.Equal(t, tt.expectedMessage, m.message)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
