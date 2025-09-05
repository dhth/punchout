package ui

import (
	"context"
	"database/sql"
	"errors"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	d "github.com/dhth/punchout/internal/domain"
	pers "github.com/dhth/punchout/internal/persistence"

	_ "modernc.org/sqlite" // sqlite driver
)

var errWorklogsEndTSIsEmpty = errors.New("worklog's end timestamp is empty")

func toggleTracking(db *sql.DB, selectedIssue string, beginTS, endTS time.Time, comment string) tea.Cmd {
	return func() tea.Msg {
		row := db.QueryRow(`
SELECT
    issue_key
FROM
    issue_log
WHERE
    active = 1
ORDER BY
    begin_ts DESC
LIMIT
    1;
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
			err = pers.InsertNewActiveWLInDB(db, selectedIssue, beginTS)
			if err != nil {
				return trackingToggledInDB{err: err}
			}
			return trackingToggledInDB{activeIssue: selectedIssue}

		default:
			err := pers.UpdateActiveWLInDB(db, activeIssue, comment, beginTS, endTS)
			if err != nil {
				return trackingToggledInDB{err: err}
			}
			return trackingToggledInDB{activeIssue: "", finished: true}
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

func insertManualEntry(db *sql.DB, worklog d.ValidatedWorkLog) tea.Cmd {
	return func() tea.Msg {
		err := pers.InsertManualWLInDB(db, worklog)

		return manualWLInsertedInDB{worklog.IssueKey, err}
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
UPDATE
    issue_log
SET
    begin_ts = ?,
    end_ts = ?,
    COMMENT = ?
WHERE
    ID = ?;
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
SELECT
    issue_key,
    begin_ts,
    COMMENT
FROM
    issue_log
WHERE
    active = 1
ORDER BY
    begin_ts DESC
LIMIT
    1;
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

func fetchUnsyncedWorkLogs(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		entries, err := pers.FetchUnsyncedWLsFromDB(db)
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

func updateSyncStatusForEntry(db *sql.DB, entry d.WorklogEntry, index int, fallbackCommentUsed bool) tea.Cmd {
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

func (m Model) fetchJIRAIssues() tea.Cmd {
	return func() tea.Msg {
		issues, statusCode, err := m.jiraSvc.GetIssues(m.jiraCfg.JQL)

		return issuesFetchedFromJIRA{issues, statusCode, err}
	}
}

func (m Model) syncWorklogWithJIRA(entry d.WorklogEntry, index int) tea.Cmd {
	return func() tea.Msg {
		var fallbackCmtUsed bool
		if entry.EndTS == nil {
			return wLSyncedToJIRA{index, entry, fallbackCmtUsed, errWorklogsEndTSIsEmpty}
		}

		var comment string
		if entry.NeedsComment() && m.jiraCfg.FallbackComment != nil {
			comment = *m.jiraCfg.FallbackComment
			fallbackCmtUsed = true
		} else if entry.Comment != nil {
			comment = *entry.Comment
		}

		err := m.jiraSvc.SyncWLToJIRA(context.TODO(), entry, comment, m.jiraCfg.TimeDeltaMins)
		return wLSyncedToJIRA{index, entry, fallbackCmtUsed, err}
	}
}

func hideHelp(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return hideHelpMsg{}
	})
}

func openURLInBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		var openCmd string
		switch runtime.GOOS {
		case "darwin":
			openCmd = "open"
		default:
			openCmd = "xdg-open"
		}
		c := exec.Command(openCmd, url)
		err := c.Run()

		return urlOpenedinBrowserMsg{url: url, err: err}
	}
}
