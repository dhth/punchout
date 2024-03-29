package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	listWidth = 140
)

func (m model) View() string {
	var content string
	var footer string

	var statusBar string
	if m.message != "" {
		statusBar = Trim(m.message, 120)
	}
	var errorMsg string
	if m.errorMessage != "" {
		errorMsg = "error: " + Trim(m.errorMessage, 120)
	}
	var activeMsg string
	if m.activeIssue != "" {
		activeMsg = activeIssueMsgStyle.Render("tracking: " + m.activeIssue + " ⚡️")
	}

	switch m.activeView {
	case IssueListView:
		content = stackListStyle.Render(m.issueList.View())
	case WorklogView:
		content = stackListStyle.Render(m.worklogList.View())
	case AskForCommentView:
		content = fmt.Sprintf(
			`
    %s

    %s


    %s
`,
			formFieldNameStyle.Render(RightPadTrim("Comment:", 16)),
			m.trackingInputs[entryComment].View(),
			formContextStyle.Render("Press ctrl+d to submit"),
		)
		for i := 0; i < m.terminalHeight-20+10; i++ {
			content += "\n"
		}
	case EntryTrackingView:

		content = fmt.Sprintf(
			`
    %s

    %s

    %s

    %s

    %s

    %s

    %s


    %s
`,
			formContextStyle.Render("Adding a manual entry. Entry the following details:"),
			formFieldNameStyle.Render("Begin Time  (format: 2006/01/02 15:04)"),
			m.trackingInputs[entryBeginTS].View(),
			formFieldNameStyle.Render("End Time  (format: 2006/01/02 15:04)"),
			m.trackingInputs[entryEndTS].View(),
			formFieldNameStyle.Render(RightPadTrim("Comment:", 16)),
			m.trackingInputs[entryComment].View(),
			formContextStyle.Render("Press ctrl+d to submit"),
		)
		for i := 0; i < m.terminalHeight-20; i++ {
			content += "\n"
		}
	case HelpView:
		if !m.helpVPReady {
			content = "\n  Initializing..."
		} else {
			content = viewPortStyle.Render(fmt.Sprintf("  %s\n\n%s\n", helpTitleStyle.Render("Help"), m.helpVP.View()))
		}
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282828")).
		Background(lipgloss.Color("#7c6f64"))

	var helpMsg string
	if m.showHelpIndicator {
		helpMsg = " " + helpMsgStyle.Render("Press ? for help")
	}

	footerStr := fmt.Sprintf("%s%s%s%s",
		modeStyle.Render("punchout"),
		helpMsg,
		activeMsg,
		errorMsg,
	)
	footer = footerStyle.Render(footerStr)

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		statusBar,
		footer,
	)
}
