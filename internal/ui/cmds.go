package ui

import (
	"database/sql"
	"os/exec"
	"runtime"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	tea "github.com/charmbracelet/bubbletea"
	_ "modernc.org/sqlite"
)

func toggleTracking(db *sql.DB, selectedIssue string, beginTs, endTs time.Time, comment string) tea.Cmd {
	return func() tea.Msg {
		row := db.QueryRow(`
SELECT issue_key
from issue_log
WHERE active=1
ORDER BY begin_ts DESC
LIMIT 1
`)
		var trackStatus trackingStatus
		var activeIssue string
		err := row.Scan(&activeIssue)
		if err == sql.ErrNoRows {
			trackStatus = trackingInactive
		} else if err != nil {
			return trackingToggledMsg{err: err}
		} else {
			trackStatus = trackingActive
		}

		switch trackStatus {
		case trackingInactive:
			err = insertNewEntry(db, selectedIssue, beginTs)
			if err != nil {
				return trackingToggledMsg{err: err}
			} else {
				return trackingToggledMsg{activeIssue: selectedIssue}
			}

		default:
			err := updateLastEntry(db, activeIssue, comment, beginTs, endTs)
			if err != nil {
				return trackingToggledMsg{err: err}
			} else {
				return trackingToggledMsg{activeIssue: "", finished: true}
			}
		}
	}
}

func insertManualEntry(db *sql.DB, issueKey string, beginTS time.Time, endTS time.Time, comment string) tea.Cmd {
	return func() tea.Msg {
		stmt, err := db.Prepare(`
INSERT INTO issue_log (issue_key, begin_ts, end_ts, comment, active, synced)
VALUES (?, ?, ?, ?, ?, ?);
`)
		if err != nil {
			return manualEntryInserted{issueKey, err}
		}
		defer stmt.Close()

		_, err = stmt.Exec(issueKey, beginTS, endTS, comment, false, false)
		if err != nil {
			return manualEntryInserted{issueKey, err}
		}

		return manualEntryInserted{issueKey, nil}
	}
}

func deleteActiveIssueLog(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		err := deleteActiveLogInDB(db)
		return activeTaskLogDeletedMsg{err}
	}
}

func updateManualEntry(db *sql.DB, rowID int, issueKey string, beginTS time.Time, endTS time.Time, comment string) tea.Cmd {
	return func() tea.Msg {
		stmt, err := db.Prepare(`
UPDATE issue_log
SET begin_ts = ?,
    end_ts = ?,
    comment = ?
WHERE ID = ?;
`)
		if err != nil {
			return manualEntryUpdated{rowID, issueKey, err}
		}
		defer stmt.Close()

		_, err = stmt.Exec(beginTS, endTS, comment, rowID)
		if err != nil {
			return manualEntryUpdated{rowID, issueKey, err}
		}

		return manualEntryUpdated{rowID, issueKey, nil}
	}
}

func fetchActiveStatus(db *sql.DB, interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		row := db.QueryRow(`
SELECT issue_key, begin_ts
from issue_log
WHERE active=1
ORDER BY begin_ts DESC
LIMIT 1
`)
		var activeIssue string
		var beginTs time.Time
		err := row.Scan(&activeIssue, &beginTs)
		if err == sql.ErrNoRows {
			return fetchActiveMsg{activeIssue: activeIssue}
		}
		if err != nil {
			return fetchActiveMsg{err: err}
		}

		return fetchActiveMsg{activeIssue: activeIssue, beginTs: beginTs}
	})
}

func fetchLogEntries(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		entries, err := fetchEntries(db)
		return logEntriesFetchedMsg{
			entries: entries,
			err:     err,
		}
	}
}

func fetchSyncedLogEntries(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		entries, err := fetchSyncedEntries(db)
		return syncedLogEntriesFetchedMsg{
			entries: entries,
			err:     err,
		}
	}
}

func deleteLogEntry(db *sql.DB, id int) tea.Cmd {
	return func() tea.Msg {
		err := deleteEntry(db, id)
		return logEntriesDeletedMsg{
			err: err,
		}
	}
}

func updateSyncStatusForEntry(db *sql.DB, entry worklogEntry, index int) tea.Cmd {
	return func() tea.Msg {
		err := updateSyncStatus(db, entry.Id)
		return logEntrySyncUpdated{
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
			var assignee string
			var totalSecsSpent int
			var status string
			if issue.Fields != nil {
				if issue.Fields.Assignee != nil {
					assignee = issue.Fields.Assignee.Name
				}

				totalSecsSpent = issue.Fields.AggregateTimeSpent

				if issue.Fields.Status != nil {
					status = issue.Fields.Status.Name
				}
			}
			issues = append(issues, Issue{
				issueKey:        issue.Key,
				issueType:       issue.Fields.Type.Name,
				summary:         issue.Fields.Summary,
				assignee:        assignee,
				status:          status,
				aggSecondsSpent: totalSecsSpent,
				trackingActive:  false,
			})
		}
		return issuesFetchedFromJIRAMsg{issues, err}
	}
}

func syncWorklogWithJIRA(cl *jira.Client, entry worklogEntry, index int, timeDeltaMins int) tea.Cmd {
	return func() tea.Msg {
		err := addWLtoJira(cl, entry, timeDeltaMins)
		return wlAddedOnJIRA{index, entry, err}
	}
}

func hideHelp(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return hideHelpMsg{}
	})
}

func openURLInBrowser(url string) tea.Cmd {
	var openCmd string
	switch runtime.GOOS {
	case "darwin":
		openCmd = "open"
	default:
		openCmd = "xdg-open"
	}
	c := exec.Command(openCmd, url)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return urlOpenedinBrowserMsg{url: url, err: err}
		}
		return tea.Msg(urlOpenedinBrowserMsg{url: url})
	})
}
