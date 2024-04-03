package ui

import (
	"github.com/charmbracelet/lipgloss"
	"hash/fnv"
)

var (
	baseStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("#282828"))

	helpMsgStyle = baseStyle.Copy().
			Bold(true).
			Foreground(lipgloss.Color("#83a598"))

	baseListStyle = lipgloss.NewStyle().PaddingTop(1).PaddingRight(2).PaddingLeft(1).PaddingBottom(1)
	viewPortStyle = baseListStyle.Copy().Width(150).PaddingLeft(4)

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

	activeIssueMsgStyle = baseStyle.Copy().
				Bold(true).
				Foreground(lipgloss.Color("#d3869b"))

	helpTitleStyle = baseStyle.Copy().
			Bold(true).
			Background(lipgloss.Color("#8ec07c")).
			Align(lipgloss.Left)

	issueTypeColors = []string{"#928374", "#d3869b", "#fabd2f", "#8ec07c", "#83a598", "#b8bb26", "#fe8019"}

	getIssueTypeStyle = func(issueType string) lipgloss.Style {
		h := fnv.New32()
		h.Write([]byte(issueType))
		hash := h.Sum32()

		color := issueTypeColors[int(hash)%len(issueTypeColors)]
		return lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("#282828")).Copy().
			Bold(true).
			Align(lipgloss.Center).
			Width(18).
			Background(lipgloss.Color(color))
	}
)
