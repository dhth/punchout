package ui

import "time"

type realTimeProvider struct{}

func (realTimeProvider) Now() time.Time {
	return time.Now()
}
