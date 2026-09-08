package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/dhth/punchout/internal/utils"
)

type timeShiftDirection uint8

const (
	shiftForward timeShiftDirection = iota
	shiftBackward
)

type timeShiftDuration uint8

const (
	shiftMinute timeShiftDuration = iota
	shiftFiveMinutes
	shiftHour
	shiftDay
)

func getShiftedTime(
	ts time.Time,
	direction timeShiftDirection,
	duration timeShiftDuration,
) time.Time {
	var d time.Duration

	switch duration {
	case shiftMinute:
		d = time.Minute
	case shiftFiveMinutes:
		d = time.Minute * 5
	case shiftHour:
		d = time.Hour
	case shiftDay:
		d = time.Hour * 24
	}

	if direction == shiftBackward {
		d = -1 * d
	}
	return ts.Add(d)
}

func getDurationValidityContext(beginStr, endStr string) (string, wlFormValidity) {
	if strings.TrimSpace(beginStr) == "" {
		return "Begin time is empty", wlSubmitErr
	}

	if strings.TrimSpace(endStr) == "" {
		return "End time is empty", wlSubmitErr
	}

	beginTS, err := time.ParseInLocation(timeFormat, beginStr, time.Local)
	if err != nil {
		return "Begin time is invalid", wlSubmitErr
	}

	endTS, err := time.ParseInLocation(timeFormat, endStr, time.Local)
	if err != nil {
		return "End time is invalid", wlSubmitErr
	}

	dur := endTS.Sub(beginTS)

	if dur == 0 {
		return "You're recording no time, change begin and/or end time", wlSubmitErr
	}

	if dur < 0 {
		return "End time is before start time", wlSubmitErr
	}

	totalSeconds := int(dur.Seconds())

	humanized := utils.HumanizeDuration(totalSeconds)
	msg := fmt.Sprintf("You're recording %s", humanized)
	if totalSeconds > wLWarningThresholdSecs {
		return msg, wlSubmitWarn
	}

	return msg, wlSubmitOk
}
