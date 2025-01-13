package ui

import (
	"hash/fnv"

	"github.com/charmbracelet/lipgloss"
)

const (
	defaultBackgroundColor  = "#282828"
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
	issueStatusColor        = "#665c54"
	toolNameColor           = "#b8bb26"
	needsCommentColor       = "#fb4934"
	syncedColor             = "#b8bb26"
	syncingColor            = "#fabd2f"
	notSyncedColor          = "#928374"
	formFieldNameColor      = "#8ec07c"
	formContextColor        = "#fabd2f"
	formHelpColor           = "#928374"
	aggTimeSpentColor       = "#928374"
	initialHelpMsgColor     = "#83a598"
	helpMsgColor            = "#7c6f64"
	helpViewTitleColor      = "#83a598"
	helpHeaderColor         = "#83a598"
	helpSectionColor        = "#fabd2f"
	fallbackIssueColor      = "#ada7ff"
	fallbackAssigneeColor   = "#ccccff"
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

	baseListStyle = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2).
			PaddingBottom(1)

	viewPortStyle = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2).
			PaddingBottom(1)

	listStyle = baseListStyle

	modeStyle = baseStyle.
			Align(lipgloss.Center).
			Bold(true).
			Background(lipgloss.Color(toolNameColor))

	statusStyle = baseStyle.
			Bold(true).
			Align(lipgloss.Center).
			Width(18)

	needsCommentStyle = statusStyle.
				Background(lipgloss.Color(needsCommentColor))

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
		baseStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(defaultBackgroundColor)).
			Bold(true).
			Align(lipgloss.Center).
			Width(20)

		h := fnv.New32()
		_, err := h.Write([]byte(issueType))
		if err != nil {
			return baseStyle.Background(lipgloss.Color(fallbackIssueColor))
		}
		hash := h.Sum32()

		color := issueTypeColors[hash%uint32(len(issueTypeColors))]
		return baseStyle.Background(lipgloss.Color(color))
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
		_, err := h.Write([]byte(assignee))
		if err != nil {
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(fallbackAssigneeColor))
		}
		hash := h.Sum32()

		color := assigneeColors[int(hash)%len(assigneeColors)]

		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(color))
	}

	issueStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(issueStatusColor))

	aggTimeSpentStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color(aggTimeSpentColor))

	unsyncedCountStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Bold(true).
				Foreground(lipgloss.Color(unsyncedCountColor))

	initialHelpMsgStyle = helpMsgStyle.
				Foreground(lipgloss.Color(initialHelpMsgColor))

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
