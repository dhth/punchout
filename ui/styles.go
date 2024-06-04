package ui

import (
	"github.com/charmbracelet/lipgloss"
	"hash/fnv"
)

const (
	trackingColor           = "#fabd2f"
	ActiveIssueKeyColor     = "#d3869b"
	ActiveIssueSummaryColor = "#8ec07c"
	IssueStatusColor        = "#665c54"
	AggTimeSpentColor       = "#928374"
	helpMsgColor            = "#83a598"
	helpViewTitleColor      = "#83a598"
	helpHeaderColor         = "#83a598"
	helpSectionColor        = "#fabd2f"
)

var (
	baseStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("#282828"))

	helpMsgStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Bold(true).
			Foreground(lipgloss.Color("#83a598"))

	baseListStyle = lipgloss.NewStyle().PaddingTop(1).PaddingRight(2).PaddingLeft(1).PaddingBottom(1)
	viewPortStyle = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2).
			PaddingLeft(1).
			PaddingBottom(1)

	stackListStyle = baseListStyle.Copy()

	modeStyle = baseStyle.Copy().
			Align(lipgloss.Center).
			Bold(true).
			Background(lipgloss.Color("#b8bb26"))

	statusStyle = baseStyle.Copy().
			Bold(true).
			Align(lipgloss.Center).
			Width(12)

	syncedStyle = statusStyle.Copy().
			Background(lipgloss.Color("#b8bb26"))

	syncingStyle = statusStyle.Copy().
			Background(lipgloss.Color("#83a598"))

	notSyncedStyle = statusStyle.Copy().
			Background(lipgloss.Color("#fb4934"))

	formContextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fe8019"))

	formFieldNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#504945"))

	trackingStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Bold(true).
			Foreground(lipgloss.Color(trackingColor))

	activeIssueKeyMsgStyle = trackingStyle.Copy().
				PaddingLeft(1).
				Foreground(lipgloss.Color(ActiveIssueKeyColor))

	activeIssueSummaryMsgStyle = trackingStyle.Copy().
					PaddingLeft(1).
					Foreground(lipgloss.Color(ActiveIssueSummaryColor))

	issueTypeColors = []string{"#928374", "#d3869b", "#fabd2f", "#8ec07c", "#83a598", "#b8bb26", "#fe8019"}

	getIssueTypeStyle = func(issueType string) lipgloss.Style {
		h := fnv.New32()
		h.Write([]byte(issueType))
		hash := h.Sum32()

		color := issueTypeColors[int(hash)%len(issueTypeColors)]
		return lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color("#282828")).Copy().
			Bold(true).
			Align(lipgloss.Center).
			Width(18).
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
				Foreground(lipgloss.Color(IssueStatusColor))

	aggTimeSpentStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color(AggTimeSpentColor))

	helpTitleStyle = baseStyle.Copy().
			Bold(true).
			Background(lipgloss.Color(helpViewTitleColor)).
			Align(lipgloss.Left)

	helpHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(helpHeaderColor))

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(helpSectionColor))
)
