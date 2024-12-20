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
			case askForCommentView:

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

				m.activeView = issueListView
				for i := range m.trackingInputs {
					m.trackingInputs[i].SetValue("")
				}
				return m, tea.Batch(cmds...)
			case manualWorklogEntryView:
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
						m.activeView = issueListView
					case worklogUpdate:
						wl, ok := m.worklogList.SelectedItem().(worklogEntry)
						if ok {
							cmds = append(cmds, updateManualEntry(m.db,
								wl.Id,
								wl.IssueKey,
								beginTS,
								endTS,
								m.trackingInputs[entryComment].Value(),
							))
							m.activeView = worklogView
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
			case askForCommentView:
				m.activeView = issueListView
				m.trackingInputs[entryComment].SetValue("")
			case manualWorklogEntryView:
				switch m.worklogSaveType {
				case worklogInsert:
					m.activeView = issueListView
				case worklogUpdate:
					m.activeView = worklogView
				}
				for i := range m.trackingInputs {
					m.trackingInputs[i].SetValue("")
				}
			}
		case "tab":
			switch m.activeView {
			case issueListView:
				m.activeView = worklogView
				cmds = append(cmds, fetchLogEntries(m.db))
			case worklogView:
				m.activeView = syncedWorklogView
				cmds = append(cmds, fetchSyncedLogEntries(m.db))
			case syncedWorklogView:
				m.activeView = issueListView
			case askForCommentView, manualWorklogEntryView:
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
			case worklogView:
				m.activeView = issueListView
			case syncedWorklogView:
				m.activeView = worklogView
				cmds = append(cmds, fetchLogEntries(m.db))
			case issueListView:
				m.activeView = syncedWorklogView
				cmds = append(cmds, fetchSyncedLogEntries(m.db))
			case askForCommentView, manualWorklogEntryView:
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
	case askForCommentView, manualWorklogEntryView:
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
			case issueListView:
				fs := m.issueList.FilterState()
				if fs == list.Filtering || fs == list.FilterApplied {
					m.issueList.ResetFilter()
				} else {
					return m, tea.Quit
				}
			case worklogView:
				fs := m.worklogList.FilterState()
				if fs == list.Filtering || fs == list.FilterApplied {
					m.worklogList.ResetFilter()
				} else {
					return m, tea.Quit
				}
			case helpView:
				m.activeView = m.lastView
			default:
				return m, tea.Quit
			}
		case "1":
			if m.activeView != issueListView {
				m.activeView = issueListView
			}
		case "2":
			if m.activeView != worklogView {
				m.activeView = worklogView
				cmds = append(cmds, fetchLogEntries(m.db))
			}
		case "3":
			if m.activeView != syncedWorklogView {
				m.activeView = syncedWorklogView
			}
		case "ctrl+r":
			switch m.activeView {
			case issueListView:
				m.issueList.Title = "fetching..."
				m.issueList.Styles.Title = m.issueList.Styles.Title.Background(lipgloss.Color(issueListUnfetchedColor))
				cmds = append(cmds, fetchJIRAIssues(m.jiraClient, m.jql))
			case worklogView:
				cmds = append(cmds, fetchLogEntries(m.db))
				m.worklogList.ResetSelected()
			case syncedWorklogView:
				cmds = append(cmds, fetchSyncedLogEntries(m.db))
				m.syncedWorklogList.ResetSelected()
			}
		case "ctrl+t":
			if m.activeView == issueListView {
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

			if m.activeView == issueListView && !m.trackingActive {
				m.activeView = manualWorklogEntryView
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
			} else if m.activeView == worklogView {
				wl, ok := m.worklogList.SelectedItem().(worklogEntry)
				if ok {
					m.activeView = manualWorklogEntryView
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
			case worklogView:
				issue, ok := m.worklogList.SelectedItem().(worklogEntry)
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
			if m.activeView == issueListView && m.trackingActive {
				cmds = append(cmds, deleteActiveIssueLog(m.db))
			}
		case "s":
			if !m.issuesFetched {
				break
			}

			switch m.activeView {
			case issueListView:
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
						if m.lastChange == updateChange {
							m.changesLocked = true
							m.activeIssueBeginTS = time.Now()
							cmds = append(cmds, toggleTracking(m.db,
								issue.issueKey,
								m.activeIssueBeginTS,
								m.activeIssueEndTS,
								"",
							))
						} else if m.lastChange == insertChange {

							currentTime := time.Now()
							beginTimeStr := m.activeIssueBeginTS.Format(timeFormat)
							currentTimeStr := currentTime.Format(timeFormat)

							m.trackingInputs[entryBeginTS].SetValue(beginTimeStr)
							m.trackingInputs[entryEndTS].SetValue(currentTimeStr)

							for i := range m.trackingInputs {
								m.trackingInputs[i].Blur()
							}

							m.activeView = askForCommentView
							m.trackingFocussedField = entryComment
							m.trackingInputs[m.trackingFocussedField].Focus()
						}
					}
				}
			case worklogView:
				for i, entry := range m.worklogList.Items() {
					if wl, ok := entry.(worklogEntry); ok {
						if !wl.Synced {
							wl.SyncInProgress = true
							m.worklogList.SetItem(i, wl)
							cmds = append(cmds, syncWorklogWithJIRA(m.jiraClient, wl, i, m.jiraTimeDeltaMins))
						}
					}
				}
			}
		case "?":
			if m.activeView == issueListView || m.activeView == worklogView || m.activeView == syncedWorklogView {
				m.lastView = m.activeView
				m.activeView = helpView
			}
		case "ctrl+b":
			if !m.issuesFetched {
				break
			}

			if m.activeView == issueListView {
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
			m.helpVP.SetContent(helpText)
			m.helpVPReady = true
		} else {
			m.helpVP.Height = m.terminalHeight - 7
			m.helpVP.Width = w - 5

		}
	case issuesFetchedFromJIRAMsg:
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
	case manualEntryInserted:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = "Error inserting worklog: " + message
			m.messages = append(m.messages, message)
		} else {
			for i := range m.trackingInputs {
				m.trackingInputs[i].SetValue("")
			}
			cmds = append(cmds, fetchLogEntries(m.db))
		}
	case manualEntryUpdated:
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
	case logEntriesFetchedMsg:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			var items []list.Item
			var secsSpent int
			for _, e := range msg.entries {
				secsSpent += e.SecsSpent()
				items = append(items, list.Item(e))
			}
			m.worklogList.SetItems(items)
			m.unsyncedWLSecsSpent = secsSpent
			m.unsyncedWLCount = uint(len(msg.entries))
			if m.debug {
				m.message = "[io: log entries]"
			}
		}
	case syncedLogEntriesFetchedMsg:
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
	case logEntrySyncUpdated:
		if msg.err != nil {
			msg.entry.Error = msg.err
			m.messages = append(m.messages, msg.err.Error())
			m.worklogList.SetItem(msg.index, msg.entry)
		} else {
			m.unsyncedWLCount--
			m.unsyncedWLSecsSpent -= msg.entry.SecsSpent()
		}
	case fetchActiveMsg:
		if msg.err != nil {
			message := msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			m.activeIssue = msg.activeIssue
			if msg.activeIssue == "" {
				m.lastChange = updateChange
			} else {
				m.lastChange = insertChange
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
	case logEntriesDeletedMsg:
		if msg.err != nil {
			message := "error deleting entry: " + msg.err.Error()
			m.message = message
			m.messages = append(m.messages, message)
		} else {
			cmds = append(cmds, fetchLogEntries(m.db))
		}
	case activeTaskLogDeletedMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Error deleting active log entry: %s", msg.err)
		} else {
			activeIssue, ok := m.issueMap[m.activeIssue]
			if ok {
				activeIssue.trackingActive = false
			}
			m.lastChange = updateChange
			m.trackingActive = false
			m.activeIssue = ""
		}
	case wlAddedOnJIRA:
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
	case trackingToggledMsg:
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
				m.lastChange = updateChange
				if activeIssue != nil {
					activeIssue.trackingActive = false
				}
				m.trackingActive = false
				cmds = append(cmds, fetchLogEntries(m.db))
			} else {
				m.lastChange = insertChange
				if activeIssue != nil {
					activeIssue.trackingActive = true
				}
				m.trackingActive = true
			}
			m.activeIssue = msg.activeIssue
		}
	case hideHelpMsg:
		m.showHelpIndicator = false
	case urlOpenedinBrowserMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Error opening url: %s", msg.err.Error())
		}
	}

	switch m.activeView {
	case issueListView:
		m.issueList, cmd = m.issueList.Update(msg)
		cmds = append(cmds, cmd)
	case worklogView:
		m.worklogList, cmd = m.worklogList.Update(msg)
		cmds = append(cmds, cmd)
	case syncedWorklogView:
		m.syncedWorklogList, cmd = m.syncedWorklogList.Update(msg)
		cmds = append(cmds, cmd)
	case helpView:
		m.helpVP, cmd = m.helpVP.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) shiftTime(direction timeShiftDirection, duration timeShiftDuration) error {
	if m.activeView == askForCommentView || m.activeView == manualWorklogEntryView {
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
