package ui

import (
	"context"
	"testing"
	"time"

	"github.com/dhth/punchout/internal/issuecache"
	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/stretchr/testify/require"
)

var referenceTime = time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC)

type testTimeProvider struct {
	fixedTime time.Time
}

func (t testTimeProvider) Now() time.Time {
	return t.fixedTime
}

func newTestModel(t *testing.T) Model {
	t.Helper()

	thm, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)
	m := InitialModel(
		context.Background(),
		nil,
		nil,
		issuecache.Store{},
		Options{},
		thm,
		testTimeProvider{fixedTime: referenceTime},
		false,
	)
	m.showHelpIndicator = false
	m.issuesFetched = true

	return m
}
