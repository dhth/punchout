package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	c "github.com/dhth/punchout/internal/common"
	pers "github.com/dhth/punchout/internal/persistence"
)

func (m *Model) getCmdToUpdateActiveWL() tea.Cmd {
	beginTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryBeginTS].Value(), time.Local)
	if err != nil {
		m.message = err.Error()
		return nil
	}
	commentValue := m.trackingInputs[entryComment].Value()

	var comment *string
	if strings.TrimSpace(commentValue) != "" {
		comment = &commentValue
	}
	m.trackingInputs[entryBeginTS].SetValue("")
	m.activeView = issueListView
	return updateActiveWL(m.db, beginTS, comment)
}

func (m *Model) getCmdToSaveActiveWL() tea.Cmd {
	beginTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryBeginTS].Value(), time.Local)
	if err != nil {
		m.message = err.Error()
		return nil
	}
	m.activeIssueBeginTS = beginTS.Local()

	endTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryEndTS].Value(), time.Local)
	if err != nil {
		m.message = err.Error()
		return nil
	}
	m.activeIssueEndTS = endTS.Local()

	if !m.activeIssueEndTS.After(m.activeIssueBeginTS) {
		return nil
	}

	comment := m.trackingInputs[entryComment].Value()

	m.activeView = issueListView
	for i := range m.trackingInputs {
		m.trackingInputs[i].SetValue("")
	}

	return toggleTracking(m.db,
		m.activeIssue,
		m.activeIssueBeginTS,
		m.activeIssueEndTS,
		comment,
	)
}

func (m *Model) getCmdToSaveOrUpdateWL() tea.Cmd {
	beginTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryBeginTS].Value(), time.Local)
	if err != nil {
		m.message = err.Error()
		return nil
	}

	endTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryEndTS].Value(), time.Local)
	if err != nil {
		m.message = err.Error()
		return nil
	}

	if !endTS.After(beginTS) {
		return nil
	}

	issue, ok := m.issueList.SelectedItem().(*c.Issue)

	var cmd tea.Cmd
	if ok {
		switch m.worklogSaveType {
		case worklogInsert:
			cmd = insertManualEntry(m.db,
				issue.IssueKey,
				beginTS,
				endTS,
				m.trackingInputs[entryComment].Value(),
			)
			m.activeView = issueListView
		case worklogUpdate:
			wl, ok := m.worklogList.SelectedItem().(c.WorklogEntry)
			if ok {
				cmd = updateManualEntry(m.db,
					wl.ID,
					wl.IssueKey,
					beginTS,
					endTS,
					m.trackingInputs[entryComment].Value(),
				)
				m.activeView = wLView
			}
		}
	}
	for i := range m.trackingInputs {
		m.trackingInputs[i].SetValue("")
	}
	return cmd
}

func (m *Model) handleEscape() bool {
	var quit bool

	switch m.activeView {
	case issueListView:
		quit = true
	case wLView:
		quit = true
	case syncedWLView:
		quit = true
	case helpView:
		quit = true
	case editActiveWLView:
		m.activeView = issueListView
	case saveActiveWLView:
		m.activeView = issueListView
		m.trackingInputs[entryComment].SetValue("")
	case wlEntryView:
		switch m.worklogSaveType {
		case worklogInsert:
			m.activeView = issueListView
		case worklogUpdate:
			m.activeView = wLView
		}
		for i := range m.trackingInputs {
			m.trackingInputs[i].SetValue("")
		}
	}

	return quit
}

func (m *Model) getCmdToGoForwardsInViews() tea.Cmd {
	var cmd tea.Cmd
	switch m.activeView {
	case issueListView:
		m.activeView = wLView
		cmd = fetchWorkLogs(m.db)
	case wLView:
		m.activeView = syncedWLView
		cmd = fetchSyncedWorkLogs(m.db)
	case syncedWLView:
		m.activeView = issueListView
	case editActiveWLView:
		switch m.trackingFocussedField {
		case entryBeginTS:
			m.trackingFocussedField = entryComment
		case entryComment:
			m.trackingFocussedField = entryBeginTS
		}
		for i := range m.trackingInputs {
			m.trackingInputs[i].Blur()
		}
		m.trackingInputs[m.trackingFocussedField].Focus()
	case saveActiveWLView, wlEntryView:
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

	return cmd
}

func (m *Model) getCmdToGoBackwardsInViews() tea.Cmd {
	var cmd tea.Cmd
	switch m.activeView {
	case wLView:
		m.activeView = issueListView
	case syncedWLView:
		m.activeView = wLView
		cmd = fetchWorkLogs(m.db)
	case issueListView:
		m.activeView = syncedWLView
		cmd = fetchSyncedWorkLogs(m.db)
	case editActiveWLView:
		switch m.trackingFocussedField {
		case entryBeginTS:
			m.trackingFocussedField = entryComment
		case entryComment:
			m.trackingFocussedField = entryBeginTS
		}
		for i := range m.trackingInputs {
			m.trackingInputs[i].Blur()
		}
		m.trackingInputs[m.trackingFocussedField].Focus()
	case saveActiveWLView, wlEntryView:
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

	return cmd
}

func (m *Model) handleRequestToGoBackOrQuit() bool {
	var quit bool
	switch m.activeView {
	case issueListView:
		fs := m.issueList.FilterState()
		if fs == list.Filtering || fs == list.FilterApplied {
			m.issueList.ResetFilter()
		} else {
			quit = true
		}
	case wLView:
		fs := m.worklogList.FilterState()
		if fs == list.Filtering || fs == list.FilterApplied {
			m.worklogList.ResetFilter()
		} else {
			m.activeView = issueListView
		}
	case syncedWLView:
		m.activeView = wLView
	case helpView:
		m.activeView = m.lastView
	default:
		quit = true
	}

	return quit
}

func (m *Model) getCmdToReloadData() tea.Cmd {
	var cmd tea.Cmd
	switch m.activeView {
	case issueListView:
		m.issueList.Title = "fetching..."
		m.issueList.Styles.Title = m.issueList.Styles.Title.Background(lipgloss.Color(issueListUnfetchedColor))
		cmd = fetchJIRAIssues(m.jiraClient, m.jql)
	case wLView:
		cmd = fetchWorkLogs(m.db)
		m.worklogList.ResetSelected()
	case syncedWLView:
		cmd = fetchSyncedWorkLogs(m.db)
		m.syncedWorklogList.ResetSelected()
	}

	return cmd
}

func (m *Model) handleRequestToGoToActiveIssue() {
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
}

func (m *Model) handleRequestToUpdateActiveWL() {
	m.activeView = editActiveWLView
	m.trackingFocussedField = entryBeginTS
	beginTSStr := m.activeIssueBeginTS.Format(timeFormat)
	m.trackingInputs[entryBeginTS].SetValue(beginTSStr)
	if m.activeIssueComment != nil {
		m.trackingInputs[entryComment].SetValue(*m.activeIssueComment)
	} else {
		m.trackingInputs[entryComment].SetValue("")
	}

	for i := range m.trackingInputs {
		m.trackingInputs[i].Blur()
	}
	m.trackingInputs[m.trackingFocussedField].Focus()
}

func (m *Model) handleRequestToCreateManualWL() {
	m.activeView = wlEntryView
	m.worklogSaveType = worklogInsert
	m.trackingFocussedField = entryBeginTS
	currentTime := time.Now()
	currentTimeStr := currentTime.Format(timeFormat)

	m.trackingInputs[entryBeginTS].SetValue(currentTimeStr)
	m.trackingInputs[entryEndTS].SetValue(currentTimeStr)

	for i := range m.trackingInputs {
		m.trackingInputs[i].Blur()
	}
	m.trackingInputs[m.trackingFocussedField].Focus()
}

func (m *Model) handleRequestToUpdateSavedWL() {
	wl, ok := m.worklogList.SelectedItem().(c.WorklogEntry)
	if !ok {
		return
	}

	m.activeView = wlEntryView
	m.worklogSaveType = worklogUpdate
	if wl.NeedsComment() {
		m.trackingFocussedField = entryComment
	} else {
		m.trackingFocussedField = entryBeginTS
	}

	beginTSStr := wl.BeginTS.Format(timeFormat)
	endTSStr := wl.EndTS.Format(timeFormat)

	m.trackingInputs[entryBeginTS].SetValue(beginTSStr)
	m.trackingInputs[entryEndTS].SetValue(endTSStr)
	var comment string
	if wl.Comment != nil {
		comment = *wl.Comment
	}
	m.trackingInputs[entryComment].SetValue(comment)

	for i := range m.trackingInputs {
		m.trackingInputs[i].Blur()
	}
	m.trackingInputs[m.trackingFocussedField].Focus()
}

func (m *Model) handleRequestToSyncTimestamps() {
	switch m.trackingFocussedField {
	case entryBeginTS:
		tsStrToSync := m.trackingInputs[entryEndTS].Value()
		_, err := time.ParseInLocation(timeFormat, tsStrToSync, time.Local)
		if err != nil {
			m.message = fmt.Sprintf("end timestamp is invalid: %s", err.Error())
			return
		}
		m.trackingInputs[entryBeginTS].SetValue(tsStrToSync)
	case entryEndTS:
		tsStrToSync := m.trackingInputs[entryBeginTS].Value()
		_, err := time.ParseInLocation(timeFormat, tsStrToSync, time.Local)
		if err != nil {
			m.message = fmt.Sprintf("begin timestamp is invalid: %s", err.Error())
			return
		}
		m.trackingInputs[entryEndTS].SetValue(tsStrToSync)
	default:
		m.message = "you need to have the cursor on either one of the two timestamps to sync them"
	}
}

func (m *Model) getCmdToDeleteWL() tea.Cmd {
	issue, ok := m.worklogList.SelectedItem().(c.WorklogEntry)
	if !ok {
		msg := "Couldn't delete worklog entry"
		m.message = msg
		m.messages = append(m.messages, msg)
		return nil
	}

	return deleteLogEntry(m.db, issue.ID)
}

func (m *Model) getCmdToQuickSwitchTracking() tea.Cmd {
	issue, ok := m.issueList.SelectedItem().(*c.Issue)
	if !ok {
		m.message = "Something went wrong"
		return nil
	}

	if issue.IssueKey == m.activeIssue {
		return nil
	}

	if !m.trackingActive {
		m.changesLocked = true
		m.activeIssueBeginTS = time.Now()
		return toggleTracking(m.db,
			issue.IssueKey,
			m.activeIssueBeginTS,
			m.activeIssueEndTS,
			"",
		)
	}

	return quickSwitchActiveIssue(m.db, issue.IssueKey, time.Now())
}

func (m *Model) getCmdToToggleTracking() tea.Cmd {
	if m.issueList.FilterState() == list.Filtering {
		return nil
	}

	if m.changesLocked {
		message := "Changes locked momentarily"
		m.message = message
		m.messages = append(m.messages, message)
		return nil
	}

	if m.lastChange == updateChange {
		return m.getCmdToStartTracking()
	}

	m.handleStoppingOfTracking()
	return nil
}

func (m *Model) getCmdToStartTracking() tea.Cmd {
	issue, ok := m.issueList.SelectedItem().(*c.Issue)
	if !ok {
		message := "Something went horribly wrong"
		m.message = message
		m.messages = append(m.messages, message)
		return nil
	}

	m.changesLocked = true
	m.activeIssueBeginTS = time.Now().Truncate(time.Second)
	return toggleTracking(m.db,
		issue.IssueKey,
		m.activeIssueBeginTS,
		m.activeIssueEndTS,
		"",
	)
}

func (m *Model) handleStoppingOfTracking() {
	currentTime := time.Now()
	beginTimeStr := m.activeIssueBeginTS.Format(timeFormat)
	currentTimeStr := currentTime.Format(timeFormat)

	m.trackingInputs[entryBeginTS].SetValue(beginTimeStr)
	m.trackingInputs[entryEndTS].SetValue(currentTimeStr)
	if m.activeIssueComment != nil {
		m.trackingInputs[entryComment].SetValue(*m.activeIssueComment)
	} else {
		m.trackingInputs[entryComment].SetValue("")
	}

	for i := range m.trackingInputs {
		m.trackingInputs[i].Blur()
	}

	m.activeView = saveActiveWLView
	m.trackingFocussedField = entryComment
	m.trackingInputs[m.trackingFocussedField].Focus()
}

func (m *Model) getCmdToSyncWLToJIRA() []tea.Cmd {
	var cmds []tea.Cmd
	toSyncNum := 0
	for i, entry := range m.worklogList.Items() {
		if wl, ok := entry.(c.WorklogEntry); ok {
			if wl.Synced {
				continue
			}

			wl.SyncInProgress = true
			m.worklogList.SetItem(i, wl)
			cmds = append(cmds, syncWorklogWithJIRA(m.jiraClient, wl, m.fallbackComment, i, m.jiraTimeDeltaMins))
			toSyncNum++
		}
	}
	if toSyncNum == 0 {
		m.message = "nothing to sync"
	}

	return cmds
}

func (m *Model) getCmdToOpenIssueInBrowser() tea.Cmd {
	selectedIssue := m.issueList.SelectedItem().FilterValue()
	return openURLInBrowser(fmt.Sprintf("%sbrowse/%s",
		m.jiraClient.BaseURL.String(),
		selectedIssue))
}

func (m *Model) handleWindowResizing(msg tea.WindowSizeMsg) {
	w, h := listStyle.GetFrameSize()
	m.terminalHeight = msg.Height
	m.issueList.SetWidth(msg.Width - w)
	m.worklogList.SetWidth(msg.Width - w)
	m.syncedWorklogList.SetWidth(msg.Width - w)
	m.issueList.SetHeight(msg.Height - h - 2)
	m.worklogList.SetHeight(msg.Height - h - 2)
	m.syncedWorklogList.SetHeight(msg.Height - h - 2)

	vw, vh := viewPortStyle.GetFrameSize()
	if !m.helpVPReady {
		m.helpVP = viewport.New(msg.Width-vw, m.terminalHeight-vh-5)
		m.helpVP.SetContent(helpText)
		m.helpVPReady = true
	} else {
		m.helpVP.Height = m.terminalHeight - vh - 5
		m.helpVP.Width = msg.Width - vw
	}
}

func (m *Model) handleIssuesFetchedFromJIRAMsg(msg issuesFetchedFromJIRA) tea.Cmd {
	if msg.err != nil {
		var remoteServerName string
		if msg.responseStatusCode >= 400 && msg.responseStatusCode < 500 {
			switch m.installationType {
			case OnPremiseInstallation:
				remoteServerName = "Your on-premise JIRA installation"
			case CloudInstallation:
				remoteServerName = "Atlassian Cloud"
			}
			m.message = fmt.Sprintf("%s returned a %d status code, check if your configuration is correct",
				remoteServerName,
				msg.responseStatusCode)
		} else {
			m.message = fmt.Sprintf("error fetching issues from JIRA: %s", msg.err.Error())
		}
		m.messages = append(m.messages, m.message)
		m.issueList.Title = "Failure"
		m.issueList.Styles.Title = m.issueList.Styles.Title.Background(lipgloss.Color(failureColor))
		return nil
	}

	issues := make([]list.Item, 0, len(msg.issues))
	for i, issue := range msg.issues {
		issue.SetDesc()
		issues = append(issues, &issue)
		m.issueMap[issue.IssueKey] = &issue
		m.issueIndexMap[issue.IssueKey] = i
	}
	m.issueList.SetItems(issues)
	m.issueList.Title = "Issues"
	m.issueList.Styles.Title = m.issueList.Styles.Title.Background(lipgloss.Color(issueListColor))
	m.issuesFetched = true

	return fetchActiveStatus(m.db, 0)
}

func (m *Model) handleManualEntryInsertedInDBMsg(msg manualWLInsertedInDB) tea.Cmd {
	if msg.err != nil {
		message := msg.err.Error()
		m.message = "Error inserting worklog: " + message
		m.messages = append(m.messages, message)
		return nil
	}

	for i := range m.trackingInputs {
		m.trackingInputs[i].SetValue("")
	}
	return fetchWorkLogs(m.db)
}

func (m *Model) handleWLUpdatedInDBMsg(msg wLUpdatedInDB) tea.Cmd {
	if msg.err != nil {
		message := msg.err.Error()
		m.message = "Error updating worklog: " + message
		m.messages = append(m.messages, message)
		return nil
	}

	m.message = "Worklog updated"
	for i := range m.trackingInputs {
		m.trackingInputs[i].SetValue("")
	}
	return fetchWorkLogs(m.db)
}

func (m *Model) handleWLEntriesFetchedFromDBMsg(msg wLEntriesFetchedFromDB) {
	if msg.err != nil {
		message := msg.err.Error()
		m.message = message
		m.messages = append(m.messages, message)
		return
	}

	items := make([]list.Item, len(msg.entries))
	var secsSpent int
	for i, e := range msg.entries {
		secsSpent += e.SecsSpent()
		e.FallbackComment = m.fallbackComment
		items[i] = list.Item(e)
	}
	m.worklogList.SetItems(items)
	m.unsyncedWLSecsSpent = secsSpent
	m.unsyncedWLCount = uint(len(msg.entries))
	if m.debug {
		m.message = "[io: log entries]"
	}
}

func (m *Model) handleSyncedWLEntriesFetchedFromDBMsg(msg syncedWLEntriesFetchedFromDB) {
	if msg.err != nil {
		message := msg.err.Error()
		m.message = "Error fetching synced worklog entries: " + message
		m.messages = append(m.messages, message)
		return
	}

	items := make([]list.Item, len(msg.entries))
	for i, e := range msg.entries {
		items[i] = list.Item(e)
	}
	m.syncedWorklogList.SetItems(items)
}

func (m *Model) handleWLSyncUpdatedInDBMsg(msg wLSyncUpdatedInDB) {
	if msg.err != nil {
		msg.entry.Error = msg.err
		m.messages = append(m.messages, msg.err.Error())
		m.worklogList.SetItem(msg.index, msg.entry)
		return
	}

	m.unsyncedWLCount--
	m.unsyncedWLSecsSpent -= msg.entry.SecsSpent()
}

func (m *Model) handleActiveWLFetchedFromDBMsg(msg activeWLFetchedFromDB) {
	if msg.err != nil {
		message := msg.err.Error()
		m.message = message
		m.messages = append(m.messages, message)
		return
	}

	m.activeIssue = msg.activeIssue
	if msg.activeIssue == "" {
		m.lastChange = updateChange
	} else {
		m.lastChange = insertChange
		activeIssue, ok := m.issueMap[m.activeIssue]
		m.activeIssueBeginTS = msg.beginTS
		m.activeIssueComment = msg.comment
		if ok {
			activeIssue.TrackingActive = true

			// go to tracked item on startup
			activeIndex, ok := m.issueIndexMap[msg.activeIssue]
			if ok {
				m.issueList.Select(activeIndex)
			}
		}
		m.trackingActive = true
	}
}

func (m *Model) handleWLDeletedFromDBMsg(msg wLDeletedFromDB) tea.Cmd {
	if msg.err != nil {
		message := "error deleting entry: " + msg.err.Error()
		m.message = message
		m.messages = append(m.messages, message)
		return nil
	}

	return fetchWorkLogs(m.db)
}

func (m *Model) handleActiveWLDeletedFromDBMsg(msg activeWLDeletedFromDB) {
	if msg.err != nil {
		m.message = fmt.Sprintf("Error deleting active log entry: %s", msg.err)
		return
	}

	activeIssue, ok := m.issueMap[m.activeIssue]
	if ok {
		activeIssue.TrackingActive = false
	}
	m.lastChange = updateChange
	m.trackingActive = false
	m.activeIssueComment = nil
	m.activeIssue = ""
}

func (m *Model) handleWLSyncedToJIRAMsg(msg wLSyncedToJIRA) tea.Cmd {
	if msg.err != nil {
		msg.entry.Error = msg.err
		m.messages = append(m.messages, msg.err.Error())
		return nil
	}

	msg.entry.Synced = true
	msg.entry.SyncInProgress = false
	if msg.fallbackCommentUsed {
		msg.entry.Comment = m.fallbackComment
	}
	m.worklogList.SetItem(msg.index, msg.entry)
	return updateSyncStatusForEntry(m.db, msg.entry, msg.index, msg.fallbackCommentUsed)
}

func (m *Model) handleActiveWLUpdatedInDBMsg(msg activeWLUpdatedInDB) {
	if msg.err != nil {
		message := msg.err.Error()
		m.message = message
		m.messages = append(m.messages, message)
		return
	}

	m.activeIssueBeginTS = msg.beginTS
	m.activeIssueComment = msg.comment
}

func (m *Model) handleTrackingToggledInDBMsg(msg trackingToggledInDB) tea.Cmd {
	if msg.err != nil {
		message := msg.err.Error()
		m.message = message
		m.messages = append(m.messages, message)
		m.trackingActive = false
		m.activeIssueComment = nil
		return nil
	}

	var activeIssue *c.Issue
	if msg.activeIssue != "" {
		activeIssue = m.issueMap[msg.activeIssue]
	} else {
		activeIssue = m.issueMap[m.activeIssue]
	}
	m.changesLocked = false
	var cmd tea.Cmd
	switch msg.finished {
	case true:
		m.lastChange = updateChange
		if activeIssue != nil {
			activeIssue.TrackingActive = false
		}
		m.trackingActive = false
		m.activeIssueComment = nil
		cmd = fetchWorkLogs(m.db)
	case false:
		m.lastChange = insertChange
		if activeIssue != nil {
			activeIssue.TrackingActive = true
		}
		m.trackingActive = true
	}

	m.activeIssue = msg.activeIssue
	return cmd
}

func (m *Model) handleActiveWLSwitchedInDBMsg(msg activeWLSwitchedInDB) {
	if msg.err != nil {
		message := msg.err.Error()
		m.message = message
		m.messages = append(m.messages, message)
		if errors.Is(msg.err, pers.ErrNoTaskIsActive) || errors.Is(msg.err, pers.ErrCouldntStartTrackingTask) {
			m.trackingActive = false
			m.activeIssueComment = nil
		}
		return
	}

	var lastActiveIssue *c.Issue
	if msg.lastActiveIssue != "" {
		lastActiveIssue = m.issueMap[msg.lastActiveIssue]
		if lastActiveIssue != nil {
			lastActiveIssue.TrackingActive = false
		}
	}

	var currentActiveIssue *c.Issue
	if msg.currentActiveIssue != "" {
		currentActiveIssue = m.issueMap[msg.currentActiveIssue]
	} else {
		currentActiveIssue = m.issueMap[m.activeIssue]
	}

	if currentActiveIssue != nil {
		currentActiveIssue.TrackingActive = true
	}
	m.activeIssue = msg.currentActiveIssue
	m.activeIssueBeginTS = msg.beginTS
	m.activeIssueComment = nil
}

func (m *Model) shiftTime(direction timeShiftDirection, duration timeShiftDuration) error {
	if m.activeView == editActiveWLView || m.activeView == saveActiveWLView || m.activeView == wlEntryView {
		if m.trackingFocussedField == entryBeginTS || m.trackingFocussedField == entryEndTS {
			ts, err := time.ParseInLocation(timeFormat, m.trackingInputs[m.trackingFocussedField].Value(), time.Local)
			if err != nil {
				return err
			}

			newTs := getShiftedTime(ts, direction, duration)

			m.trackingInputs[m.trackingFocussedField].SetValue(newTs.Format(timeFormat))
		}
	}
	return nil
}
