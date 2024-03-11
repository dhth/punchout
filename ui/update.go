package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const useHighPerformanceRenderer = false

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.message = ""
	m.errorMessage = ""

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			if m.activeView == AskForCommentView {
				m.activeView = IssueListView
				if m.commentInput.Value() != "" {
					cmds = append(cmds, toggleTracking(m.db, m.activeIssue, m.commentInput.Value()))
					m.commentInput.SetValue("")
					return m, tea.Batch(cmds...)
				}
			}
		case "esc":
			if m.activeView == AskForCommentView {
				m.activeView = IssueListView
				m.commentInput.SetValue("")
			}
		}
	}

	switch m.activeView {
	case AskForCommentView:
		m.commentInput, cmd = m.commentInput.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			switch m.activeView {
			case IssueListView:
				fs := m.issueList.FilterState()
				if fs == list.Filtering || fs == list.FilterApplied {
					m.issueList.ResetFilter()
				} else {
					return m, tea.Quit
				}
			case WorklogView:
				fs := m.worklogList.FilterState()
				if fs == list.Filtering || fs == list.FilterApplied {
					m.worklogList.ResetFilter()
				} else {
					return m, tea.Quit
				}
			case HelpView:
				m.activeView = IssueListView
			}
		case "1":
			if m.activeView != IssueListView {
				m.activeView = IssueListView
				return m, tea.Batch(cmds...)
			}
		case "2":
			if m.activeView != WorklogView {
				m.activeView = WorklogView
				cmds = append(cmds, fetchLogEntries(m.db))
				return m, tea.Batch(cmds...)
			}
		case "ctrl+r":
			if m.activeView == WorklogView {
				cmds = append(cmds, fetchLogEntries(m.db))
				return m, tea.Batch(cmds...)
			}
		case "d":
			switch m.activeView {
			case WorklogView:
				issue, ok := m.worklogList.SelectedItem().(WorklogEntry)
				if ok {
					cmds = append(cmds, deleteLogEntry(m.db, issue.Id))
					return m, tea.Batch(cmds...)
				} else {
					msg := "Couldn't delete worklog entry"
					m.message = msg
					m.messages = append(m.messages, msg)
				}
			}
		case "s":
			switch m.activeView {
			case IssueListView:
				if m.changesLocked {
					message := "Changes locked momentarily"
					m.message = message
					m.messages = append(m.messages, message)
					return m, tea.Batch(cmds...)
				}
				issue, ok := m.issueList.SelectedItem().(Issue)
				if !ok {
					message := "Something went horribly wrong"
					m.message = message
					m.messages = append(m.messages, message)
				} else {
					if m.lastChange == UpdateChange {
						m.changesLocked = true
						cmds = append(cmds, toggleTracking(m.db, issue.IssueKey, ""))
					} else if m.lastChange == InsertChange {
						m.activeView = AskForCommentView
					}
				}
			case WorklogView:
				for i, entry := range m.worklogList.Items() {
					if wl, ok := entry.(WorklogEntry); ok {
						if !wl.Synced {
							wl.SyncInProgress = true
							m.worklogList.SetItem(i, wl)
							cmds = append(cmds, syncWorklogWithJIRA(m.jiraClient, wl, i, m.jiraTimeDeltaMins))
						}
					}
				}
				return m, tea.Sequence(cmds...)
			}
		case "?":
			m.lastView = m.activeView
			m.activeView = HelpView
			return m, tea.Batch(cmds...)
		}

	case tea.WindowSizeMsg:
		_, h1 := stackListStyle.GetFrameSize()
		m.terminalHeight = msg.Height
		m.issueList.SetHeight(msg.Height - h1 - 2)
		m.worklogList.SetHeight(msg.Height - h1 - 2)
		if !m.helpVPReady {
			m.helpVP = viewport.New(120, m.terminalHeight-7)
			m.helpVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.helpVP.SetContent(HelpText)
			m.helpVPReady = true
		}
	case IssuesFetchedFromJIRAMsg:
		if msg.err != nil {
			message := "error fetching issues from JIRA: " + msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			issues := make([]list.Item, 0, len(msg.issues))
			for _, issue := range msg.issues {
				issues = append(issues, issue)
			}
			m.issueList.SetItems(issues)
			m.issueList.Title = "Issues"
		}
	case InsertEntryMsg:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			m.activeIssue = msg.issueKey
		}
	case LogEntriesFetchedMsg:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			var items []list.Item
			for _, e := range msg.entries {
				items = append(items, list.Item(e))
			}
			m.worklogList.SetItems(items)
		}
	case UpdateEntryMsg:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			m.activeIssue = ""
		}
	case LogEntrySyncUpdated:
		if msg.err != nil {
			msg.entry.Error = msg.err
			m.messages = append(m.messages, msg.err.Error())
			m.worklogList.SetItem(msg.index, msg.entry)
		}
	case FetchActiveMsg:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			m.activeIssue = msg.activeIssue
			if msg.activeIssue == "" {
				m.lastChange = UpdateChange
			} else {
				m.lastChange = InsertChange
			}
		}
	case LogEntriesDeletedMsg:
		if msg.err != nil {
			message := "error deleting entry: " + msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			cmds = append(cmds, fetchLogEntries(m.db))
		}
	case WLAddedOnJIRA:
		if msg.err != nil {
			msg.entry.Error = msg.err
			m.messages = append(m.messages, msg.err.Error())
		} else {
			msg.entry.Synced = true
			msg.entry.SyncInProgress = false
			m.worklogList.SetItem(msg.index, msg.entry)
			cmds = append(cmds, updateSyncStatusForEntry(m.db, msg.entry, msg.index))

		}
	case TrackingToggledMsg:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			m.activeIssue = msg.activeIssue
			m.changesLocked = false
			if msg.finished {
				m.lastChange = UpdateChange
				m.message = "Saved!"
			} else {
				m.lastChange = InsertChange
			}
		}
	case HideHelpMsg:
		m.showHelpIndicator = false
	}

	switch m.activeView {
	case IssueListView:
		m.issueList, cmd = m.issueList.Update(msg)
		cmds = append(cmds, cmd)
	case WorklogView:
		m.worklogList, cmd = m.worklogList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
