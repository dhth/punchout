package persistence

import (
	"database/sql"
	"testing"
	"time"

	d "github.com/dhth/punchout/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // sqlite driver
)

func TestQuickSwitchActiveWLInDB(t *testing.T) {
	db := setupTestDB(t)

	// GIVEN
	now := time.Now()
	activeIssueKey := "OLD-ACTIVE"
	newActiveIssueKey := "NEW-ACTIVE"
	beginTS := now.Add(-30 * time.Minute)
	_, err := db.Exec(`
INSERT INTO issue_log (issue_key, begin_ts, active, synced)
VALUES (?, ?, ?, ?);
`, activeIssueKey, beginTS, true, 0)
	require.NoError(t, err, "couldn't insert active worklog")

	// WHEN
	err = QuickSwitchActiveWLInDB(db, activeIssueKey, newActiveIssueKey, now)
	require.NoError(t, err, "quick switching returned an error")

	// THEN
	numActiveIssues, err := getNumActiveIssuesFromDB(db)
	require.NoError(t, err, "couldn't get number of active issues")
	gotNewActive, err := GetActiveIssueFromDB(db)
	require.NoError(t, err, "couldn't get active issue")
	worklogs, err := getWorkLogsForIssueFromDB(db, activeIssueKey)
	require.NoError(t, err, "couldn't get worklog entries for active issue")
	activeWorklog, err := FetchActiveWLFromDB(db)
	require.NoError(t, err, "couldn't get active worklog")

	assert.Equal(t, 1, numActiveIssues, "number of active issues is incorrect")
	assert.Equal(t, newActiveIssueKey, gotNewActive, "new active issue key is incorrect")
	assert.Len(t, worklogs, 1, "work log entries for older issue is incorrect")
	require.NotNil(t, activeWorklog, "active worklog is missing")
	assert.Equal(t, newActiveIssueKey, activeWorklog.IssueKey, "active worklog issue key is incorrect")
}

func TestFetchActiveWLFromDB(t *testing.T) {
	t.Run("returns nil when no worklog is active", func(t *testing.T) {
		db := setupTestDB(t)

		worklog, err := FetchActiveWLFromDB(db)

		require.NoError(t, err)
		assert.Nil(t, worklog)
	})

	t.Run("returns active worklog", func(t *testing.T) {
		db := setupTestDB(t)
		beginTS := time.Now().Truncate(time.Second)
		insertActiveWorklog(t, db, "ACTIVE-ISSUE", beginTS, "comment")

		worklog, err := FetchActiveWLFromDB(db)

		require.NoError(t, err)
		require.NotNil(t, worklog)
		assert.Equal(t, "ACTIVE-ISSUE", worklog.IssueKey)
		assert.True(t, beginTS.Equal(worklog.BeginTS))
		assert.Equal(t, "comment", worklog.Comment)
	})

	t.Run("normalizes NULL comment to empty string", func(t *testing.T) {
		db := setupTestDB(t)
		insertActiveWorklog(t, db, "NULL-COMMENT", time.Now(), nil)

		worklog, err := FetchActiveWLFromDB(db)

		require.NoError(t, err)
		require.NotNil(t, worklog)
		assert.Empty(t, worklog.Comment)
	})
}

func TestCompletedWorklogQueries(t *testing.T) {
	type queryFunc func(*sql.DB, string) ([]d.StoredWorklog, error)

	queries := []struct {
		name   string
		synced bool
		query  queryFunc
	}{
		{
			name: "worklogs for issue",
			query: func(db *sql.DB, issueKey string) ([]d.StoredWorklog, error) {
				return getWorkLogsForIssueFromDB(db, issueKey)
			},
		},
	}

	for _, queryTest := range queries {
		t.Run(queryTest.name, func(t *testing.T) {
			t.Run("normalizes NULL comment to empty string", func(t *testing.T) {
				db := setupTestDB(t)
				issueKey := "NULL-COMMENT"
				insertCompletedWorklog(t, db, issueKey, time.Now(), queryTest.synced)

				worklogs, err := queryTest.query(db, issueKey)

				require.NoError(t, err)
				require.Len(t, worklogs, 1)
				assert.Empty(t, worklogs[0].Comment)
			})

			t.Run("rejects NULL end timestamp", func(t *testing.T) {
				db := setupTestDB(t)
				issueKey := "NULL-END"
				insertCompletedWorklog(t, db, issueKey, nil, queryTest.synced)

				_, err := queryTest.query(db, issueKey)

				require.Error(t, err)
			})
		})
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoErrorf(t, err, "error opening DB: %v", err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	err = InitDB(db)
	require.NoErrorf(t, err, "error initializing DB: %v", err)

	return db
}

func insertActiveWorklog(t *testing.T, db *sql.DB, issueKey string, beginTS time.Time, comment any) {
	t.Helper()

	_, err := db.Exec(`
INSERT INTO issue_log (issue_key, begin_ts, COMMENT, active, synced)
VALUES (?, ?, ?, true, false);
`, issueKey, beginTS, comment)
	require.NoError(t, err)
}

func insertCompletedWorklog(t *testing.T, db *sql.DB, issueKey string, endTS any, synced bool) {
	t.Helper()
	_, err := db.Exec(`
INSERT INTO issue_log (issue_key, begin_ts, end_ts, COMMENT, active, synced)
VALUES (?, ?, ?, NULL, false, ?);
`, issueKey, time.Now().Add(-time.Hour), endTS, synced)
	require.NoError(t, err)
}
