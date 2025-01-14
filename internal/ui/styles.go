package ui

import (
	"github.com/charmbracelet/lipgloss"
	c "github.com/dhth/punchout/internal/common"
)

const (
	issueListUnfetchedColor = "#928374"
	failureColor            = "#fb4934"
	issueListColor          = "#fe8019"
	worklogListColor        = "#fabd2f"
	syncedWorklogListColor  = "#b8bb26"
	trackingColor           = "#fe8019"
	unsyncedCountColor      = "#fabd2f"
	activeIssueKeyColor     = "#d3869b"
	activeIssueSummaryColor = "#8ec07c"
	trackingBeganColor      = "#fabd2f"
	toolNameColor           = "#b8bb26"
	formFieldNameColor      = "#8ec07c"
	formContextColor        = "#fabd2f"
	formHelpColor           = "#928374"
	initialHelpMsgColor     = "#83a598"
	helpMsgColor            = "#7c6f64"
	helpViewTitleColor      = "#83a598"
	helpHeaderColor         = "#83a598"
	helpSectionColor        = "#fabd2f"
)

var (
	helpMsgStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Bold(true).
			Foreground(lipgloss.Color(helpMsgColor))

	baseListStyle = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2).
			PaddingBottom(1)

	viewPortStyle = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2).
			PaddingBottom(1)

	listStyle = baseListStyle

	modeStyle = c.BaseStyle.
			Align(lipgloss.Center).
			Bold(true).
			Background(lipgloss.Color(toolNameColor))

	formContextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(formContextColor))

	formFieldNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(formFieldNameColor))

	formHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(formHelpColor))

	trackingStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Bold(true).
			Foreground(lipgloss.Color(trackingColor))

	activeIssueKeyMsgStyle = trackingStyle.
				PaddingLeft(1).
				Foreground(lipgloss.Color(activeIssueKeyColor))

	activeIssueSummaryMsgStyle = trackingStyle.
					PaddingLeft(1).
					Foreground(lipgloss.Color(activeIssueSummaryColor))

	trackingBeganStyle = trackingStyle.
				PaddingLeft(1).
				Foreground(lipgloss.Color(trackingBeganColor))

	unsyncedCountStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Bold(true).
				Foreground(lipgloss.Color(unsyncedCountColor))

	initialHelpMsgStyle = helpMsgStyle.
				Foreground(lipgloss.Color(initialHelpMsgColor))

	helpTitleStyle = c.BaseStyle.
			Bold(true).
			Background(lipgloss.Color(helpViewTitleColor)).
			Align(lipgloss.Left)

	helpHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(helpHeaderColor))

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(helpSectionColor))
)
