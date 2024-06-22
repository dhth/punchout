package ui

import (
	"github.com/charmbracelet/lipgloss"
	"hash/fnv"
)

const (
	defaultBackgroundColor  = "#282828"
	issueListUnfetchedColor = "#928374"
	issueListColor          = "#fe8019"
	worklogListColor        = "#fabd2f"
	syncedWorklogListColor  = "#b8bb26"
	trackingColor           = "#fe8019"
	unsyncedCountColor      = "#fabd2f"
	activeIssueKeyColor     = "#d3869b"
	activeIssueSummaryColor = "#8ec07c"
	issueStatusColor        = "#665c54"
	toolNameColor           = "#b8bb26"
	syncedColor             = "#b8bb26"
	syncingColor            = "#fabd2f"
	notSyncedColor          = "#928374"
	formFieldNameColor      = "#8ec07c"
	formContextColor        = "#fabd2f"
	aggTimeSpentColor       = "#928374"
	helpMsgColor            = "#7c6f64"
	helpViewTitleColor      = "#83a598"
	helpHeaderColor         = "#83a598"
	helpSectionColor        = "#fabd2f"
)

var (
	baseStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color(defaultBackgroundColor))

	helpMsgStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Bold(true).
			Foreground(lipgloss.Color(helpMsgColor))

	baseListStyle = lipgloss.NewStyle().PaddingTop(1).PaddingRight(2).PaddingLeft(1).PaddingBottom(1)
	viewPortStyle = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2).
			PaddingLeft(1).
			PaddingBottom(1)

	listStyle = baseListStyle

	modeStyle = baseStyle.
			Align(lipgloss.Center).
			Bold(true).
			Background(lipgloss.Color(toolNameColor))

	statusStyle = baseStyle.
			Bold(true).
			Align(lipgloss.Center).
			Width(14)

	syncedStyle = statusStyle.
			Background(lipgloss.Color(syncedColor))

	syncingStyle = statusStyle.
			Background(lipgloss.Color(syncingColor))

	notSyncedStyle = statusStyle.
			Background(lipgloss.Color(notSyncedColor))

	formContextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(formContextColor))

	formFieldNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(formFieldNameColor))

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

	issueTypeColors = []string{
		"#d3869b",
		"#b5e48c",
		"#90e0ef",
		"#ca7df9",
		"#ada7ff",
		"#bbd0ff",
		"#48cae4",
		"#8187dc",
		"#ffb4a2",
		"#b8bb26",
		"#ffc6ff",
		"#4895ef",
		"#83a598",
		"#fabd2f",
	}

	getIssueTypeStyle = func(issueType string) lipgloss.Style {
		h := fnv.New32()
		h.Write([]byte(issueType))
		hash := h.Sum32()

		color := issueTypeColors[int(hash)%len(issueTypeColors)]
		return lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color(defaultBackgroundColor)).
			Bold(true).
			Align(lipgloss.Center).
			Width(20).
			Background(lipgloss.Color(color))
	}

	assigneeColors = []string{
		"#ccccff", // Lavender Blue
		"#ffa87d", // Light orange
		"#7385D8", // Light blue
		"#fabd2f", // Bright Yellow
		"#00abe5", // Deep Sky
		"#d3691e", // Chocolate
	}
	assigneeStyle = func(assignee string) lipgloss.Style {
		h := fnv.New32()
		h.Write([]byte(assignee))
		hash := h.Sum32()

		color := assigneeColors[int(hash)%len(assigneeColors)]

		st := lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color(color))

		return st
	}

	issueStatusStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(lipgloss.Color(issueStatusColor))

	aggTimeSpentStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color(aggTimeSpentColor))

	unsyncedCountStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Bold(true).
				Foreground(lipgloss.Color(unsyncedCountColor))

	helpTitleStyle = baseStyle.
			Bold(true).
			Background(lipgloss.Color(helpViewTitleColor)).
			Align(lipgloss.Left)

	helpHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(helpHeaderColor))

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(helpSectionColor))
)
