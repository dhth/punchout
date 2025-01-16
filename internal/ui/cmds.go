package ui

import (
	"database/sql"
	"errors"
	"os/exec"
	"runtime"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	tea "github.com/charmbracelet/bubbletea"
	common "github.com/dhth/punchout/internal/common"
	pers "github.com/dhth/punchout/internal/persistence"

	_ "modernc.org/sqlite" // sqlite driver
)

var errWorklogsEndTSIsEmpty = errors.New("worklog's end timestamp is empty")

func toggleTracking(db *sql.DB, selectedIssue string, beginTS, endTS time.Time, comment string) tea.Cmd {
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
		if errors.Is(err, sql.ErrNoRows) {
			trackStatus = trackingInactive
		} else if err != nil {
			return trackingToggledInDB{err: err}
		} else {
			trackStatus = trackingActive
		}

		switch trackStatus {
		case trackingInactive:
			err = pers.InsertNewWLInDB(db, selectedIssue, beginTS)
			if err != nil {
				return trackingToggledInDB{err: err}
			} else {
				return trackingToggledInDB{activeIssue: selectedIssue}
			}

		default:
			err := pers.UpdateActiveWLInDB(db, activeIssue, comment, beginTS, endTS)
			if err != nil {
				return trackingToggledInDB{err: err}
			} else {
				return trackingToggledInDB{activeIssue: "", finished: true}
			}
		}
	}
}

func quickSwitchActiveIssue(db *sql.DB, selectedIssue string, currentTime time.Time) tea.Cmd {
	return func() tea.Msg {
		activeIssue, err := pers.GetActiveIssueFromDB(db)
		if err != nil {
			return activeWLSwitchedInDB{"", selectedIssue, currentTime, err}
		}

		err = pers.QuickSwitchActiveWLInDB(db, activeIssue, selectedIssue, currentTime)
		if err != nil {
			return activeWLSwitchedInDB{activeIssue, selectedIssue, currentTime, err}
		}

		return activeWLSwitchedInDB{activeIssue, selectedIssue, currentTime, nil}
	}
}

func updateActiveWL(db *sql.DB, beginTS time.Time, comment *string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if comment == nil {
			err = pers.UpdateActiveWLBeginTSInDB(db, beginTS)
		} else {
			err = pers.UpdateActiveWLBeginTSAndCommentInDB(db, beginTS, *comment)
		}

		return activeWLUpdatedInDB{beginTS, comment, err}
	}
}

func insertManualEntry(db *sql.DB, issueKey string, beginTS time.Time, endTS time.Time, comment string) tea.Cmd {
	return func() tea.Msg {
		stmt, err := db.Prepare(`
INSERT INTO issue_log (issue_key, begin_ts, end_ts, comment, active, synced)
VALUES (?, ?, ?, ?, ?, ?);
`)
		if err != nil {
			return manualWLInsertedInDB{issueKey, err}
		}
		defer stmt.Close()

		_, err = stmt.Exec(issueKey, beginTS, endTS, comment, false, false)
		if err != nil {
			return manualWLInsertedInDB{issueKey, err}
		}

		return manualWLInsertedInDB{issueKey, nil}
	}
}

func deleteActiveIssueLog(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		err := pers.DeleteActiveLogInDB(db)
		return activeWLDeletedFromDB{err}
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
			return wLUpdatedInDB{rowID, issueKey, err}
		}
		defer stmt.Close()

		_, err = stmt.Exec(beginTS.UTC(), endTS.UTC(), comment, rowID)
		if err != nil {
			return wLUpdatedInDB{rowID, issueKey, err}
		}

		return wLUpdatedInDB{rowID, issueKey, nil}
	}
}

func fetchActiveStatus(db *sql.DB, interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		row := db.QueryRow(`
SELECT issue_key, begin_ts, comment
from issue_log
WHERE active=1
ORDER BY begin_ts DESC
LIMIT 1
`)
		var activeIssue string
		var beginTS time.Time
		var comment *string
		err := row.Scan(&activeIssue, &beginTS, &comment)
		if err == sql.ErrNoRows {
			return activeWLFetchedFromDB{activeIssue: activeIssue}
		}
		if err != nil {
			return activeWLFetchedFromDB{err: err}
		}

		return activeWLFetchedFromDB{activeIssue: activeIssue, beginTS: beginTS, comment: comment}
	})
}

func fetchWorkLogs(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		entries, err := pers.FetchWLsFromDB(db)
		return wLEntriesFetchedFromDB{
			entries: entries,
			err:     err,
		}
	}
}

func fetchSyncedWorkLogs(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		entries, err := pers.FetchSyncedWLsFromDB(db)
		return syncedWLEntriesFetchedFromDB{
			entries: entries,
			err:     err,
		}
	}
}

func deleteLogEntry(db *sql.DB, id int) tea.Cmd {
	return func() tea.Msg {
		err := pers.DeleteWLInDB(db, id)
		return wLDeletedFromDB{
			err: err,
		}
	}
}

func updateSyncStatusForEntry(db *sql.DB, entry common.WorklogEntry, index int, fallbackCommentUsed bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		var comment string
		if entry.Comment != nil {
			comment = *entry.Comment
		}
		if fallbackCommentUsed {
			err = pers.UpdateSyncStatusAndCommentForWLInDB(db, entry.ID, comment)
		} else {
			err = pers.UpdateSyncStatusForWLInDB(db, entry.ID)
		}

		return wLSyncUpdatedInDB{
			entry: entry,
			index: index,
			err:   err,
		}
	}
}

func fetchJIRAIssues(cl *jira.Client, jql string) tea.Cmd {
	return func() tea.Msg {
		jIssues, statusCode, err := getIssues(cl, jql)
		var issues []common.Issue
		if err != nil {
			return issuesFetchedFromJIRA{issues, statusCode, err}
		}

		for _, issue := range jIssues {
			var assignee string
			var totalSecsSpent int
			var status string
			if issue.Fields != nil {
				if issue.Fields.Assignee != nil {
					assignee = issue.Fields.Assignee.DisplayName
				}

				totalSecsSpent = issue.Fields.AggregateTimeSpent

				if issue.Fields.Status != nil {
					status = issue.Fields.Status.Name
				}
			}
			issues = append(issues, common.Issue{
				IssueKey:        issue.Key,
				IssueType:       issue.Fields.Type.Name,
				Summary:         issue.Fields.Summary,
				Assignee:        assignee,
				Status:          status,
				AggSecondsSpent: totalSecsSpent,
				TrackingActive:  false,
			})
		}
		return issuesFetchedFromJIRA{issues, statusCode, nil}
	}
}

func syncWorklogWithJIRA(cl *jira.Client, entry common.WorklogEntry, fallbackComment *string, index int, timeDeltaMins int) tea.Cmd {
	return func() tea.Msg {
		var fallbackCmtUsed bool
		if entry.EndTS == nil {
			return wLSyncedToJIRA{index, entry, fallbackCmtUsed, errWorklogsEndTSIsEmpty}
		}

		var comment string
		if entry.NeedsComment() && fallbackComment != nil {
			comment = *fallbackComment
			fallbackCmtUsed = true
		} else if entry.Comment != nil {
			comment = *entry.Comment
		}

		err := syncWLToJIRA(cl, entry.IssueKey, entry.BeginTS, *entry.EndTS, comment, timeDeltaMins)
		return wLSyncedToJIRA{index, entry, fallbackCmtUsed, err}
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
