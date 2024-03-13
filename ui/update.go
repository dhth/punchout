package ui

import (
	"log"
	"time"

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
			switch m.activeView {
			case AskForCommentView:
				m.activeView = IssueListView
				if m.trackingInputs[entryComment].Value() != "" {
					cmds = append(cmds, toggleTracking(m.db, m.activeIssue, m.trackingInputs[entryComment].Value()))
					m.trackingInputs[entryComment].SetValue("")
					return m, tea.Batch(cmds...)
				}
			case EntryTrackingView:
				m.activeView = IssueListView

				issue, ok := m.issueList.SelectedItem().(Issue)
				beginTS, err := time.ParseInLocation(string(timeFormat), m.trackingInputs[entryBeginTS].Value(), time.Local)
				if err != nil {
					m.errorMessage = err.Error()
					log.Println(err.Error())
					return m, tea.Batch(cmds...)
				}

				endTS, err := time.ParseInLocation(string(timeFormat), m.trackingInputs[entryEndTS].Value(), time.Local)

				if err != nil {
					m.errorMessage = err.Error()
					return m, tea.Batch(cmds...)
				}

				comment := m.trackingInputs[entryComment].Value()

				if len(comment) == 0 {
					m.errorMessage = "Comment cannot be empty"
					return m, tea.Batch(cmds...)

				}

				if ok {
					cmds = append(cmds, insertManualEntry(m.db, issue.IssueKey, beginTS.Local(), endTS.Local(), comment))
				}

				for i := range m.trackingInputs {
					m.trackingInputs[i].SetValue("")
				}
				return m, tea.Batch(cmds...)
			}
		case "esc":
			switch m.activeView {
			case AskForCommentView:
				m.activeView = IssueListView
				m.trackingInputs[entryComment].SetValue("")
			case EntryTrackingView:
				m.activeView = IssueListView
				for i := range m.trackingInputs {
					m.trackingInputs[i].SetValue("")
				}
			}
		case "tab":
			switch m.activeView {
			case EntryTrackingView:
				switch m.trackingFocussedField {
				case entryBeginTS:
					m.trackingFocussedField = entryEndTS
				case entryEndTS:
					m.trackingFocussedField = entryComment
				case entryComment:
					m.trackingFocussedField = entryBeginTS
				}
				for i := range m.trackingInputs {
					m.trackingInputs[i].Blur()
				}
				m.trackingInputs[m.trackingFocussedField].Focus()
			}
		case "shift+tab":
			switch m.activeView {
			case EntryTrackingView:
				switch m.trackingFocussedField {
				case entryBeginTS:
					m.trackingFocussedField = entryComment
				case entryEndTS:
					m.trackingFocussedField = entryBeginTS
				case entryComment:
					m.trackingFocussedField = entryEndTS
				}
				for i := range m.trackingInputs {
					m.trackingInputs[i].Blur()
				}
				m.trackingInputs[m.trackingFocussedField].Focus()
			}
		}
	}

	switch m.activeView {
	case AskForCommentView:
		m.trackingInputs[entryComment], cmd = m.trackingInputs[entryComment].Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case EntryTrackingView:
		for i := range m.trackingInputs {
			m.trackingInputs[i], cmd = m.trackingInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
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
		case "ctrl+s":
			if m.activeView == IssueListView {
				m.activeView = EntryTrackingView
				m.trackingFocussedField = entryBeginTS
				currentTime := time.Now()
				dateString := currentTime.Format("2006/01/02")
				currentTimeStr := currentTime.Format(timeFormat)

				m.trackingInputs[entryBeginTS].SetValue(dateString + " ")
				m.trackingInputs[entryEndTS].SetValue(currentTimeStr)

				for i := range m.trackingInputs {
					m.trackingInputs[i].Blur()
				}
				m.trackingInputs[m.trackingFocussedField].Focus()
				return m, tea.Batch(cmds...)
			}
		case "ctrl+d":
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
				if m.issueList.FilterState() != list.Filtering {
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
							m.trackingFocussedField = entryComment
							m.trackingInputs[m.trackingFocussedField].Focus()
						}
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
		w, h := stackListStyle.GetFrameSize()
		m.terminalHeight = msg.Height
		m.issueList.SetHeight(msg.Height - h - 2)
		m.worklogList.SetHeight(msg.Height - h - 2)
		if !m.helpVPReady {
			m.helpVP = viewport.New(w-5, m.terminalHeight-7)
			m.helpVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.helpVP.SetContent(HelpText)
			m.helpVPReady = true
		} else {
			m.helpVP.Height = m.terminalHeight - 7
			m.helpVP.Width = w - 5

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
	case ManualEntryInserted:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			m.message = "Manual entry saved"
			for i := range m.trackingInputs {
				m.trackingInputs[i].SetValue("")
			}
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
			m.worklogList.SetItem(msg.index, msg.entry)
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
	case HelpView:
		m.helpVP, cmd = m.helpVP.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
