package ui

import "time"

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
)

func getShiftedTime(ts time.Time, direction timeShiftDirection, duration timeShiftDuration) time.Time {
	var d time.Duration

	switch duration {
	case shiftMinute:
		d = time.Minute
	case shiftFiveMinutes:
		d = time.Minute * 5
	case shiftHour:
		d = time.Hour
	}

	if direction == shiftBackward {
		d = -1 * d
	}
	return ts.Add(d)
}
