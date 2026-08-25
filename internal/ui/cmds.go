package ui

import (
	"context"
	"database/sql"
	"errors"
	"os/exec"
	"runtime"
	"time"

	tea "charm.land/bubbletea/v2"
	d "github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/issuecache"
	pers "github.com/dhth/punchout/internal/persistence"

	_ "modernc.org/sqlite" // sqlite driver
)

func toggleTracking(db *sql.DB, operation trackingToggleOperation, selectedIssue string, beginTS, endTS time.Time, comment string) tea.Cmd {
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
			return trackingToggledInDB{activeIssue: selectedIssue, operation: operation, err: err}
		} else {
			trackStatus = trackingActive
		}

		switch operation {
		case trackingToggleStart:
			if trackStatus != trackingInactive {
				return trackingToggledInDB{operation: operation, reconcileActiveStatus: true, err: errors.New("cannot start tracking while another worklog is active")}
			}
			err = pers.InsertNewActiveWLInDB(db, selectedIssue, beginTS)
			if err != nil {
				return trackingToggledInDB{operation: operation, err: err}
			}
			return trackingToggledInDB{activeIssue: selectedIssue, operation: operation}

		case trackingToggleFinish:
			if trackStatus != trackingActive {
				return trackingToggledInDB{activeIssue: selectedIssue, operation: operation, reconcileActiveStatus: true, err: errors.New("cannot finish tracking when no worklog is active")}
			}
			err := pers.UpdateActiveWLInDB(db, d.Worklog{IssueKey: activeIssue, BeginTS: beginTS, EndTS: endTS, Comment: comment})
			if err != nil {
				return trackingToggledInDB{activeIssue: activeIssue, operation: operation, err: err}
			}
			return trackingToggledInDB{activeIssue: "", finished: true, operation: operation}

		default:
			return trackingToggledInDB{err: errors.New("unknown tracking operation")}
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

func insertManualEntry(ctx context.Context, store WorklogStore, worklog d.Worklog) tea.Cmd {
	return func() tea.Msg {
		err := store.AddWorklog(ctx, worklog)

		return manualWLInsertedInDB{worklog.IssueKey, err}
	}
}

func deleteActiveIssueLog(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		err := pers.DeleteActiveLogInDB(db)
		return activeWLDeletedFromDB{err}
	}
}

func updateManualEntry(db *sql.DB, rowID int, worklog d.Worklog) tea.Cmd {
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
			return wLUpdatedInDB{rowID, worklog.IssueKey, err}
		}
		defer stmt.Close()

		_, err = stmt.Exec(worklog.BeginTS.UTC(), worklog.EndTS.UTC(), worklog.Comment, rowID)
		if err != nil {
			return wLUpdatedInDB{rowID, worklog.IssueKey, err}
		}

		return wLUpdatedInDB{rowID, worklog.IssueKey, nil}
	}
}

func fetchActiveStatus(db *sql.DB, interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		worklog, err := pers.FetchActiveWLFromDB(db)
		if err != nil {
			return activeWLFetchedFromDB{err: err}
		}

		return activeWLFetchedFromDB{worklog: worklog}
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

func updateSyncStatusForEntry(db *sql.DB, entry worklogListItem, index int, fallbackCommentUsed bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if fallbackCommentUsed {
			err = pers.UpdateSyncStatusAndCommentForWLInDB(db, entry.ID, entry.Comment)
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

func (m Model) fetchIssuesFromJIRA(afterCacheLoadFailure bool) tea.Cmd {
	return func() tea.Msg {
		issues, err := m.jiraSvc.GetIssues(m.opts.Jira.JQL)
		if err != nil {
			return issuesLoaded{
				source:                issueSourceJIRA,
				afterCacheLoadFailure: afterCacheLoadFailure,
				err:                   err,
			}
		}

		return issuesLoaded{
			issues:                issues,
			fetchedAt:             time.Now().UTC(),
			source:                issueSourceJIRA,
			afterCacheLoadFailure: afterCacheLoadFailure,
		}
	}
}

func (m Model) loadIssuesFromCache() tea.Cmd {
	return func() tea.Msg {
		snapshot, err := m.issueStore.Load()
		if err != nil {
			return issuesLoaded{
				source: issueSourceCache,
				err:    err,
			}
		}

		return issuesLoaded{
			issues:    snapshot.Issues,
			fetchedAt: snapshot.FetchedAt,
			source:    issueSourceCache,
		}
	}
}

func (m Model) saveIssuesToCache(snapshot issuecache.Snapshot) tea.Cmd {
	return func() tea.Msg {
		return issuesSavedToCache{err: m.issueStore.Save(snapshot)}
	}
}

func (m Model) syncWorklogWithJIRA(entry worklogListItem, index int) tea.Cmd {
	return func() tea.Msg {
		var fallbackCmtUsed bool
		worklog := entry.Worklog
		if worklog.NeedsComment() && m.opts.Jira.FallbackComment != nil {
			worklog.Comment = *m.opts.Jira.FallbackComment
			fallbackCmtUsed = true
		}

		err := m.jiraSvc.SyncWLToJIRA(context.TODO(), worklog, m.opts.Jira.TimeDeltaMins)
		return wLSyncedToJIRA{index, entry, fallbackCmtUsed, err}
	}
}

func hideHelp(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return hideHelpMsg{}
	})
}

func clearUserMsgAfter(message userMsg) tea.Cmd {
	return tea.Tick(message.duration(), func(time.Time) tea.Msg {
		return clearUserMsgMsg{id: message.id}
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
