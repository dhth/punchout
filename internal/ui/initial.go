package ui

import (
	"database/sql"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func InitialModel(db *sql.DB, jiraClient *jira.Client, jql string, jiraTimeDeltaMins int) model {
	var stackItems []list.Item
	var worklogListItems []list.Item
	var syncedWorklogListItems []list.Item

	trackingInputs := make([]textinput.Model, 3)
	trackingInputs[entryBeginTS] = textinput.New()
	trackingInputs[entryBeginTS].Placeholder = "09:30"
	trackingInputs[entryBeginTS].Focus()
	trackingInputs[entryBeginTS].CharLimit = len(string(timeFormat))
	trackingInputs[entryBeginTS].Width = 30

	trackingInputs[entryEndTS] = textinput.New()
	trackingInputs[entryEndTS].Placeholder = "12:30pm"
	trackingInputs[entryEndTS].Focus()
	trackingInputs[entryEndTS].CharLimit = len(string(timeFormat))
	trackingInputs[entryEndTS].Width = 30

	trackingInputs[entryComment] = textinput.New()
	trackingInputs[entryComment].Placeholder = "Your comment goes here"
	trackingInputs[entryComment].Focus()
	trackingInputs[entryComment].CharLimit = 255
	trackingInputs[entryComment].Width = 60

	m := model{
		db:                db,
		jiraClient:        jiraClient,
		jql:               jql,
		issueList:         list.New(stackItems, newItemDelegate(lipgloss.Color(issueListColor)), listWidth, 0),
		issueMap:          make(map[string]*Issue),
		issueIndexMap:     make(map[string]int),
		worklogList:       list.New(worklogListItems, newItemDelegate(lipgloss.Color(worklogListColor)), listWidth, 0),
		syncedWorklogList: list.New(syncedWorklogListItems, newItemDelegate(syncedWorklogListColor), listWidth, 0),
		jiraTimeDeltaMins: jiraTimeDeltaMins,
		showHelpIndicator: true,
		trackingInputs:    trackingInputs,
	}
	m.issueList.Title = "fetching..."
	m.issueList.SetStatusBarItemName("issue", "issues")
	m.issueList.DisableQuitKeybindings()
	m.issueList.SetShowHelp(false)
	m.issueList.Styles.Title = m.issueList.Styles.Title.Foreground(lipgloss.Color(defaultBackgroundColor)).
		Background(lipgloss.Color(issueListUnfetchedColor)).
		Bold(true)

	m.worklogList.Title = "Worklog Entries"
	m.worklogList.SetStatusBarItemName("entry", "entries")
	m.worklogList.SetFilteringEnabled(false)
	m.worklogList.DisableQuitKeybindings()
	m.worklogList.SetShowHelp(false)
	m.worklogList.Styles.Title = m.worklogList.Styles.Title.Foreground(lipgloss.Color(defaultBackgroundColor)).
		Background(lipgloss.Color(worklogListColor)).
		Bold(true)

	m.syncedWorklogList.Title = "Synced Worklog Entries (from local db)"
	m.syncedWorklogList.SetStatusBarItemName("entry", "entries")
	m.syncedWorklogList.SetFilteringEnabled(false)
	m.syncedWorklogList.DisableQuitKeybindings()
	m.syncedWorklogList.SetShowHelp(false)
	m.syncedWorklogList.Styles.Title = m.syncedWorklogList.Styles.Title.Foreground(lipgloss.Color(defaultBackgroundColor)).
		Background(lipgloss.Color(syncedWorklogListColor)).
		Bold(true)

	return m
}
