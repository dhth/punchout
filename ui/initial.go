package ui

import (
	"database/sql"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
)

func InitialModel(db *sql.DB, jiraClient *jira.Client, jql string, jiraTimeDeltaMins int) model {
	var stackItems []list.Item
	var worklogListItems []list.Item

	var appDelegateKeys = newDelegateKeyMap()
	itemDel := newItemDelegate(appDelegateKeys)

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
		issueList:         list.New(stackItems, itemDel, listWidth, 0),
		worklogList:       list.New(worklogListItems, itemDel, listWidth, 0),
		jiraTimeDeltaMins: jiraTimeDeltaMins,
		showHelpIndicator: true,
		trackingInputs:    trackingInputs,
	}
	m.issueList.Title = "Issues (fetching ...)"
	m.issueList.SetStatusBarItemName("issue", "issues")
	m.issueList.DisableQuitKeybindings()
	m.issueList.SetShowHelp(false)

	m.worklogList.Title = "Worklog Entries"
	m.worklogList.SetStatusBarItemName("entry", "entries")
	m.worklogList.SetFilteringEnabled(false)
	m.worklogList.DisableQuitKeybindings()
	m.worklogList.SetShowHelp(false)

	return m
}
