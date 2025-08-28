package ui

import (
	"database/sql"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	d "github.com/dhth/punchout/internal/domain"
	svc "github.com/dhth/punchout/internal/service"
)

type trackingStatus uint

const (
	trackingInactive trackingStatus = iota
	trackingActive
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

const (
	timeFormat     = "2006/01/02 15:04"
	timeOnlyFormat = "15:04"
)

type Model struct {
	activeView            stateView
	lastView              stateView
	db                    *sql.DB
	jiraSvc               svc.Jira
	jiraCfg               d.JiraConfig
	issueList             list.Model
	issueMap              map[string]*d.Issue
	issueIndexMap         map[string]int
	issuesFetched         bool
	worklogList           list.Model
	unsyncedWLCount       uint
	unsyncedWLSecsSpent   int
	syncedWorklogList     list.Model
	activeIssueBeginTS    time.Time
	activeIssueEndTS      time.Time
	activeIssueComment    *string
	trackingInputs        []textinput.Model
	trackingFocussedField trackingFocussedField
	helpVP                viewport.Model
	helpVPReady           bool
	lastChange            dBChange
	changesLocked         bool
	activeIssue           string
	worklogSaveType       worklogSaveType
	message               string
	messages              []string
	showHelpIndicator     bool
	terminalHeight        int
	trackingActive        bool
	debug                 bool
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		hideHelp(time.Minute*1),
		m.fetchJIRAIssues(),
		fetchUnsyncedWorkLogs(m.db),
		fetchSyncedWorkLogs(m.db),
	)
}
