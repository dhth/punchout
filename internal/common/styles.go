package common

import (
	"hash/fnv"

	"github.com/charmbracelet/lipgloss"
)

const (
	DefaultBackgroundColor = "#282828"
	issueStatusColor       = "#665c54"
	FallbackCommentColor   = "#83a598"
	syncedColor            = "#b8bb26"
	syncingColor           = "#fabd2f"
	notSyncedColor         = "#928374"
	aggTimeSpentColor      = "#928374"
	fallbackIssueColor     = "#ada7ff"
	fallbackAssigneeColor  = "#ccccff"
)

var (
	BaseStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color(DefaultBackgroundColor))

	statusStyle = BaseStyle.
			Bold(true).
			Align(lipgloss.Center).
			Width(14)

	usingFallbackCommentStyle = statusStyle.
					Width(20).
					MarginLeft(2).
					Background(lipgloss.Color(FallbackCommentColor))

	syncedStyle = statusStyle.
			Background(lipgloss.Color(syncedColor))

	syncingStyle = statusStyle.
			Background(lipgloss.Color(syncingColor))

	notSyncedStyle = statusStyle.
			Background(lipgloss.Color(notSyncedColor))

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
			Foreground(lipgloss.Color(DefaultBackgroundColor)).
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
)
