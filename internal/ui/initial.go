package ui

import (
	"database/sql"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	d "github.com/dhth/punchout/internal/domain"
)

func InitialModel(db *sql.DB, jiraClient *jira.Client, installationType JiraInstallationType, jql string, jiraTimeDeltaMins int, fallbackComment *string, debug bool) Model {
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

	m := Model{
		db:                db,
		jiraClient:        jiraClient,
		installationType:  installationType,
		jql:               jql,
		fallbackComment:   fallbackComment,
		issueList:         list.New(stackItems, newItemDelegate(lipgloss.Color(issueListColor)), listWidth, 0),
		issueMap:          make(map[string]*d.Issue),
		issueIndexMap:     make(map[string]int),
		worklogList:       list.New(worklogListItems, newItemDelegate(lipgloss.Color(worklogListColor)), listWidth, 0),
		syncedWorklogList: list.New(syncedWorklogListItems, newItemDelegate(syncedWorklogListColor), listWidth, 0),
		jiraTimeDeltaMins: jiraTimeDeltaMins,
		showHelpIndicator: true,
		trackingInputs:    trackingInputs,
		debug:             debug,
	}
	m.issueList.Title = "fetching..."
	m.issueList.SetStatusBarItemName("issue", "issues")
	m.issueList.DisableQuitKeybindings()
	m.issueList.SetShowHelp(false)
	m.issueList.Styles.Title = m.issueList.Styles.Title.Foreground(lipgloss.Color(d.DefaultBackgroundColor)).
		Background(lipgloss.Color(issueListUnfetchedColor)).
		Bold(true)
	m.issueList.KeyMap.PrevPage.SetKeys("left", "h", "pgup")
	m.issueList.KeyMap.NextPage.SetKeys("right", "l", "pgdown")

	m.worklogList.Title = "▫▪▫ Worklog Entries"
	m.worklogList.SetStatusBarItemName("entry", "entries")
	m.worklogList.SetFilteringEnabled(false)
	m.worklogList.DisableQuitKeybindings()
	m.worklogList.SetShowHelp(false)
	m.worklogList.Styles.Title = m.worklogList.Styles.Title.Foreground(lipgloss.Color(d.DefaultBackgroundColor)).
		Background(lipgloss.Color(worklogListColor)).
		Bold(true)
	m.worklogList.KeyMap.PrevPage.SetKeys("left", "h", "pgup")
	m.worklogList.KeyMap.NextPage.SetKeys("right", "l", "pgdown")

	m.syncedWorklogList.Title = "▫▫▪ Synced Entries"
	m.syncedWorklogList.SetStatusBarItemName("entry", "entries")
	m.syncedWorklogList.SetFilteringEnabled(false)
	m.syncedWorklogList.DisableQuitKeybindings()
	m.syncedWorklogList.SetShowHelp(false)
	m.syncedWorklogList.Styles.Title = m.syncedWorklogList.Styles.Title.Foreground(lipgloss.Color(d.DefaultBackgroundColor)).
		Background(lipgloss.Color(syncedWorklogListColor)).
		Bold(true)
	m.syncedWorklogList.KeyMap.PrevPage.SetKeys("left", "h", "pgup")
	m.syncedWorklogList.KeyMap.NextPage.SetKeys("right", "l", "pgdown")

	return m
}
