package ui

import (
	"database/sql"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	c "github.com/dhth/punchout/internal/common"
)

type JiraInstallationType uint

const (
	OnPremiseInstallation JiraInstallationType = iota
	CloudInstallation
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
	issueListView stateView = iota
	worklogView
	syncedWorklogView
	editActiveWLView
	askForCommentView
	manualWorklogEntryView
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
	timeFormat       = "2006/01/02 15:04"
	dayAndTimeFormat = "Mon, 15:04"
	dateFormat       = "2006/01/02"
	timeOnlyFormat   = "15:04"
)

type Model struct {
	activeView            stateView
	lastView              stateView
	db                    *sql.DB
	jiraClient            *jira.Client
	installationType      JiraInstallationType
	jql                   string
	fallbackComment       *string
	issueList             list.Model
	issueMap              map[string]*c.Issue
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
	jiraTimeDeltaMins     int
	showHelpIndicator     bool
	terminalHeight        int
	trackingActive        bool
	debug                 bool
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		hideHelp(time.Minute*1),
		fetchJIRAIssues(m.jiraClient, m.jql),
		fetchLogEntries(m.db),
		fetchSyncedLogEntries(m.db),
	)
}
