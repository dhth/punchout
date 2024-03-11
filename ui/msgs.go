package ui

import (
	"database/sql"
)

type HideHelpMsg struct{}

type IssueToggledMsg struct {
	index int
	issue Issue
}

type SetupDBMsg struct {
	db  *sql.DB
	err error
}

type TrackingToggledMsg struct {
	activeIssue string
	finished    bool
	err         error
}

type InsertEntryMsg struct {
	issueKey string
	err      error
}

type UpdateEntryMsg struct {
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
