package persistence

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLiteStoreInitializesDB(t *testing.T) {
	// GIVEN
	dbPath := filepath.Join(t.TempDir(), "nested", "punchout.db")

	// WHEN
	store, err := NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	// THEN
	assert.Equal(t, 1, store.db.Stats().MaxOpenConnections)

	_, err = store.db.Exec(`
INSERT INTO
    issue_log (issue_key, begin_ts, active, synced)
VALUES
    (?, ?, TRUE, false);
`,
		"TEST-1",
		time.Now().UTC(),
	)
	assert.NoError(t, err)
}

func TestNewSQLiteStorePrunesOldWorklogs(t *testing.T) {
	// GIVEN
	dbPath := filepath.Join(t.TempDir(), "punchout.db")

	store, err := NewSQLiteStore(dbPath)
	require.NoError(t, err)
	now := time.Now().UTC()

	_, err = store.db.Exec(`
INSERT INTO
    issue_log (issue_key, begin_ts, end_ts, active, synced)
VALUES
    (?, ?, ?, false, false);
	`,
		"OLD-1",
		now.AddDate(0, 0, -62),
		now.AddDate(0, 0, -61),
	)
	require.NoError(t, err)
	require.NoError(t, store.Close())

	// WHEN
	store, err = NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	// THEN
	var count int
	err = store.db.QueryRow(`
SELECT
    COUNT(*)
FROM
    issue_log
WHERE
    issue_key = ?
`,
		"OLD-1",
	).Scan(&count)

	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestSQLiteStoreClosesCorrectly(t *testing.T) {
	// GIVEN
	store, err := NewSQLiteStore(filepath.Join(t.TempDir(), "punchout.db"))
	require.NoError(t, err)

	// WHEN
	require.NoError(t, store.Close())

	// THEN
	require.Error(t, store.db.Ping())
}
