package persistence

import (
	"testing"
	"time"

	"github.com/dhth/punchout/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // sqlite driver
)

func TestSQLiteStoreActiveWorklog(t *testing.T) {
	t.Run("returns the active worklog", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)
		location := time.FixedZone("UTC+1", int(time.Hour.Seconds()))
		beginTS := time.Date(2026, time.August, 24, 9, 30, 0, 0, location)
		endTS := beginTS.Add(time.Hour).UTC()
		comment := "active worklog"
		insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "COMPLETED-1",
			beginTS:  beginTS.UTC(),
			endTS:    &endTS,
		})
		insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "ACTIVE-1",
			beginTS:  beginTS.UTC(),
			comment:  &comment,
			active:   true,
		})

		// WHEN
		got, err := store.ActiveWorklog(t.Context())

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &domain.InProgressWorklog{
			IssueKey: "ACTIVE-1",
			BeginTS:  beginTS.UTC().Local(),
			Comment:  comment,
		}, got)
	})

	t.Run("normalizes a null comment to an empty string", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)
		beginTS := time.Date(2026, time.August, 24, 9, 30, 0, 0, time.UTC)
		insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "ACTIVE-1",
			beginTS:  beginTS,
			active:   true,
		})

		// WHEN
		got, err := store.ActiveWorklog(t.Context())

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &domain.InProgressWorklog{
			IssueKey: "ACTIVE-1",
			BeginTS:  beginTS.Local(),
			Comment:  "",
		}, got)
	})

	t.Run("returns nil when there is no active worklog", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)
		beginTS := time.Date(2026, time.August, 24, 9, 30, 0, 0, time.UTC)
		endTS := beginTS.Add(time.Hour)
		insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "COMPLETED-1",
			beginTS:  beginTS,
			endTS:    &endTS,
		})

		// WHEN
		got, err := store.ActiveWorklog(t.Context())

		// THEN
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestSQLiteStoreStartWorklog(t *testing.T) {
	t.Run("starts a worklog", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)
		location := time.FixedZone("UTC+1", int(time.Hour.Seconds()))
		beginTS := time.Date(2026, time.August, 24, 9, 30, 0, 0, location)

		// WHEN
		err := store.StartWorklog(t.Context(), "TEST-1", beginTS)

		// THEN
		require.NoError(t, err)
		gotRows := fetchAllWorklogRows(t, store)
		require.Len(t, gotRows, 1)
		got := gotRows[0]

		assert.Equal(t, "TEST-1", got.issueKey)
		assert.Equal(t, beginTS.UTC(), got.beginTS)
		assert.False(t, got.endTS.Valid)
		assert.False(t, got.comment.Valid)
		assert.True(t, got.active)
		assert.False(t, got.synced)
	})

	t.Run("rejects a second active worklog", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)
		firstBeginTS := time.Date(2026, time.August, 24, 9, 30, 0, 0, time.UTC)
		require.NoError(t, store.StartWorklog(t.Context(), "TEST-1", firstBeginTS))

		// WHEN
		err := store.StartWorklog(t.Context(), "TEST-2", firstBeginTS.Add(time.Hour))

		// THEN
		require.ErrorContains(t, err, "Only one row with active=1 is allowed")
		gotRows := fetchAllWorklogRows(t, store)
		require.Len(t, gotRows, 1)
		got := gotRows[0]

		assert.Equal(t, "TEST-1", got.issueKey)
		assert.Equal(t, firstBeginTS, got.beginTS)
		assert.True(t, got.active)
	})
}

func TestSQLiteStoreFinishWorklog(t *testing.T) {
	t.Run("finishes the matching active worklog", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)
		location := time.FixedZone("UTC+1", int(time.Hour.Seconds()))
		originalBeginTS := time.Date(2026, time.August, 24, 9, 0, 0, 0, time.UTC)
		beginTS := time.Date(2026, time.August, 24, 10, 0, 0, 0, location)
		endTS := beginTS.Add(time.Hour)
		completedEndTS := originalBeginTS.Add(30 * time.Minute)
		completedComment := "existing completed worklog"
		completedID := insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "TEST-1",
			beginTS:  originalBeginTS,
			endTS:    &completedEndTS,
			comment:  &completedComment,
		})
		insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "TEST-1",
			beginTS:  originalBeginTS,
			active:   true,
		})
		worklog := domain.Worklog{
			IssueKey: "TEST-1",
			BeginTS:  beginTS,
			EndTS:    endTS,
			Comment:  "finished worklog",
		}

		// WHEN
		err := store.FinishWorklog(t.Context(), worklog)

		// THEN
		require.NoError(t, err)
		gotRows := fetchAllWorklogRows(t, store)
		require.Len(t, gotRows, 2)
		completed := gotRows[0]
		got := gotRows[1]

		assert.Equal(t, completedID, completed.id)
		assert.Equal(t, originalBeginTS, completed.beginTS)
		require.True(t, completed.endTS.Valid)
		assert.Equal(t, completedEndTS, completed.endTS.Time)
		require.True(t, completed.comment.Valid)
		assert.Equal(t, completedComment, completed.comment.String)
		assert.False(t, completed.active)

		assert.Equal(t, "TEST-1", got.issueKey)
		assert.Equal(t, beginTS.UTC(), got.beginTS)
		require.True(t, got.endTS.Valid)
		assert.Equal(t, endTS.UTC(), got.endTS.Time)
		require.True(t, got.comment.Valid)
		assert.Equal(t, "finished worklog", got.comment.String)
		assert.False(t, got.active)
		assert.False(t, got.synced)
	})

	t.Run("returns an error when a different issue is active", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)
		beginTS := time.Date(2026, time.August, 24, 9, 0, 0, 0, time.UTC)
		insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "TEST-1",
			beginTS:  beginTS,
			active:   true,
		})

		// WHEN
		err := store.FinishWorklog(t.Context(), domain.Worklog{
			IssueKey: "TEST-2",
			BeginTS:  beginTS,
			EndTS:    beginTS.Add(time.Hour),
		})

		// THEN
		require.Error(t, err)
		require.ErrorIs(t, err, ErrIssueHasNoActiveWorklog)
		require.ErrorContains(t, err, "TEST-2")
		gotRows := fetchAllWorklogRows(t, store)
		require.Len(t, gotRows, 1)
		got := gotRows[0]

		assert.Equal(t, "TEST-1", got.issueKey)
		assert.Equal(t, beginTS, got.beginTS)
		assert.True(t, got.active)
	})

	t.Run("returns an error when there is no active worklog", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)
		beginTS := time.Date(2026, time.August, 24, 9, 0, 0, 0, time.UTC)

		// WHEN
		err := store.FinishWorklog(t.Context(), domain.Worklog{
			IssueKey: "TEST-1",
			BeginTS:  beginTS,
			EndTS:    beginTS.Add(time.Hour),
		})

		// THEN
		require.Error(t, err)
		require.ErrorIs(t, err, ErrIssueHasNoActiveWorklog)
		require.ErrorContains(t, err, "TEST-1")
	})
}

func TestSQLiteStoreAddWorklog(t *testing.T) {
	// GIVEN
	store := setupTestStore(t)
	location := time.FixedZone("UTC+1", int(time.Hour.Seconds()))
	worklog := domain.Worklog{
		IssueKey: "TEST-1",
		BeginTS:  time.Date(2026, time.August, 24, 9, 30, 0, 0, location),
		EndTS:    time.Date(2026, time.August, 24, 10, 45, 0, 0, location),
		Comment:  "implemented persistence boundary",
	}

	// WHEN
	err := store.AddWorklog(t.Context(), worklog)

	// THEN
	require.NoError(t, err)
	gotRows := fetchAllWorklogRows(t, store)
	require.Len(t, gotRows, 1)
	got := gotRows[0]

	assert.Equal(t, worklog.IssueKey, got.issueKey)
	assert.Equal(t, worklog.BeginTS.UTC(), got.beginTS)
	require.True(t, got.endTS.Valid)
	assert.Equal(t, worklog.EndTS.UTC(), got.endTS.Time)
	require.True(t, got.comment.Valid)
	assert.Equal(t, worklog.Comment, got.comment.String)
	assert.False(t, got.active)
	assert.False(t, got.synced)
}

func TestSQLiteStoreUnsyncedWorklogs(t *testing.T) {
	t.Run("returns completed unsynced worklogs newest first", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)
		location := time.FixedZone("UTC+1", int(time.Hour.Seconds()))
		olderBegin := time.Date(2026, time.August, 24, 9, 0, 0, 0, location)
		olderEnd := olderBegin.Add(time.Hour)
		newerBegin := olderEnd.Add(time.Hour)
		newerEnd := newerBegin.Add(90 * time.Minute)
		olderEndUTC := olderEnd.UTC()
		newerEndUTC := newerEnd.UTC()
		olderComment := "older worklog"

		olderID := insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "TEST-1",
			beginTS:  olderBegin.UTC(),
			endTS:    &olderEndUTC,
			comment:  &olderComment,
		})
		newerID := insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "TEST-2",
			beginTS:  newerBegin.UTC(),
			endTS:    &newerEndUTC,
			comment:  nil,
		})
		insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "SYNCED-1",
			beginTS:  olderBegin.UTC(),
			endTS:    &olderEndUTC,
			synced:   true,
		})
		insertTestWorklogRow(t, store, testWorklogRow{
			issueKey: "ACTIVE-1",
			beginTS:  newerBegin.UTC(),
			endTS:    nil,
			active:   true,
		})

		// WHEN
		got, err := store.UnsyncedWorklogs(t.Context())

		// THEN
		require.NoError(t, err)
		assert.Equal(t, []domain.StoredWorklog{
			{
				ID: newerID,
				Worklog: domain.Worklog{
					IssueKey: "TEST-2",
					BeginTS:  newerBegin.UTC().Local(),
					EndTS:    newerEnd.UTC().Local(),
					Comment:  "",
				},
				Synced: false,
			},
			{
				ID: olderID,
				Worklog: domain.Worklog{
					IssueKey: "TEST-1",
					BeginTS:  olderBegin.UTC().Local(),
					EndTS:    olderEnd.UTC().Local(),
					Comment:  "older worklog",
				},
				Synced: false,
			},
		}, got)
	})

	t.Run("returns an empty non-nil slice when there are no unsynced worklogs", func(t *testing.T) {
		// GIVEN
		store := setupTestStore(t)

		// WHEN
		got, err := store.UnsyncedWorklogs(t.Context())

		// THEN
		require.NoError(t, err)
		assert.Empty(t, got)
		assert.NotNil(t, got)
	})
}

func TestSQLiteStoreMarkWorklogSynced(t *testing.T) {
	tests := []struct {
		name            string
		storedComment   string
		updatedComment  *string
		expectedComment string
	}{
		{
			name:            "preserves the existing comment when no comment is supplied",
			storedComment:   "existing comment",
			updatedComment:  nil,
			expectedComment: "existing comment",
		},
		{
			name:            "replaces the existing comment when a comment is supplied",
			storedComment:   "existing comment",
			updatedComment:  ptrTo("fallback comment"),
			expectedComment: "fallback comment",
		},
		{
			name:            "stores an empty supplied comment",
			storedComment:   "existing comment",
			updatedComment:  ptrTo(""),
			expectedComment: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// GIVEN
			store := setupTestStore(t)
			beginTS := time.Date(2026, time.August, 24, 9, 0, 0, 0, time.UTC)
			endTS := beginTS.Add(time.Hour)
			id := insertTestWorklogRow(t, store, testWorklogRow{
				issueKey: "TEST-1",
				beginTS:  beginTS,
				endTS:    &endTS,
				comment:  &test.storedComment,
			})

			// WHEN
			err := store.MarkWorklogSynced(t.Context(), id, test.updatedComment)

			// THEN
			require.NoError(t, err)
			gotRows := fetchAllWorklogRows(t, store)
			require.Len(t, gotRows, 1)
			got := gotRows[0]

			assert.Equal(t, id, got.id)
			assert.Equal(t, "TEST-1", got.issueKey)
			assert.Equal(t, beginTS, got.beginTS)
			require.True(t, got.endTS.Valid)
			assert.Equal(t, endTS, got.endTS.Time)
			require.True(t, got.comment.Valid)
			assert.Equal(t, test.expectedComment, got.comment.String)
			assert.False(t, got.active)
			assert.True(t, got.synced)
		})
	}
}

func ptrTo[T any](value T) *T {
	return &value
}
