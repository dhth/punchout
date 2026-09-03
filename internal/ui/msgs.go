package ui

import (
	"time"

	d "github.com/dhth/punchout/internal/domain"
)

type hideHelpMsg struct{}

type clearUserMsgMsg struct {
	id uint64
}

type trackingToggleOperation uint

const (
	trackingToggleUnknown trackingToggleOperation = iota
	trackingToggleStart
	trackingToggleFinish
)

type trackingToggledInDB struct {
	activeIssue           string
	finished              bool
	operation             trackingToggleOperation
	reconcileActiveStatus bool
	err                   error
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
	worklog *d.InProgressWorklog
	err     error
}

type wLEntriesFetchedFromDB struct {
	entries []d.StoredWorklog
	err     error
}

type syncedWLEntriesFetchedFromDB struct {
	entries []d.StoredWorklog
	err     error
}

type wLDeletedFromDB struct {
	err error
}

type wLSyncUpdatedInDB struct {
	entry     worklogListItem
	indexHint int
	err       error
}

type issueSource uint

const (
	issueSourceJIRA issueSource = iota
	issueSourceCache
)

type issuesLoaded struct {
	issues                []d.Issue
	fetchedAt             time.Time
	source                issueSource
	afterCacheLoadFailure bool
	err                   error
}

type issuesSavedToCache struct {
	err error
}

type wLSyncedToJIRA struct {
	indexHint           int
	entry               worklogListItem
	fallbackCommentUsed bool
	err                 error
}

type urlOpenedinBrowserMsg struct {
	url string
	err error
}
