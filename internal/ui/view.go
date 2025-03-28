package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	c "github.com/dhth/punchout/internal/common"
)

var listWidth = 140

func (m Model) View() string {
	var content string
	var footer string

	var statusBar string
	var helpMsg string
	if m.message != "" {
		statusBar = c.Trim(m.message, 120)
	}
	var activeMsg string

	var fallbackCommentMsg string
	if m.fallbackComment != nil {
		fallbackCommentMsg = " (a fallback is configured)"
	}

	if m.issuesFetched {
		if m.activeIssue != "" {
			var issueSummaryMsg, trackingSinceMsg string
			issue, ok := m.issueMap[m.activeIssue]
			if ok {
				issueSummaryMsg = fmt.Sprintf("(%s)", c.Trim(issue.Summary, 50))
				if m.activeView != saveActiveWLView {
					trackingSinceMsg = fmt.Sprintf("(since %s)", m.activeIssueBeginTS.Format(timeOnlyFormat))
				}
			}
			activeMsg = fmt.Sprintf("%s%s%s%s",
				trackingStyle.Render("tracking:"),
				activeIssueKeyMsgStyle.Render(m.activeIssue),
				activeIssueSummaryMsgStyle.Render(issueSummaryMsg),
				trackingBeganStyle.Render(trackingSinceMsg),
			)
		}

		if m.showHelpIndicator {
			// first time help
			if m.activeView == issueListView && len(m.syncedWorklogList.Items()) == 0 && m.unsyncedWLCount == 0 {
				if m.trackingActive {
					helpMsg += initialHelpMsgStyle.Render("Press s to stop tracking time")
				} else {
					helpMsg += initialHelpMsgStyle.Render("Press s to start tracking time")
				}
			}
		}
	}

	formHeadingText := "Enter/update the following details:"
	formHelp := "Use tab/shift-tab to move between sections; esc to go back."
	formBeginTimeHelp := "Begin Time* (format: 2006/01/02 15:04)"
	formEndTimeHelp := "End Time* (format: 2006/01/02 15:04)"
	formTimeShiftHelp := "(k/j/K/J moves time, when correct)"
	formCommentHelp := fmt.Sprintf("Comment%s", fallbackCommentMsg)
	formSubmitHelp := "Press enter to submit"

	switch m.activeView {
	case issueListView:
		content = listStyle.Render(m.issueList.View())
	case wLView:
		content = listStyle.Render(m.worklogList.View())
	case syncedWLView:
		content = listStyle.Render(m.syncedWorklogList.View())
	case editActiveWLView:
		content = fmt.Sprintf(
			`
  %s

  %s

  %s

  %s

  %s    %s

  %s

  %s


  %s
`,
			workLogEntryHeadingStyle.Render("Edit Active Worklog"),
			formContextStyle.Render(formHeadingText),
			formHelpStyle.Render(formHelp),
			formFieldNameStyle.Render(formBeginTimeHelp),
			m.trackingInputs[entryBeginTS].View(),
			formHelpStyle.Render(formTimeShiftHelp),
			formFieldNameStyle.Render(formCommentHelp),
			m.trackingInputs[entryComment].View(),
			formContextStyle.Render(formSubmitHelp),
		)
		for i := 0; i < m.terminalHeight-20; i++ {
			content += "\n"
		}
	case saveActiveWLView:
		content = fmt.Sprintf(
			`
  %s

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
			workLogEntryHeadingStyle.Render("Save Worklog"),
			formContextStyle.Render(formHeadingText),
			formHelpStyle.Render(formHelp),
			formFieldNameStyle.Render(formBeginTimeHelp),
			m.trackingInputs[entryBeginTS].View(),
			formHelpStyle.Render(formTimeShiftHelp),
			formFieldNameStyle.Render(formEndTimeHelp),
			m.trackingInputs[entryEndTS].View(),
			formHelpStyle.Render(formTimeShiftHelp),
			formFieldNameStyle.Render(formCommentHelp),
			m.trackingInputs[entryComment].View(),
			formContextStyle.Render(formSubmitHelp),
		)
		for i := 0; i < m.terminalHeight-24; i++ {
			content += "\n"
		}
	case wlEntryView:
		var formHeading string
		switch m.worklogSaveType {
		case worklogInsert:
			formHeading = "Save Worklog (manual)"
		case worklogUpdate:
			formHeading = "Update Worklog"
		}

		content = fmt.Sprintf(
			`
  %s

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
			workLogEntryHeadingStyle.Render(formHeading),
			formContextStyle.Render(formHeadingText),
			formHelpStyle.Render(formHelp),
			formFieldNameStyle.Render(formBeginTimeHelp),
			m.trackingInputs[entryBeginTS].View(),
			formHelpStyle.Render(formTimeShiftHelp),
			formFieldNameStyle.Render(formEndTimeHelp),
			m.trackingInputs[entryEndTS].View(),
			formHelpStyle.Render(formTimeShiftHelp),
			formFieldNameStyle.Render(formCommentHelp),
			m.trackingInputs[entryComment].View(),
			formContextStyle.Render(formSubmitHelp),
		)
		for i := 0; i < m.terminalHeight-24; i++ {
			content += "\n"
		}
	case helpView:
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
		helpMsg += helpMsgStyle.Render("Press ? for help")
	}

	var unsyncedMsg string
	if m.unsyncedWLCount > 0 {
		entryWord := "entries"
		if m.unsyncedWLCount == 1 {
			entryWord = "entry"
		}
		unsyncedTimeMsg := c.HumanizeDuration(m.unsyncedWLSecsSpent)
		unsyncedMsg = unsyncedCountStyle.Render(fmt.Sprintf("%d unsynced %s (%s)", m.unsyncedWLCount, entryWord, unsyncedTimeMsg))
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
