package ui

import (
	"time"

	d "github.com/dhth/punchout/internal/domain"
)

type hideHelpMsg struct{}

type trackingToggledInDB struct {
	activeIssue string
	finished    bool
	err         error
}

type activeWLSwitchedInDB struct {
	lastActiveIssue    string
	currentActiveIssue string
	beginTS            time.Time
	err                error
}

type activeWLUpdatedInDB struct {
	beginTS time.Time
	comment *string
	err     error
}

type manualWLInsertedInDB struct {
	issueKey string
	err      error
}

type activeWLDeletedFromDB struct {
	err error
}

type wLUpdatedInDB struct {
	rowID    int
	issueKey string
	err      error
}

type activeWLFetchedFromDB struct {
	activeIssue string
	beginTS     time.Time
	comment     *string
	err         error
}

type wLEntriesFetchedFromDB struct {
	entries []d.WorklogEntry
	err     error
}

type syncedWLEntriesFetchedFromDB struct {
	entries []d.SyncedWorklogEntry
	err     error
}

type wLDeletedFromDB struct {
	err error
}

type wLSyncUpdatedInDB struct {
	entry d.WorklogEntry
	index int
	err   error
}

type issuesFetchedFromJIRA struct {
	issues             []d.Issue
	responseStatusCode int
	err                error
}

type wLSyncedToJIRA struct {
	index               int
	entry               d.WorklogEntry
	fallbackCommentUsed bool
	err                 error
}

type urlOpenedinBrowserMsg struct {
	url string
	err error
}
