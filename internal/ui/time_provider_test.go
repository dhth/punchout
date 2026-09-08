package ui

import "time"

type testTimeProvider struct {
	fixedTime time.Time
}

func (t testTimeProvider) Now() time.Time {
	return t.fixedTime
}
