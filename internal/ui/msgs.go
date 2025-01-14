package ui

import (
	"time"

	c "github.com/dhth/punchout/internal/common"
)

type hideHelpMsg struct{}

type trackingToggledMsg struct {
	activeIssue string
	finished    bool
	err         error
}

type activeIssueSwitchedMsg struct {
	lastActiveIssue    string
	currentActiveIssue string
	beginTs            time.Time
	err                error
}

type manualEntryInserted struct {
	issueKey string
	err      error
}

type activeTaskLogDeletedMsg struct {
	err error
}

type manualEntryUpdated struct {
	rowID    int
	issueKey string
	err      error
}

type fetchActiveMsg struct {
	activeIssue string
	beginTs     time.Time
	err         error
}

type logEntriesFetchedMsg struct {
	entries []c.WorklogEntry
	err     error
}

type syncedLogEntriesFetchedMsg struct {
	entries []c.SyncedWorklogEntry
	err     error
}

type logEntriesDeletedMsg struct {
	err error
}

type logEntrySyncUpdated struct {
	entry c.WorklogEntry
	index int
	err   error
}

type issuesFetchedFromJIRAMsg struct {
	issues             []c.Issue
	responseStatusCode int
	err                error
}

type wlAddedOnJIRA struct {
	index               int
	entry               c.WorklogEntry
	fallbackCommentUsed bool
	err                 error
}

type urlOpenedinBrowserMsg struct {
	url string
	err error
}
