package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const useHighPerformanceRenderer = false

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.message = ""

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.issueList.FilterState() == list.Filtering {
			m.issueList, cmd = m.issueList.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.activeView {
			case AskForCommentView:

				beginTS, err := time.ParseInLocation(string(timeFormat), m.trackingInputs[entryBeginTS].Value(), time.Local)
				if err != nil {
					m.message = err.Error()
					return m, tea.Batch(cmds...)
				}
				m.activeIssueBeginTS = beginTS.Local()

				endTS, err := time.ParseInLocation(string(timeFormat), m.trackingInputs[entryEndTS].Value(), time.Local)

				if err != nil {
					m.message = err.Error()
					return m, tea.Batch(cmds...)
				}
				m.activeIssueEndTS = endTS.Local()

				if m.trackingInputs[entryComment].Value() == "" {
					m.message = "Comment cannot be empty"
					return m, tea.Batch(cmds...)
				}

				if m.activeIssueEndTS.Sub(m.activeIssueBeginTS).Seconds() <= 0 {
					m.message = "time spent needs to be greater than zero"
					return m, tea.Batch(cmds...)
				}

				cmds = append(cmds, toggleTracking(m.db,
					m.activeIssue,
					m.activeIssueBeginTS,
					m.activeIssueEndTS,
					m.trackingInputs[entryComment].Value(),
				))

				m.activeView = IssueListView
				for i := range m.trackingInputs {
					m.trackingInputs[i].SetValue("")
				}
				return m, tea.Batch(cmds...)
			case ManualWorklogEntryView:
				beginTS, err := time.ParseInLocation(string(timeFormat), m.trackingInputs[entryBeginTS].Value(), time.Local)
				if err != nil {
					m.message = err.Error()
					return m, tea.Batch(cmds...)
				}
				beginTS = beginTS.Local()

				endTS, err := time.ParseInLocation(string(timeFormat), m.trackingInputs[entryEndTS].Value(), time.Local)

				if err != nil {
					m.message = err.Error()
					return m, tea.Batch(cmds...)
				}
				endTS = endTS.Local()

				if m.trackingInputs[entryComment].Value() == "" {
					m.message = "Comment cannot be empty"
					return m, tea.Batch(cmds...)
				}

				if endTS.Sub(beginTS).Seconds() <= 0 {
					m.message = "time spent needs to be greater than zero"
					return m, tea.Batch(cmds...)
				}

				issue, ok := m.issueList.SelectedItem().(*Issue)

				if ok {
					switch m.worklogSaveType {
					case worklogInsert:
						cmds = append(cmds, insertManualEntry(m.db,
							issue.issueKey,
							beginTS,
							endTS,
							m.trackingInputs[entryComment].Value(),
						))
						m.activeView = IssueListView
					case worklogUpdate:
						wl, ok := m.worklogList.SelectedItem().(WorklogEntry)
						if ok {
							cmds = append(cmds, updateManualEntry(m.db,
								wl.Id,
								wl.IssueKey,
								beginTS,
								endTS,
								m.trackingInputs[entryComment].Value(),
							))
							m.activeView = WorklogView
						}
					}
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
			case ManualWorklogEntryView:
				switch m.worklogSaveType {
				case worklogInsert:
					m.activeView = IssueListView
				case worklogUpdate:
					m.activeView = WorklogView
				}
				for i := range m.trackingInputs {
					m.trackingInputs[i].SetValue("")
				}
			}
		case "tab":
			switch m.activeView {
			case IssueListView:
				m.activeView = WorklogView
				cmds = append(cmds, fetchLogEntries(m.db))
			case WorklogView:
				m.activeView = SyncedWorklogView
				cmds = append(cmds, fetchSyncedLogEntries(m.db))
			case SyncedWorklogView:
				m.activeView = IssueListView
			case AskForCommentView, ManualWorklogEntryView:
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
			case WorklogView:
				m.activeView = IssueListView
			case SyncedWorklogView:
				m.activeView = WorklogView
				cmds = append(cmds, fetchLogEntries(m.db))
			case IssueListView:
				m.activeView = SyncedWorklogView
				cmds = append(cmds, fetchSyncedLogEntries(m.db))
			case AskForCommentView, ManualWorklogEntryView:
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
		case "k":
			err := m.shiftTime(shiftBackward, shiftMinute)
			if err != nil {
				return m, tea.Batch(cmds...)
			}
		case "j":
			err := m.shiftTime(shiftForward, shiftMinute)
			if err != nil {
				return m, tea.Batch(cmds...)
			}
		case "K":
			err := m.shiftTime(shiftBackward, shiftFiveMinutes)
			if err != nil {
				return m, tea.Batch(cmds...)
			}
		case "J":
			err := m.shiftTime(shiftForward, shiftFiveMinutes)
			if err != nil {
				return m, tea.Batch(cmds...)
			}
		}
	}

	switch m.activeView {
	case AskForCommentView, ManualWorklogEntryView:
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
				m.activeView = m.lastView
			default:
				return m, tea.Quit
			}
		case "1":
			if m.activeView != IssueListView {
				m.activeView = IssueListView
			}
		case "2":
			if m.activeView != WorklogView {
				m.activeView = WorklogView
				cmds = append(cmds, fetchLogEntries(m.db))
			}
		case "3":
			if m.activeView != SyncedWorklogView {
				m.activeView = SyncedWorklogView
			}
		case "ctrl+r":
			switch m.activeView {
			case IssueListView:
				m.issueList.Title = "fetching..."
				m.issueList.Styles.Title = m.issueList.Styles.Title.Background(lipgloss.Color(issueListUnfetchedColor))
				cmds = append(cmds, fetchJIRAIssues(m.jiraClient, m.jql))
			case WorklogView:
				cmds = append(cmds, fetchLogEntries(m.db))
				m.worklogList.ResetSelected()
			case SyncedWorklogView:
				cmds = append(cmds, fetchSyncedLogEntries(m.db))
				m.syncedWorklogList.ResetSelected()
			}
		case "ctrl+t":
			if m.activeView == IssueListView {
				if m.trackingActive {
					if m.issueList.IsFiltered() {
						m.issueList.ResetFilter()
					}
					activeIndex, ok := m.issueIndexMap[m.activeIssue]
					if ok {
						m.issueList.Select(activeIndex)
					}
				} else {
					m.message = "Nothing is being tracked right now"
				}
			}
		case "ctrl+s":
			if !m.issuesFetched {
				break
			}

			if m.activeView == IssueListView && !m.trackingActive {
				m.activeView = ManualWorklogEntryView
				m.worklogSaveType = worklogInsert
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
			} else if m.activeView == WorklogView {
				wl, ok := m.worklogList.SelectedItem().(WorklogEntry)
				if ok {
					m.activeView = ManualWorklogEntryView
					m.worklogSaveType = worklogUpdate
					m.trackingFocussedField = entryBeginTS

					beginTSStr := wl.BeginTS.Format(timeFormat)
					endTSStr := wl.EndTS.Format(timeFormat)

					m.trackingInputs[entryBeginTS].SetValue(beginTSStr)
					m.trackingInputs[entryEndTS].SetValue(endTSStr)
					m.trackingInputs[entryComment].SetValue(wl.Comment)

					for i := range m.trackingInputs {
						m.trackingInputs[i].Blur()
					}
					m.trackingInputs[m.trackingFocussedField].Focus()
				}
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
		case "ctrl+x":
			if m.activeView == IssueListView && m.trackingActive {
				cmds = append(cmds, deleteActiveIssueLog(m.db))
			}
		case "s":
			if !m.issuesFetched {
				break
			}

			switch m.activeView {
			case IssueListView:
				if m.issueList.FilterState() != list.Filtering {
					if m.changesLocked {
						message := "Changes locked momentarily"
						m.message = message
						m.messages = append(m.messages, message)
					}
					issue, ok := m.issueList.SelectedItem().(*Issue)
					if !ok {
						message := "Something went horribly wrong"
						m.message = message
						m.messages = append(m.messages, message)
					} else {
						if m.lastChange == UpdateChange {
							m.changesLocked = true
							m.activeIssueBeginTS = time.Now()
							cmds = append(cmds, toggleTracking(m.db,
								issue.issueKey,
								m.activeIssueBeginTS,
								m.activeIssueEndTS,
								"",
							))
						} else if m.lastChange == InsertChange {

							currentTime := time.Now()
							beginTimeStr := m.activeIssueBeginTS.Format(timeFormat)
							currentTimeStr := currentTime.Format(timeFormat)

							m.trackingInputs[entryBeginTS].SetValue(beginTimeStr)
							m.trackingInputs[entryEndTS].SetValue(currentTimeStr)

							for i := range m.trackingInputs {
								m.trackingInputs[i].Blur()
							}

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
			}
		case "?":
			if m.activeView == IssueListView || m.activeView == WorklogView || m.activeView == SyncedWorklogView {
				m.lastView = m.activeView
				m.activeView = HelpView
			}
		case "ctrl+b":
			if !m.issuesFetched {
				break
			}

			if m.activeView == IssueListView {
				selectedIssue := m.issueList.SelectedItem().FilterValue()
				cmds = append(cmds, openURLInBrowser(fmt.Sprintf("%sbrowse/%s", m.jiraClient.BaseURL.String(), selectedIssue)))
			}
		}

	case tea.WindowSizeMsg:
		w, h := listStyle.GetFrameSize()
		m.terminalHeight = msg.Height

		m.issueList.SetWidth(msg.Width - w)
		m.worklogList.SetWidth(msg.Width - w)
		m.syncedWorklogList.SetWidth(msg.Width - w)
		m.issueList.SetHeight(msg.Height - h - 2)
		m.worklogList.SetHeight(msg.Height - h - 2)
		m.syncedWorklogList.SetHeight(msg.Height - h - 2)

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
			for i, issue := range msg.issues {
				issue.setDesc()
				issues = append(issues, &issue)
				m.issueMap[issue.issueKey] = &issue
				m.issueIndexMap[issue.issueKey] = i
			}
			m.issueList.SetItems(issues)
			m.issueList.Title = "Issues"
			m.issueList.Styles.Title = m.issueList.Styles.Title.Background(lipgloss.Color(issueListColor))
			m.issuesFetched = true

			cmds = append(cmds, fetchActiveStatus(m.db, 0))
		}
	case ManualEntryInserted:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = "Error inserting worklog: " + message
			m.messages = append(m.messages, message)
		} else {
			for i := range m.trackingInputs {
				m.trackingInputs[i].SetValue("")
			}
			m.unsyncedWLCount++
		}
	case ManualEntryUpdated:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = "Error updating worklog: " + message
			m.messages = append(m.messages, message)
		} else {
			m.message = "Worklog updated"
			for i := range m.trackingInputs {
				m.trackingInputs[i].SetValue("")
			}
			cmds = append(cmds, fetchLogEntries(m.db))
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
			m.unsyncedWLCount = uint(len(msg.entries))
		}
	case SyncedLogEntriesFetchedMsg:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = "Error fetching synced worklog entries: " + message
			m.messages = append(m.messages, message)
		} else {
			var items []list.Item
			for _, e := range msg.entries {
				items = append(items, list.Item(e))
			}
			m.syncedWorklogList.SetItems(items)
		}
	case LogEntrySyncUpdated:
		if msg.err != nil {
			msg.entry.Error = msg.err
			m.messages = append(m.messages, msg.err.Error())
			m.worklogList.SetItem(msg.index, msg.entry)
		} else {
			m.unsyncedWLCount--
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
				activeIssue, ok := m.issueMap[m.activeIssue]
				m.activeIssueBeginTS = msg.beginTs
				if ok {
					activeIssue.trackingActive = true

					// go to tracked item on startup
					activeIndex, ok := m.issueIndexMap[msg.activeIssue]
					if ok {
						m.issueList.Select(activeIndex)
					}
				}
				m.trackingActive = true
			}
		}
	case LogEntriesDeletedMsg:
		if msg.err != nil {
			message := "error deleting entry: " + msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			cmds = append(cmds, fetchLogEntries(m.db))
			m.unsyncedWLCount--
		}
	case activeTaskLogDeletedMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Error deleting active log entry: %s", msg.err)
		} else {
			activeIssue, ok := m.issueMap[m.activeIssue]
			if ok {
				activeIssue.trackingActive = false
			}
			m.lastChange = UpdateChange
			m.trackingActive = false
			m.activeIssue = ""
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
			m.trackingActive = false
		} else {
			var activeIssue *Issue
			if msg.activeIssue != "" {
				activeIssue = m.issueMap[msg.activeIssue]
			} else {
				activeIssue = m.issueMap[m.activeIssue]
			}
			m.changesLocked = false
			if msg.finished {
				m.lastChange = UpdateChange
				if activeIssue != nil {
					activeIssue.trackingActive = false
				}
				m.trackingActive = false
				m.unsyncedWLCount++
			} else {
				m.lastChange = InsertChange
				if activeIssue != nil {
					activeIssue.trackingActive = true
				}
				m.trackingActive = true
			}
			m.activeIssue = msg.activeIssue
		}
	case HideHelpMsg:
		m.showHelpIndicator = false
	case URLOpenedinBrowserMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Error opening url: %s", msg.err.Error())
		}
	}

	switch m.activeView {
	case IssueListView:
		m.issueList, cmd = m.issueList.Update(msg)
		cmds = append(cmds, cmd)
	case WorklogView:
		m.worklogList, cmd = m.worklogList.Update(msg)
		cmds = append(cmds, cmd)
	case SyncedWorklogView:
		m.syncedWorklogList, cmd = m.syncedWorklogList.Update(msg)
		cmds = append(cmds, cmd)
	case HelpView:
		m.helpVP, cmd = m.helpVP.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) shiftTime(direction timeShiftDirection, duration timeShiftDuration) error {
	if m.activeView == AskForCommentView || m.activeView == ManualWorklogEntryView {
		if m.trackingFocussedField == entryBeginTS || m.trackingFocussedField == entryEndTS {
			ts, err := time.ParseInLocation(string(timeFormat), m.trackingInputs[m.trackingFocussedField].Value(), time.Local)
			if err != nil {
				return err
			}

			newTs := getShiftedTime(ts, direction, duration)

			m.trackingInputs[m.trackingFocussedField].SetValue(newTs.Format(timeFormat))
		}
	}
	return nil
}
