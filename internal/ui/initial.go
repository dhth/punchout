package ui

import (
	"database/sql"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	d "github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/issuecache"
	svc "github.com/dhth/punchout/internal/service"
	"github.com/dhth/punchout/internal/ui/theme"
)

func InitialModel(
	db *sql.DB,
	worklogStore WorklogStore,
	jiraSvc svc.Jira,
	issueStore issuecache.Store,
	opts Options,
	thm theme.Theme,
	debug bool,
) Model {
	styles := newStyles(thm)
	var stackItems []list.Item
	var worklogListItems []list.Item
	var syncedWorklogListItems []list.Item

	trackingInputs := make([]textinput.Model, 3)
	trackingInputs[entryBeginTS] = textinput.New()
	trackingInputs[entryBeginTS].Placeholder = "09:30"
	trackingInputs[entryBeginTS].Focus()
	trackingInputs[entryBeginTS].CharLimit = len(string(timeFormat))
	trackingInputs[entryBeginTS].SetWidth(30)

	trackingInputs[entryEndTS] = textinput.New()
	trackingInputs[entryEndTS].Placeholder = "12:30pm"
	trackingInputs[entryEndTS].Focus()
	trackingInputs[entryEndTS].CharLimit = len(string(timeFormat))
	trackingInputs[entryEndTS].SetWidth(30)

	trackingInputs[entryComment] = textinput.New()
	trackingInputs[entryComment].Placeholder = "Your comment goes here"
	trackingInputs[entryComment].Focus()
	trackingInputs[entryComment].CharLimit = 255
	trackingInputs[entryComment].SetWidth(60)

	m := Model{
		theme:             thm,
		styles:            styles,
		db:                db,
		worklogStore:      worklogStore,
		jiraSvc:           jiraSvc,
		issueStore:        issueStore,
		opts:              opts,
		issueList:         list.New(stackItems, newItemDelegate(thm, styles, thm.Accent1), listWidth, 0),
		issueMap:          make(map[string]*d.Issue),
		issueIndexMap:     make(map[string]int),
		worklogList:       list.New(worklogListItems, newItemDelegate(thm, styles, thm.Accent2), listWidth, 0),
		syncedWorklogList: list.New(syncedWorklogListItems, newItemDelegate(thm, styles, thm.Accent4), listWidth, 0),
		showHelpIndicator: true,
		trackingInputs:    trackingInputs,
		debug:             debug,
	}
	m.issueList.Title = issueListFetchingTitle
	m.issueList.SetStatusBarItemName("issue", "issues")
	m.issueList.DisableQuitKeybindings()
	m.issueList.SetShowHelp(false)
	m.issueList.Styles.Title = styles.issueListUnfetchedTitle
	m.issueList.KeyMap.PrevPage.SetKeys("left", "h", "pgup")
	m.issueList.KeyMap.NextPage.SetKeys("right", "l", "pgdown")

	m.worklogList.Title = "▫▪▫ Worklog Entries"
	m.worklogList.SetStatusBarItemName("entry", "entries")
	m.worklogList.SetFilteringEnabled(false)
	m.worklogList.DisableQuitKeybindings()
	m.worklogList.SetShowHelp(false)
	m.worklogList.Styles.Title = styles.worklogListTitle
	m.worklogList.KeyMap.PrevPage.SetKeys("left", "h", "pgup")
	m.worklogList.KeyMap.NextPage.SetKeys("right", "l", "pgdown")

	m.syncedWorklogList.Title = "▫▫▪ Synced Entries"
	m.syncedWorklogList.SetStatusBarItemName("entry", "entries")
	m.syncedWorklogList.SetFilteringEnabled(false)
	m.syncedWorklogList.DisableQuitKeybindings()
	m.syncedWorklogList.SetShowHelp(false)
	m.syncedWorklogList.Styles.Title = styles.syncedWorklogListTitle
	m.syncedWorklogList.KeyMap.PrevPage.SetKeys("left", "h", "pgup")
	m.syncedWorklogList.KeyMap.NextPage.SetKeys("right", "l", "pgdown")

	return m
}
