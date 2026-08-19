package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestStoredWorklogJSONFlattensWorklogFields(t *testing.T) {
	beginTS := time.Date(2026, time.August, 14, 9, 30, 0, 0, time.UTC)
	worklog := StoredWorklog{
		Worklog: Worklog{
			IssueKey: "PROJ-123",
			BeginTS:  beginTS,
			EndTS:    beginTS.Add(90 * time.Minute),
			Comment:  "",
		},
		ID:     42,
		Synced: false,
	}

	got, err := json.MarshalIndent(worklog, "", "  ")
	require.NoError(t, err)

	snaps.MatchStandaloneSnapshot(t, string(got))
}
