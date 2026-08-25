package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	d "github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/issuecache"
	pers "github.com/dhth/punchout/internal/persistence"
	"github.com/dhth/punchout/internal/utils"
)

func (m *Model) getCmdToUpdateActiveWL() tea.Cmd {
	beginTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryBeginTS].Value(), time.Local)
	if err != nil {
		m.setErrorMsg(err.Error())
		return nil
	}
	commentValue := m.trackingInputs[entryComment].Value()

	var comment *string
	if strings.TrimSpace(commentValue) != "" {
		comment = &commentValue
	}
	m.trackingInputs[entryBeginTS].SetValue("")
	m.activeView = issueListView
	return updateActiveWL(m.ctx, m.worklogStore, beginTS, comment)
}

func (m *Model) getCmdToSaveActiveWL() tea.Cmd {
	if m.activeWorklog == nil {
		return nil
	}

	beginTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryBeginTS].Value(), time.Local)
	if err != nil {
		m.setErrorMsg(err.Error())
		return nil
	}
	m.activeWorklog.BeginTS = beginTS.Local()

	endTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryEndTS].Value(), time.Local)
	if err != nil {
		m.setErrorMsg(err.Error())
		return nil
	}
	m.activeIssueEndTS = endTS.Local()

	if !m.isDurationValid(m.activeWorklog.BeginTS, m.activeIssueEndTS) {
		return nil
	}

	comment := m.trackingInputs[entryComment].Value()

	m.activeView = issueListView
	for i := range m.trackingInputs {
		m.trackingInputs[i].SetValue("")
	}

	return toggleTracking(
		m.db,
		trackingToggleFinish,
		m.activeIssue,
		m.activeWorklog.BeginTS,
		m.activeIssueEndTS,
		comment,
	)
}

func (m *Model) getCmdToSaveOrUpdateWL() tea.Cmd {
	beginTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryBeginTS].Value(), time.Local)
	if err != nil {
		m.setErrorMsg(err.Error())
		return nil
	}

	endTS, err := time.ParseInLocation(timeFormat, m.trackingInputs[entryEndTS].Value(), time.Local)
	if err != nil {
		m.setErrorMsg(err.Error())
		return nil
	}

	if !m.isDurationValid(beginTS, endTS) {
		return nil
	}

	issue, ok := m.issueList.SelectedItem().(*d.Issue)

	var cmd tea.Cmd
	if ok {
		switch m.worklogSaveType {
		case worklogInsert:
			worklog := d.Worklog{
				IssueKey: issue.IssueKey,
				BeginTS:  beginTS,
				EndTS:    endTS,
				Comment:  m.trackingInputs[entryComment].Value(),
			}
			cmd = insertManualEntry(m.ctx, m.worklogStore, worklog)
			m.activeView = issueListView
		case worklogUpdate:
			wl, ok := m.worklogList.SelectedItem().(worklogListItem)
			if ok {
				cmd = updateManualEntry(
					m.ctx,
					m.worklogStore,
					wl.ID,
					d.Worklog{
						IssueKey: wl.IssueKey,
						BeginTS:  beginTS,
						EndTS:    endTS,
						Comment:  m.trackingInputs[entryComment].Value(),
					},
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
		cmd = fetchUnsyncedWorkLogs(m.ctx, m.worklogStore)
	case wLView:
		m.activeView = syncedWLView
		cmd = fetchSyncedWorkLogs(m.ctx, m.worklogStore)
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
		cmd = fetchUnsyncedWorkLogs(m.ctx, m.worklogStore)
	case issueListView:
		m.activeView = syncedWLView
		cmd = fetchSyncedWorkLogs(m.ctx, m.worklogStore)
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
		m.issueList.Title = issueListFetchingTitle
		m.issueList.Styles.Title = m.styles.issueListUnfetchedTitle
		cmd = m.fetchIssuesFromJIRA(false)
	case wLView:
		cmd = fetchUnsyncedWorkLogs(m.ctx, m.worklogStore)
		m.worklogList.ResetSelected()
	case syncedWLView:
		cmd = fetchSyncedWorkLogs(m.ctx, m.worklogStore)
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
			m.setInfoMsg("nothing is being tracked right now")
		}
	}
}

func (m *Model) handleRequestToUpdateActiveWL() {
	if m.activeWorklog == nil {
		return
	}

	m.activeView = editActiveWLView
	m.trackingFocussedField = entryBeginTS
	beginTSStr := m.activeWorklog.BeginTS.Format(timeFormat)
	m.trackingInputs[entryBeginTS].SetValue(beginTSStr)
	m.trackingInputs[entryComment].SetValue(m.activeWorklog.Comment)

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
	wl, ok := m.worklogList.SelectedItem().(worklogListItem)
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
	m.trackingInputs[entryComment].SetValue(wl.Comment)

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
			m.setErrorMsg(fmt.Sprintf("end timestamp is invalid: %s", err.Error()))
			return
		}
		m.trackingInputs[entryBeginTS].SetValue(tsStrToSync)
	case entryEndTS:
		tsStrToSync := m.trackingInputs[entryBeginTS].Value()
		_, err := time.ParseInLocation(timeFormat, tsStrToSync, time.Local)
		if err != nil {
			m.setErrorMsg(fmt.Sprintf("begin timestamp is invalid: %s", err.Error()))
			return
		}
		m.trackingInputs[entryEndTS].SetValue(tsStrToSync)
	default:
		m.setErrorMsg("you need to have the cursor on either one of the two timestamps to sync them")
	}
}

func (m *Model) getCmdToDeleteWL() tea.Cmd {
	issue, ok := m.worklogList.SelectedItem().(worklogListItem)
	if !ok {
		m.setErrorMsg("couldn't delete worklog entry")
		return nil
	}

	return deleteLogEntry(m.ctx, m.worklogStore, issue.ID)
}

func (m *Model) getCmdToQuickSwitchTracking() tea.Cmd {
	issue, ok := m.issueList.SelectedItem().(*d.Issue)
	if !ok {
		m.setErrorMsg("something went wrong")
		return nil
	}

	if issue.IssueKey == m.activeIssue {
		return nil
	}

	if !m.trackingActive {
		m.changesLocked = true
		m.activeWorklog = &d.InProgressWorklog{IssueKey: issue.IssueKey, BeginTS: time.Now()}
		return toggleTracking(
			m.db,
			trackingToggleStart,
			issue.IssueKey,
			m.activeWorklog.BeginTS,
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
		m.setInfoMsg("changes locked momentarily")
		return nil
	}

	if m.lastChange == updateChange {
		return m.getCmdToStartTracking()
	}

	m.handleStoppingOfTracking()
	return nil
}

func (m *Model) getCmdToStartTracking() tea.Cmd {
	issue, ok := m.issueList.SelectedItem().(*d.Issue)
	if !ok {
		m.setErrorMsg("something went horribly wrong")
		return nil
	}

	m.changesLocked = true
	m.activeWorklog = &d.InProgressWorklog{IssueKey: issue.IssueKey, BeginTS: time.Now().Truncate(time.Second)}
	return toggleTracking(
		m.db,
		trackingToggleStart,
		issue.IssueKey,
		m.activeWorklog.BeginTS,
		m.activeIssueEndTS,
		"",
	)
}

func (m *Model) handleStoppingOfTracking() {
	if m.activeWorklog == nil {
		return
	}

	currentTime := time.Now()
	beginTimeStr := m.activeWorklog.BeginTS.Format(timeFormat)
	currentTimeStr := currentTime.Format(timeFormat)

	m.trackingInputs[entryBeginTS].SetValue(beginTimeStr)
	m.trackingInputs[entryEndTS].SetValue(currentTimeStr)
	m.trackingInputs[entryComment].SetValue(m.activeWorklog.Comment)

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
		if wl, ok := entry.(worklogListItem); ok {
			if wl.Synced {
				continue
			}

			wl.syncInProgress = true
			wl.err = nil
			m.worklogList.SetItem(i, wl)
			cmds = append(cmds, m.syncWorklogWithJIRA(wl, i))
			toSyncNum++
		}
	}
	if toSyncNum == 0 {
		m.setInfoMsg("nothing to sync")
	}

	return cmds
}

func (m *Model) getCmdToOpenIssueInBrowser() tea.Cmd {
	selectedIssue := m.issueList.SelectedItem()
	if selectedIssue == nil {
		return nil
	}

	return openURLInBrowser(fmt.Sprintf("%sbrowse/%s",
		m.jiraSvc.URL(),
		selectedIssue.FilterValue()))
}

func (m *Model) handleWindowResizing(msg tea.WindowSizeMsg) {
	w, h := m.styles.list.GetFrameSize()
	m.terminalHeight = msg.Height
	m.issueList.SetWidth(msg.Width - w)
	m.worklogList.SetWidth(msg.Width - w)
	m.syncedWorklogList.SetWidth(msg.Width - w)
	m.issueList.SetHeight(msg.Height - h - 2)
	m.worklogList.SetHeight(msg.Height - h - 2)
	m.syncedWorklogList.SetHeight(msg.Height - h - 2)

	vw, vh := m.styles.viewPort.GetFrameSize()
	if !m.helpVPReady {
		m.helpVP = viewport.New(viewport.WithWidth(msg.Width-vw), viewport.WithHeight(m.terminalHeight-vh-5))
		m.helpVP.SetContent(renderHelp(m.styles))
		m.helpVPReady = true
	} else {
		m.helpVP.SetHeight(m.terminalHeight - vh - 5)
		m.helpVP.SetWidth(msg.Width - vw)
	}
}

func (m *Model) handleIssuesLoadedMsg(msg issuesLoaded) []tea.Cmd {
	if msg.err != nil {
		if msg.source == issueSourceCache {
			return []tea.Cmd{m.fetchIssuesFromJIRA(true)}
		}

		var errorMsg string
		if msg.afterCacheLoadFailure {
			errorMsg = fmt.Sprintf("cache unavailable and couldn't fetch issues from JIRA: %s", msg.err.Error())
		} else {
			errorMsg = fmt.Sprintf("couldn't fetch issues from JIRA: %s", msg.err.Error())
		}

		m.setErrorMsg(errorMsg)
		m.issueList.Title = failureTitle
		m.issueList.Styles.Title = m.styles.issueListFailureTitle
		return nil
	}

	issues := make([]list.Item, 0, len(msg.issues))
	for i, issue := range msg.issues {
		issues = append(issues, &issue)
		m.issueMap[issue.IssueKey] = &issue
		m.issueIndexMap[issue.IssueKey] = i
	}
	m.issueList.SetItems(issues)
	m.issueList.Title = "▪▫▫ Issues"
	m.issueList.Styles.Title = m.styles.issueListTitle
	m.issuesFetched = true

	cmds := []tea.Cmd{fetchActiveStatus(m.db, 0)}
	switch msg.source {
	case issueSourceCache:
		if len(msg.issues) == 0 {
			m.setInfoMsg("no issues found in cache")
			break
		}

		secondsSinceFetch := int(time.Since(msg.fetchedAt).Seconds())
		if secondsSinceFetch <= 0 {
			m.setInfoMsg("issues loaded from cache")
		} else {
			fetchedAgo := utils.HumanizeDuration(secondsSinceFetch)
			m.setInfoMsg(fmt.Sprintf("issues loaded from cache • fetched %s ago", fetchedAgo))
		}
	case issueSourceJIRA:
		if msg.afterCacheLoadFailure {
			m.setInfoMsg("cache unavailable • issues fetched from JIRA")
		}

		cmds = append(cmds, m.saveIssuesToCache(issuecache.Snapshot{
			Issues:    msg.issues,
			FetchedAt: msg.fetchedAt,
		}))
	}

	return cmds
}

func (m *Model) handleIssuesSavedToCacheMsg(msg issuesSavedToCache) {
	if msg.err == nil {
		return
	}

	m.setErrorMsg(fmt.Sprintf("error saving issues to cache: %s", msg.err.Error()))
}

func (m *Model) handleManualEntryInsertedInDBMsg(msg manualWLInsertedInDB) tea.Cmd {
	if msg.err != nil {
		m.setErrorMsg("error inserting worklog: " + msg.err.Error())
		return nil
	}

	for i := range m.trackingInputs {
		m.trackingInputs[i].SetValue("")
	}
	return fetchUnsyncedWorkLogs(m.ctx, m.worklogStore)
}

func (m *Model) handleWLUpdatedInDBMsg(msg wLUpdatedInDB) tea.Cmd {
	if msg.err != nil {
		m.setErrorMsg("error updating worklog: " + msg.err.Error())
		return nil
	}

	m.setInfoMsg("worklog updated")
	for i := range m.trackingInputs {
		m.trackingInputs[i].SetValue("")
	}
	return fetchUnsyncedWorkLogs(m.ctx, m.worklogStore)
}

func (m *Model) handleWLEntriesFetchedFromDBMsg(msg wLEntriesFetchedFromDB) {
	if msg.err != nil {
		m.setErrorMsg(msg.err.Error())
		return
	}

	items := make([]list.Item, len(msg.entries))
	var secsSpent int
	for i, e := range msg.entries {
		secsSpent += e.SecsSpent()
		items[i] = worklogListItem{StoredWorklog: e, fallbackComment: m.opts.Jira.FallbackComment}
	}
	m.worklogList.SetItems(items)
	m.unsyncedWLSecsSpent = secsSpent
	m.unsyncedWLCount = uint(len(msg.entries))
	if m.debug {
		m.setInfoMsg("[io: log entries]")
	}
}

func (m *Model) handleSyncedWLEntriesFetchedFromDBMsg(msg syncedWLEntriesFetchedFromDB) {
	if msg.err != nil {
		m.setErrorMsg("error fetching synced worklog entries: " + msg.err.Error())
		return
	}

	items := make([]list.Item, len(msg.entries))
	for i, e := range msg.entries {
		items[i] = syncedWorklogListItem{StoredWorklog: e}
	}
	m.syncedWorklogList.SetItems(items)
}

func (m *Model) handleWLSyncUpdatedInDBMsg(msg wLSyncUpdatedInDB) {
	if msg.err != nil {
		msg.entry.err = msg.err
		m.worklogList.SetItem(msg.index, msg.entry)
		return
	}

	m.unsyncedWLCount--
	m.unsyncedWLSecsSpent -= msg.entry.SecsSpent()
}

func (m *Model) handleActiveWLFetchedFromDBMsg(msg activeWLFetchedFromDB) {
	if msg.err != nil {
		m.setErrorMsg(msg.err.Error())
		return
	}

	if previouslyActiveIssue := m.issueMap[m.activeIssue]; previouslyActiveIssue != nil {
		previouslyActiveIssue.TrackingActive = false
	}

	if msg.worklog == nil {
		m.activeIssue = ""
		m.activeWorklog = nil
		m.lastChange = updateChange
		m.trackingActive = false
	} else {
		m.activeIssue = msg.worklog.IssueKey
		m.lastChange = insertChange
		activeIssue, ok := m.issueMap[m.activeIssue]
		m.activeWorklog = msg.worklog
		if ok {
			activeIssue.TrackingActive = true

			// go to tracked item on startup
			activeIndex, ok := m.issueIndexMap[msg.worklog.IssueKey]
			if ok {
				m.issueList.Select(activeIndex)
			}
		}
		m.trackingActive = true
	}
}

func (m *Model) handleWLDeletedFromDBMsg(msg wLDeletedFromDB) tea.Cmd {
	if msg.err != nil {
		m.setErrorMsg("error deleting entry: " + msg.err.Error())
		return nil
	}

	return fetchUnsyncedWorkLogs(m.ctx, m.worklogStore)
}

func (m *Model) handleActiveWLDeletedFromDBMsg(msg activeWLDeletedFromDB) {
	if msg.err != nil {
		m.setErrorMsg(fmt.Sprintf("error deleting active log entry: %s", msg.err))
		return
	}

	activeIssue, ok := m.issueMap[m.activeIssue]
	if ok {
		activeIssue.TrackingActive = false
	}
	m.lastChange = updateChange
	m.trackingActive = false
	m.activeWorklog = nil
	m.activeIssue = ""
}

func (m *Model) handleWLSyncedToJIRAMsg(msg wLSyncedToJIRA) tea.Cmd {
	if msg.err != nil {
		msg.entry.err = msg.err
		msg.entry.syncInProgress = false
		m.worklogList.SetItem(msg.index, msg.entry)
		return nil
	}

	msg.entry.Synced = true
	msg.entry.syncInProgress = false
	if msg.fallbackCommentUsed {
		msg.entry.Comment = *m.opts.Jira.FallbackComment
	}
	m.worklogList.SetItem(msg.index, msg.entry)
	return updateSyncStatusForEntry(m.ctx, m.worklogStore, msg.entry, msg.index, msg.fallbackCommentUsed)
}

func (m *Model) handleActiveWLUpdatedInDBMsg(msg activeWLUpdatedInDB) {
	if msg.err != nil {
		m.setErrorMsg(msg.err.Error())
		return
	}

	if m.activeWorklog == nil {
		return
	}

	m.activeWorklog.BeginTS = msg.beginTS
	if msg.comment != nil {
		m.activeWorklog.Comment = *msg.comment
	}
}

func (m *Model) handleTrackingToggledInDBMsg(msg trackingToggledInDB) tea.Cmd {
	if msg.err != nil {
		m.setErrorMsg(msg.err.Error())
		m.changesLocked = false
		if msg.reconcileActiveStatus {
			return fetchActiveStatus(m.db, 0)
		}
		switch msg.operation {
		case trackingToggleStart:
			m.trackingActive = false
			m.activeIssue = ""
			m.activeWorklog = nil
		case trackingToggleFinish:
			m.trackingActive = true
			m.activeIssue = msg.activeIssue
			if m.activeWorklog != nil {
				m.activeWorklog.IssueKey = msg.activeIssue
			}
			if activeIssue := m.issueMap[msg.activeIssue]; activeIssue != nil {
				activeIssue.TrackingActive = true
			}
		}
		return nil
	}

	var activeIssue *d.Issue
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
		m.activeWorklog = nil
		cmd = fetchUnsyncedWorkLogs(m.ctx, m.worklogStore)
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
		m.setErrorMsg(msg.err.Error())
		switch {
		case errors.Is(msg.err, pers.ErrNoTaskIsActive):
			if activeIssue := m.issueMap[m.activeIssue]; activeIssue != nil {
				activeIssue.TrackingActive = false
			}
			m.activeIssue = ""
			m.trackingActive = false
			m.activeWorklog = nil
		case errors.Is(msg.err, pers.ErrCouldntStartTrackingTask):
			if lastActiveIssue := m.issueMap[msg.lastActiveIssue]; lastActiveIssue != nil {
				lastActiveIssue.TrackingActive = false
			}
			m.activeIssue = ""
			m.trackingActive = false
			m.activeWorklog = nil
		case errors.Is(msg.err, pers.ErrCouldntStopActiveTask):
			m.activeIssue = msg.lastActiveIssue
			m.trackingActive = true
			if m.activeWorklog != nil {
				m.activeWorklog.IssueKey = msg.lastActiveIssue
			}
			if activeIssue := m.issueMap[msg.lastActiveIssue]; activeIssue != nil {
				activeIssue.TrackingActive = true
			}
		}
		return
	}

	var lastActiveIssue *d.Issue
	if msg.lastActiveIssue != "" {
		lastActiveIssue = m.issueMap[msg.lastActiveIssue]
		if lastActiveIssue != nil {
			lastActiveIssue.TrackingActive = false
		}
	}

	var currentActiveIssue *d.Issue
	if msg.currentActiveIssue != "" {
		currentActiveIssue = m.issueMap[msg.currentActiveIssue]
	} else {
		currentActiveIssue = m.issueMap[m.activeIssue]
	}

	if currentActiveIssue != nil {
		currentActiveIssue.TrackingActive = true
	}
	m.activeIssue = msg.currentActiveIssue
	m.activeWorklog = &d.InProgressWorklog{IssueKey: msg.currentActiveIssue, BeginTS: msg.beginTS}
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

func (m *Model) getCmdToQuickFinishActiveWL() tea.Cmd {
	if m.activeWorklog == nil {
		return nil
	}

	now := time.Now().Truncate(time.Second)
	if !m.isDurationValid(m.activeWorklog.BeginTS, now) {
		return nil
	}

	m.activeIssueEndTS = now

	return toggleTracking(m.db, trackingToggleFinish, m.activeIssue, m.activeWorklog.BeginTS, m.activeIssueEndTS, "")
}

func (m *Model) isDurationValid(start, end time.Time) bool {
	if end.Sub(start).Seconds() < 60 {
		m.setErrorMsg("time spent needs to be at least a minute")
		return false
	}
	return true
}
