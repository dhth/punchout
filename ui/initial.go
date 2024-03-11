package ui

import (
	"database/sql"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
)

func InitialModel(db *sql.DB, jiraClient *jira.Client, jql string, jiraTimeDeltaMins int) model {
	var stackItems []list.Item
	var worklogListItems []list.Item

	var appDelegateKeys = newDelegateKeyMap()
	itemDel := newItemDelegate(appDelegateKeys)

	ta := textarea.New()
	ta.MaxHeight = 10
	ta.Focus()

	m := model{
		db:                db,
		jiraClient:        jiraClient,
		jql:               jql,
		issueList:         list.New(stackItems, itemDel, listWidth, 0),
		worklogList:       list.New(worklogListItems, itemDel, listWidth, 0),
		commentInput:      ta,
		jiraTimeDeltaMins: jiraTimeDeltaMins,
		showHelpIndicator: true,
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
