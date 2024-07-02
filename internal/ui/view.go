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
	var helpMsg string
	if m.message != "" {
		statusBar = Trim(m.message, 120)
	}
	var activeMsg string
	if m.issuesFetched {
		if m.activeIssue != "" {
			var issueSummaryMsg, trackingSinceMsg string
			issue, ok := m.issueMap[m.activeIssue]
			if ok {
				issueSummaryMsg = fmt.Sprintf("(%s)", Trim(issue.summary, 50))
				if m.activeView != AskForCommentView {
					trackingSinceMsg = fmt.Sprintf("(since %s)", m.activeIssueBeginTS.Format(timeOnlyFormat))
				}
			}
			activeMsg = fmt.Sprintf("%s%s%s%s",
				trackingStyle.Render("tracking:"),
				activeIssueKeyMsgStyle.Render(m.activeIssue),
				activeIssueSummaryMsgStyle.Render(issueSummaryMsg),
				trackingBeginStyle.Render(trackingSinceMsg),
			)
		}

		if m.showHelpIndicator {
			// first time help
			if m.activeView == IssueListView && len(m.syncedWorklogList.Items()) == 0 && m.unsyncedWLCount == 0 {
				if m.trackingActive {
					helpMsg += " " + initialHelpMsgStyle.Render("Press s to stop tracking time")
				} else {
					helpMsg += " " + initialHelpMsgStyle.Render("Press s to start tracking time")
				}
			}
		}
	}

	switch m.activeView {
	case IssueListView:
		content = listStyle.Render(m.issueList.View())
	case WorklogView:
		content = listStyle.Render(m.worklogList.View())
	case SyncedWorklogView:
		content = listStyle.Render(m.syncedWorklogList.View())
	case AskForCommentView:
		formHeadingText := "Saving worklog entry. Enter/update the following details:"
		content = fmt.Sprintf(
			`
    %s

    %s

    %s

    %s    %s

    %s

    %s    %s

    %s

    %s


    %s
`,
			formContextStyle.Render(formHeadingText),
			formHelpStyle.Render("Use tab/shift-tab to move between sections; esc to go back."),
			formFieldNameStyle.Render("Begin Time  (format: 2006/01/02 15:04)"),
			m.trackingInputs[entryBeginTS].View(),
			formHelpStyle.Render("(k/j/K/J moves time, when correct)"),
			formFieldNameStyle.Render("End Time  (format: 2006/01/02 15:04)"),
			m.trackingInputs[entryEndTS].View(),
			formHelpStyle.Render("(k/j/K/J moves time, when correct)"),
			formFieldNameStyle.Render(RightPadTrim("Comment:", 16)),
			m.trackingInputs[entryComment].View(),
			formContextStyle.Render("Press enter to submit"),
		)
		for i := 0; i < m.terminalHeight-22; i++ {
			content += "\n"
		}
	case ManualWorklogEntryView:
		var formHeadingText string
		switch m.worklogSaveType {
		case worklogInsert:
			formHeadingText = "Adding a manual entry. Enter the following details:"
		case worklogUpdate:
			formHeadingText = "Updating worklog entry. Enter the following details:"
		}

		content = fmt.Sprintf(
			`
    %s

    %s

    %s

    %s    %s

    %s

    %s    %s

    %s

    %s


    %s
`,
			formContextStyle.Render(formHeadingText),
			formHelpStyle.Render("Use tab/shift-tab to move between sections; esc to go back."),
			formFieldNameStyle.Render("Begin Time  (format: 2006/01/02 15:04)"),
			m.trackingInputs[entryBeginTS].View(),
			formHelpStyle.Render("(k/j/K/J moves time, when correct)"),
			formFieldNameStyle.Render("End Time  (format: 2006/01/02 15:04)"),
			m.trackingInputs[entryEndTS].View(),
			formHelpStyle.Render("(k/j/K/J moves time, when correct)"),
			formFieldNameStyle.Render(RightPadTrim("Comment:", 16)),
			m.trackingInputs[entryComment].View(),
			formContextStyle.Render("Press enter to submit"),
		)
		for i := 0; i < m.terminalHeight-22; i++ {
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

	if m.showHelpIndicator {
		helpMsg += " " + helpMsgStyle.Render("Press ? for help")
	}

	var unsyncedMsg string
	if m.unsyncedWLCount > 0 {
		entryWord := "entries"
		if m.unsyncedWLCount == 1 {
			entryWord = "entry"
		}
		unsyncedMsg = unsyncedCountStyle.Render(fmt.Sprintf("(%d unsynced %s)", m.unsyncedWLCount, entryWord))
	}
	footerStr := fmt.Sprintf("%s%s%s%s",
		modeStyle.Render("punchout"),
		helpMsg,
		unsyncedMsg,
		activeMsg,
	)
	footer = footerStyle.Render(footerStr)

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		statusBar,
		footer,
	)
}
