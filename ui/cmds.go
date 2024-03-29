package ui

import (
	"database/sql"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	tea "github.com/charmbracelet/bubbletea"
	_ "modernc.org/sqlite"
)

func toggleTracking(db *sql.DB, selectedIssue string, comment string) tea.Cmd {
	return func() tea.Msg {

		row := db.QueryRow(`
SELECT issue_key
from issue_log
WHERE active=1
ORDER BY begin_ts DESC
LIMIT 1
`)
		var trackStatus TrackingStatus
		var activeIssue string
		err := row.Scan(&activeIssue)
		if err == sql.ErrNoRows {
			trackStatus = TrackingInactive
		} else if err != nil {
			return TrackingToggledMsg{err: err}
		} else {
			trackStatus = TrackingActive
		}

		switch trackStatus {
		case TrackingInactive:
			err = insertNewEntry(db, selectedIssue)
			if err != nil {
				return TrackingToggledMsg{err: err}
			} else {
				return TrackingToggledMsg{activeIssue: selectedIssue}
			}

		default:
			err := updateLastEntry(db, activeIssue, comment)
			if err != nil {
				return TrackingToggledMsg{err: err}
			} else {
				return TrackingToggledMsg{activeIssue: "", finished: true}
			}
		}
	}
}

func insertEntry(db *sql.DB, issue_key string) tea.Cmd {
	return func() tea.Msg {

		stmt, err := db.Prepare(`
    INSERT INTO issue_log (issue_key, begin_ts, active, synced)
    VALUES (?, ?, ?, ?);
    `)

		if err != nil {
			return InsertEntryMsg{issue_key, err}
		}
		defer stmt.Close()

		_, err = stmt.Exec(issue_key, time.Now(), true, 0)
		if err != nil {
			return InsertEntryMsg{issue_key, err}
		}

		return InsertEntryMsg{issue_key, nil}
	}
}

func insertManualEntry(db *sql.DB, issueKey string, beginTS time.Time, endTS time.Time, comment string) tea.Cmd {
	return func() tea.Msg {

		stmt, err := db.Prepare(`
    INSERT INTO issue_log (issue_key, begin_ts, end_ts, comment, active, synced)
    VALUES (?, ?, ?, ?, ?, ?);
    `)

		if err != nil {
			return ManualEntryInserted{issueKey, err}
		}
		defer stmt.Close()

		_, err = stmt.Exec(issueKey, beginTS, endTS, comment, false, false)
		if err != nil {
			return ManualEntryInserted{issueKey, err}
		}

		return ManualEntryInserted{issueKey, nil}
	}
}

func updateEntry(db *sql.DB, issue_key string) tea.Cmd {
	return func() tea.Msg {

		stmt, err := db.Prepare(`
UPDATE issue_log
SET active = 0,
    end_ts = ?
WHERE issue_key = ?
AND active = 1;
`)
		if err != nil {
			return UpdateEntryMsg{issue_key, err}
		}
		defer stmt.Close()

		_, err = stmt.Exec(time.Now(), issue_key)
		if err != nil {
			return UpdateEntryMsg{issue_key, err}
		}

		return UpdateEntryMsg{issue_key, nil}
	}
}

func fetchActiveStatus(db *sql.DB, interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		row := db.QueryRow(`
SELECT issue_key
from issue_log
WHERE active=1
ORDER BY begin_ts DESC
LIMIT 1
`)
		var activeIssue string
		err := row.Scan(&activeIssue)
		if err == sql.ErrNoRows {
			return FetchActiveMsg{activeIssue: activeIssue}
		}
		if err != nil {
			return FetchActiveMsg{err: err}
		}

		return FetchActiveMsg{activeIssue: activeIssue}
	})
}

func fetchLogEntries(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		entries, err := fetchEntries(db)
		return LogEntriesFetchedMsg{
			entries: entries,
			err:     err,
		}
	}
}

func deleteLogEntry(db *sql.DB, id int) tea.Cmd {
	return func() tea.Msg {
		err := deleteEntry(db, id)
		return LogEntriesDeletedMsg{
			err: err,
		}
	}
}

func updateSyncStatusForEntry(db *sql.DB, entry WorklogEntry, index int) tea.Cmd {
	return func() tea.Msg {
		err := updateSyncStatus(db, entry.Id)
		return LogEntrySyncUpdated{
			entry: entry,
			index: index,
			err:   err,
		}
	}
}

func fetchJIRAIssues(cl *jira.Client, jql string) tea.Cmd {
	return func() tea.Msg {
		jIssues, err := getIssues(cl, jql)
		var issues []Issue
		for _, issue := range jIssues {
			issues = append(issues, Issue{issue.Key, issue.Fields.Type.Name, issue.Fields.Summary})
		}
		return IssuesFetchedFromJIRAMsg{issues, err}
	}
}

func syncWorklogWithJIRA(cl *jira.Client, entry WorklogEntry, index int, timeDeltaMins int) tea.Cmd {
	return func() tea.Msg {
		err := addWLtoJira(cl, entry, timeDeltaMins)
		return WLAddedOnJIRA{index, entry, err}
	}
}

func hideHelp(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return HideHelpMsg{}
	})
}
