package ui

import (
	"database/sql"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const activeStatusTickInterval = time.Second * 3

type TrackingStatus uint

const (
	TrackingInactive TrackingStatus = iota
	TrackingActive
)

type DBChange uint

const (
	InsertChange DBChange = iota
	UpdateChange
)

type StateView uint

const (
	IssueListView StateView = iota
	WorklogView
	AskForCommentView
	HelpView
)

type model struct {
	activeView        StateView
	lastView          StateView
	db                *sql.DB
	jiraClient        *jira.Client
	jql               string
	issueList         list.Model
	worklogList       list.Model
	helpVP            viewport.Model
	helpVPReady       bool
	commentInput      textarea.Model
	lastChange        DBChange
	changesLocked     bool
	activeIssue       string
	activeIssueIndex  int
	issueDetails      map[string]string
	message           string
	errorMessage      string
	messages          []string
	jiraTimeDeltaMins int
	showHelpIndicator bool
	terminalHeight    int
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		hideHelp(time.Second*15),
		fetchJIRAIssues(m.jiraClient, m.jql),
		fetchActiveStatus(m.db, 0),
	)
}
