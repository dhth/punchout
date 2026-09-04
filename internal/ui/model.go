package ui

import (
	"context"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/dhth/punchout/internal/config"
	d "github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/issuecache"
	svc "github.com/dhth/punchout/internal/service"
	"github.com/dhth/punchout/internal/ui/theme"
)

type dBChange uint

const (
	insertChange dBChange = iota
	updateChange
)

type stateView uint

const (
	issueListView    stateView = iota // shows issues
	wLView                            // shows worklogs that aren't yet synced
	syncedWLView                      // shows worklogs that are synced
	editActiveWLView                  // edit the active worklog
	saveActiveWLView                  // finish the active worklog
	wlEntryView                       // for saving manual worklog, or for updating a saved worklog
	helpView
)

type trackingFocussedField uint

const (
	entryBeginTS trackingFocussedField = iota
	entryEndTS
	entryComment
)

type worklogSaveType uint

const (
	worklogInsert worklogSaveType = iota
	worklogUpdate
)

type userMsgKind uint

const (
	userMsgInfo userMsgKind = iota
	userMsgError
)

const (
	userMsgInfoDuration  = 3 * time.Second
	userMsgErrorDuration = 5 * time.Second
)

type userMsg struct {
	id    uint64
	value string
	kind  userMsgKind
}

func (m userMsg) isActive() bool {
	return m.id != 0
}

func (m userMsg) duration() time.Duration {
	if m.kind == userMsgError {
		return userMsgErrorDuration
	}

	return userMsgInfoDuration
}

const (
	timeFormat             = "2006/01/02 15:04"
	timeOnlyFormat         = "15:04"
	issueListFetchingTitle = "fetching..."
	failureTitle           = "Failure"
)

type Options struct {
	Jira              config.JiraOptions
	UseCacheOnStartup bool
}

type Model struct {
	ctx                   context.Context
	theme                 theme.Theme
	styles                styles
	activeView            stateView
	lastView              stateView
	worklogStore          WorklogStore
	jiraSvc               svc.Jira
	issueStore            issuecache.Store
	opts                  Options
	issueList             list.Model
	issueMap              map[string]*d.Issue
	issueIndexMap         map[string]int
	issuesFetched         bool
	worklogList           list.Model
	worklogListGen        uint64
	worklogSyncsRemaining int
	unsyncedWLCount       uint
	unsyncedWLSecsSpent   int
	syncedWorklogList     list.Model
	activeIssueEndTS      time.Time
	activeWorklog         *d.InProgressWorklog
	trackingInputs        []textinput.Model
	trackingFocussedField trackingFocussedField
	helpVP                viewport.Model
	helpVPReady           bool
	lastChange            dBChange
	changesLocked         bool
	activeIssue           string
	worklogSaveType       worklogSaveType
	message               userMsg
	nextMessageID         uint64
	showHelpIndicator     bool
	terminalHeight        int
	trackingActive        bool
	debug                 bool
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{hideHelp(time.Minute * 1)}
	if m.opts.UseCacheOnStartup {
		cmds = append(cmds, m.loadIssuesFromCache())
	} else {
		cmds = append(cmds, m.fetchIssuesFromJIRA(false))
	}
	cmds = append(cmds, fetchUnsyncedWorkLogs(m.ctx, m.worklogStore, m.worklogListGen), fetchSyncedWorkLogs(m.ctx, m.worklogStore))

	return tea.Batch(cmds...)
}

func (m *Model) setInfoMsg(value string) {
	m.setUserMsg(value, userMsgInfo)
}

func (m *Model) setErrorMsg(value string) {
	m.setUserMsg(value, userMsgError)
}

func (m *Model) setUserMsg(value string, kind userMsgKind) {
	m.nextMessageID++
	m.message = userMsg{
		id:    m.nextMessageID,
		value: value,
		kind:  kind,
	}
}

func (m *Model) applyTheme(thm theme.Theme) {
	m.theme = thm
	m.styles = newStyles(thm)

	fallbackCommentConfigured := m.opts.Jira.FallbackComment != nil
	m.issueList.SetDelegate(newItemDelegate(thm, m.styles, thm.Accent1, m.issueMap, fallbackCommentConfigured))
	m.worklogList.SetDelegate(newItemDelegate(thm, m.styles, thm.Accent2, m.issueMap, fallbackCommentConfigured))
	m.syncedWorklogList.SetDelegate(newItemDelegate(thm, m.styles, thm.Accent4, m.issueMap, fallbackCommentConfigured))

	switch m.issueList.Title {
	case issueListFetchingTitle:
		m.issueList.Styles.Title = m.styles.issueListUnfetchedTitle
	case failureTitle:
		m.issueList.Styles.Title = m.styles.issueListFailureTitle
	default:
		m.issueList.Styles.Title = m.styles.issueListTitle
	}
	m.worklogList.Styles.Title = m.styles.worklogListTitle
	m.syncedWorklogList.Styles.Title = m.styles.syncedWorklogListTitle

	if m.helpVPReady {
		m.helpVP.SetContent(renderHelp(m.styles))
	}
}
