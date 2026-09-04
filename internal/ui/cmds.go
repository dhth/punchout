package ui

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"time"

	tea "charm.land/bubbletea/v2"
	d "github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/issuecache"
)

func toggleTracking(
	ctx context.Context,
	store WorklogStore,
	operation trackingToggleOperation,
	selectedIssue string,
	beginTS,
	endTS time.Time,
	comment string,
) tea.Cmd {
	return func() tea.Msg {
		activeWorklog, err := store.ActiveWorklog(ctx)
		if err != nil {
			return trackingToggledInDB{
				activeIssue: selectedIssue,
				operation:   operation,
				err:         err,
			}
		}

		switch operation {
		case trackingToggleStart:
			if activeWorklog != nil {
				return trackingToggledInDB{
					operation:             operation,
					reconcileActiveStatus: true,
					err:                   errors.New("cannot start tracking while another worklog is active"),
				}
			}
			err = store.StartWorklog(ctx, selectedIssue, beginTS)
			if err != nil {
				return trackingToggledInDB{
					operation:             operation,
					reconcileActiveStatus: true,
					err:                   err,
				}
			}
			return trackingToggledInDB{activeIssue: selectedIssue, operation: operation}

		case trackingToggleFinish:
			if activeWorklog == nil {
				return trackingToggledInDB{
					activeIssue:           selectedIssue,
					operation:             operation,
					reconcileActiveStatus: true,
					err:                   errors.New("cannot finish tracking when no worklog is active"),
				}
			}
			err := store.FinishWorklog(ctx,
				d.Worklog{
					IssueKey: activeWorklog.IssueKey,
					BeginTS:  beginTS,
					EndTS:    endTS,
					Comment:  comment,
				},
			)
			if err != nil {
				return trackingToggledInDB{
					activeIssue:           activeWorklog.IssueKey,
					operation:             operation,
					reconcileActiveStatus: true,
					err:                   err,
				}
			}
			return trackingToggledInDB{activeIssue: "", finished: true, operation: operation}

		default:
			return trackingToggledInDB{err: errors.New("unknown tracking operation")}
		}
	}
}

func quickSwitchActiveIssue(ctx context.Context, store WorklogStore, selectedIssue string, currentTime time.Time) tea.Cmd {
	return func() tea.Msg {
		previousIssue, err := store.SwitchActiveWorklog(ctx, selectedIssue, currentTime)
		return activeWLSwitchedInDB{previousIssue, selectedIssue, currentTime, err}
	}
}

func updateActiveWL(ctx context.Context, store WorklogStore, beginTS time.Time, comment *string) tea.Cmd {
	return func() tea.Msg {
		err := store.UpdateActiveWorklog(ctx, beginTS, comment)
		return activeWLUpdatedInDB{beginTS, comment, err}
	}
}

func insertManualEntry(ctx context.Context, store WorklogStore, worklog d.Worklog) tea.Cmd {
	return func() tea.Msg {
		err := store.AddWorklog(ctx, worklog)

		return manualWLInsertedInDB{worklog.IssueKey, err}
	}
}

func deleteActiveIssueLog(ctx context.Context, store WorklogStore) tea.Cmd {
	return func() tea.Msg {
		err := store.DeleteActiveWorklog(ctx)
		return activeWLDeletedFromDB{err}
	}
}

func updateManualEntry(ctx context.Context, store WorklogStore, rowID int, worklog d.Worklog) tea.Cmd {
	return func() tea.Msg {
		err := store.UpdateWorklog(ctx, rowID, worklog)
		return wLUpdatedInDB{rowID, worklog.IssueKey, err}
	}
}

func fetchActiveStatus(ctx context.Context, store WorklogStore, interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		worklog, err := store.ActiveWorklog(ctx)
		if err != nil {
			return activeWLFetchedFromDB{err: err}
		}

		return activeWLFetchedFromDB{worklog: worklog}
	})
}

func fetchUnsyncedWorkLogs(ctx context.Context, store WorklogStore, generation uint64) tea.Cmd {
	return func() tea.Msg {
		entries, err := store.UnsyncedWorklogs(ctx)
		return wLEntriesFetchedFromDB{
			entries:    entries,
			generation: generation,
			err:        err,
		}
	}
}

func fetchSyncedWorkLogs(ctx context.Context, store WorklogStore) tea.Cmd {
	return func() tea.Msg {
		entries, err := store.SyncedWorklogs(ctx)
		return syncedWLEntriesFetchedFromDB{
			entries: entries,
			err:     err,
		}
	}
}

func deleteLogEntry(ctx context.Context, store WorklogStore, id int) tea.Cmd {
	return func() tea.Msg {
		err := store.DeleteWorklog(ctx, id)
		return wLDeletedFromDB{
			err: err,
		}
	}
}

func updateSyncStatusForEntry(ctx context.Context, store WorklogStore, worklog d.StoredWorklog, indexHint int, fallbackCommentUsed bool) tea.Cmd {
	return func() tea.Msg {
		var comment *string
		if fallbackCommentUsed {
			comment = &worklog.Comment
		}

		err := store.MarkWorklogSynced(ctx, worklog.ID, comment)
		return wLSyncUpdatedInDB{
			worklog:   worklog,
			indexHint: indexHint,
			err:       err,
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

func (m Model) syncWorklogWithJIRA(worklog d.StoredWorklog, indexHint int) tea.Cmd {
	return func() tea.Msg {
		var fallbackCmtUsed bool
		if worklog.NeedsComment() && m.opts.Jira.FallbackComment != nil {
			worklog.Comment = *m.opts.Jira.FallbackComment
			fallbackCmtUsed = true
		}

		err := m.jiraSvc.SyncWLToJIRA(m.ctx, worklog.Worklog, m.opts.Jira.TimeDeltaMins)
		return wLSyncedToJIRA{
			indexHint:           indexHint,
			worklog:             worklog,
			fallbackCommentUsed: fallbackCmtUsed,
			err:                 err,
		}
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
