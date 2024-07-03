package ui

import "time"

type hideHelpMsg struct{}

type trackingToggledMsg struct {
	activeIssue string
	finished    bool
	err         error
}

type manualEntryInserted struct {
	issueKey string
	err      error
}

type activeTaskLogDeletedMsg struct {
	err error
}

type manualEntryUpdated struct {
	rowId    int
	issueKey string
	err      error
}

type fetchActiveMsg struct {
	activeIssue string
	beginTs     time.Time
	err         error
}

type logEntriesFetchedMsg struct {
	entries []worklogEntry
	err     error
}

type syncedLogEntriesFetchedMsg struct {
	entries []syncedWorklogEntry
	err     error
}

type logEntriesDeletedMsg struct {
	err error
}

type logEntrySyncUpdated struct {
	entry worklogEntry
	index int
	err   error
}

type issuesFetchedFromJIRAMsg struct {
	issues []Issue
	err    error
}

type wlAddedOnJIRA struct {
	index int
	entry worklogEntry
	err   error
}

type urlOpenedinBrowserMsg struct {
	url string
	err error
}
