package persistence

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // sqlite driver
)

func TestQueries(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoErrorf(t, err, "error opening DB: %v", err)

	err = InitDB(db)
	require.NoErrorf(t, err, "error initializing DB: %v", err)

	t.Run("TestQuickSwitchActiveIssue", func(t *testing.T) {
		t.Cleanup(func() { cleanupDB(t, db) })

		// GIVEN
		now := time.Now()
		activeIssueKey := "OLD-ACTIVE"
		newActiveIssueKey := "NEW-ACTIVE"
		beginTS := now.Add(time.Minute * -1 * time.Duration(30))
		_, err := db.Exec(`
INSERT INTO issue_log (issue_key, begin_ts, active, synced)
VALUES (?, ?, ?, ?);
`, activeIssueKey, beginTS, true, 0)
		require.NoError(t, err, "couldn't insert active worklog")

		// WHEN
		err = QuickSwitchActiveIssue(db, activeIssueKey, newActiveIssueKey, now)
		require.NoError(t, err, "quick switching returned an error")

		// THEN
		numActiveIssues, err := getNumActiveIssues(db)
		require.NoError(t, err, "couldn't get number of active issues")
		gotNewActive, err := GetActiveIssue(db)
		require.NoError(t, err, "couldn't get active issue")
		wl1, err := getWorklogEntriesForIssue(db, activeIssueKey)
		require.NoError(t, err, "couldn't get worklog entries for active issue")
		wl2, err := getWorklogEntriesForIssue(db, newActiveIssueKey)
		require.NoError(t, err, "couldn't get worklog entries for new issue")

		assert.Equal(t, 1, numActiveIssues, "number of active issues is incorrect")
		assert.Equal(t, newActiveIssueKey, gotNewActive, "new active issue key is incorrect")
		assert.Len(t, wl1, 1, "work log entries for older issue is incorrect")
		assert.Len(t, wl2, 1, "work log entries for new issue is incorrect")
	})
}

func cleanupDB(t *testing.T, testDB *sql.DB) {
	t.Helper()

	var err error
	for _, tbl := range []string{"issue_log"} {
		_, err = testDB.Exec(fmt.Sprintf("DELETE FROM %s", tbl))
		require.NoErrorf(t, err, "failed to clean up table %q: %v", tbl, err)

		_, err := testDB.Exec("DELETE FROM sqlite_sequence WHERE name=?;", tbl)
		require.NoErrorf(t, err, "failed to reset auto increment for table %q: %v", tbl, err)
	}
}
