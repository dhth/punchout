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
