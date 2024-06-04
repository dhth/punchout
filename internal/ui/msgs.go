package ui

type HideHelpMsg struct{}

type TrackingToggledMsg struct {
	activeIssue string
	issueIndex  int
	finished    bool
	err         error
}

type ManualEntryInserted struct {
	issueKey string
	err      error
}

type ManualEntryUpdated struct {
	rowId    int
	issueKey string
	err      error
}

type FetchActiveMsg struct {
	activeIssue string
	err         error
}

type LogEntriesFetchedMsg struct {
	entries []WorklogEntry
	err     error
}

type SyncedLogEntriesFetchedMsg struct {
	entries []SyncedWorklogEntry
	err     error
}

type LogEntriesDeletedMsg struct {
	err error
}

type LogEntrySyncUpdated struct {
	entry WorklogEntry
	index int
	err   error
}

type IssuesFetchedFromJIRAMsg struct {
	issues []Issue
	err    error
}

type WLAddedOnJIRA struct {
	index int
	entry WorklogEntry
	err   error
}

type URLOpenedinBrowserMsg struct {
	url string
	err error
}
