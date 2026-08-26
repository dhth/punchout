package persistence

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // sqlite driver
)

type storedWorklogRow struct {
	id       int
	issueKey string
	beginTS  time.Time
	endTS    sql.NullTime
	comment  sql.NullString
	active   bool
	synced   bool
}

type testWorklogRow struct {
	issueKey string
	beginTS  time.Time
	endTS    *time.Time
	comment  *string
	active   bool
	synced   bool
}

func setupTestStore(t *testing.T) *SQLiteStore {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	require.NoError(t, InitDB(db))

	store := &SQLiteStore{db: db}
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	return store
}

func fetchAllWorklogRows(t *testing.T, store *SQLiteStore) []storedWorklogRow {
	t.Helper()

	rows, err := store.db.Query(`
SELECT
    ID,
    issue_key,
    begin_ts,
    end_ts,
    comment,
    active,
    synced
FROM
    issue_log;
`)
	require.NoError(t, err)
	defer rows.Close()

	worklogs := make([]storedWorklogRow, 0)
	for rows.Next() {
		var worklog storedWorklogRow
		require.NoError(t, rows.Scan(
			&worklog.id,
			&worklog.issueKey,
			&worklog.beginTS,
			&worklog.endTS,
			&worklog.comment,
			&worklog.active,
			&worklog.synced,
		))
		worklogs = append(worklogs, worklog)
	}
	require.NoError(t, rows.Err())

	return worklogs
}

func insertTestWorklogRow(t *testing.T, store *SQLiteStore, worklog testWorklogRow) int {
	t.Helper()

	result, err := store.db.Exec(`
INSERT INTO
    issue_log (
        issue_key,
        begin_ts,
        end_ts,
        comment,
        active,
        synced
    )
VALUES
    (?, ?, ?, ?, ?, ?);
`,
		worklog.issueKey,
		worklog.beginTS,
		worklog.endTS,
		worklog.comment,
		worklog.active,
		worklog.synced,
	)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}
