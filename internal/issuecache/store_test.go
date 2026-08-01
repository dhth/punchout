package issuecache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dhth/punchout/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreRoundTripsSnapshot(t *testing.T) {
	store := Store{
		filePath: filepath.Join(t.TempDir(), "issues", "issues.json"),
	}
	expected := Snapshot{
		Issues: []domain.Issue{
			{
				IssueKey:        "TEST-123",
				IssueType:       "Task",
				Summary:         "Implement issue caching",
				Assignee:        "user@example.com",
				Status:          "In Progress",
				AggSecondsSpent: 3600,
			},
			{
				IssueKey:  "TEST-456",
				IssueType: "Bug",
				Summary:   "Fix cache loading",
				Status:    "Open",
			},
		},
		FetchedAt: time.Date(2026, time.August, 1, 12, 30, 0, 0, time.UTC),
	}

	err := store.Save(expected)
	require.NoError(t, err, "saving snapshot failed")

	actual, err := store.Load()
	require.NoError(t, err, "loading snapshot failed")
	assert.Equal(t, expected, actual)
}

func TestStoreNormalizesNilIssues(t *testing.T) {
	store := Store{
		filePath: filepath.Join(t.TempDir(), "issues", "issues.json"),
	}
	snapshot := Snapshot{
		FetchedAt: time.Date(2026, time.August, 1, 12, 30, 0, 0, time.UTC),
	}

	err := store.Save(snapshot)
	require.NoError(t, err, "saving snapshot failed")

	actual, err := store.Load()
	require.NoError(t, err, "loading snapshot failed")
	assert.NotNil(t, actual.Issues)
	assert.Empty(t, actual.Issues)
}

func TestStoreRejectsSnapshotWithZeroFetchedAt(t *testing.T) {
	store := Store{
		filePath: filepath.Join(t.TempDir(), "issues", "issues.json"),
	}

	err := store.Save(Snapshot{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetched-at timestamp is zero")
}

func TestStoreRejectsInvalidCacheFiles(t *testing.T) {
	cases := []struct {
		name          string
		contents      string
		expectedError string
	}{
		{
			name:          "JSON is malformed",
			contents:      `{`,
			expectedError: "couldn't decode issue cache",
		},
		{
			name:          "issues are missing",
			contents:      `{"fetched_at":"2026-08-01T12:30:00Z"}`,
			expectedError: "issues are missing or null",
		},
		{
			name:          "issues are null",
			contents:      `{"issues":null,"fetched_at":"2026-08-01T12:30:00Z"}`,
			expectedError: "issues are missing or null",
		},
		{
			name:          "fetched-at timestamp is missing",
			contents:      `{"issues":[]}`,
			expectedError: "zero fetched-at timestamp",
		},
		{
			name:          "fetched-at timestamp is null",
			contents:      `{"issues":[],"fetched_at":null}`,
			expectedError: "zero fetched-at timestamp",
		},
		{
			name:          "fetched-at timestamp is zero",
			contents:      `{"issues":[],"fetched_at":"0001-01-01T00:00:00Z"}`,
			expectedError: "zero fetched-at timestamp",
		},
		{
			name:          "fetched-at timestamp is malformed",
			contents:      `{"issues":[],"fetched_at":"not-a-timestamp"}`,
			expectedError: "couldn't decode issue cache",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := Store{
				filePath: filepath.Join(t.TempDir(), "issues.json"),
			}
			err := os.WriteFile(store.filePath, []byte(tc.contents), 0o600)
			require.NoError(t, err, "writing cache file failed")

			_, err = store.Load()
			require.Error(t, err, "loading should've failed")
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestStoreReplacesExistingSnapshot(t *testing.T) {
	store := Store{
		filePath: filepath.Join(t.TempDir(), "issues", "issues.json"),
	}
	initial := Snapshot{
		Issues: []domain.Issue{
			{
				IssueKey:  "TEST-123",
				IssueType: "Task",
				Summary:   "Initial issue",
				Status:    "Open",
			},
		},
		FetchedAt: time.Date(2026, time.August, 1, 12, 30, 0, 0, time.UTC),
	}
	replacement := Snapshot{
		Issues: []domain.Issue{
			{
				IssueKey:        "TEST-456",
				IssueType:       "Bug",
				Summary:         "Replacement issue",
				Assignee:        "user@example.com",
				Status:          "In Progress",
				AggSecondsSpent: 1800,
			},
		},
		FetchedAt: time.Date(2026, time.August, 1, 13, 30, 0, 0, time.UTC),
	}

	err := store.Save(initial)
	require.NoError(t, err, "saving initial snapshot failed")

	err = store.Save(replacement)
	require.NoError(t, err, "saving replacement snapshot failed")

	actual, err := store.Load()
	require.NoError(t, err, "loading snapshot failed")
	assert.Equal(t, replacement, actual)
}
