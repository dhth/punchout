package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dhth/punchout/internal/utils"
)

const wLWarningThresholdSecs = 8 * 60 * 60

var listWidth = 140

type wlFormValidity uint

const (
	wlSubmitOk wlFormValidity = iota
	wlSubmitWarn
	wlSubmitErr
)

func (m Model) View() tea.View {
	var content string
	var footer string

	var statusBar string
	var helpMsg string
	if m.message.isActive() {
		if m.message.kind == userMsgError {
			statusBar = m.styles.userMsgError.Render(m.message.value)
		} else {
			statusBar = m.styles.userMsgInfo.Render(m.message.value)
		}
	}
	var activeMsg string

	var fallbackCommentMsg string
	if m.opts.Jira.FallbackComment != nil {
		fallbackCommentMsg = " (a fallback is configured)"
	}

	if m.issuesFetched {
		if m.activeIssue != "" {
			var issueSummaryMsg, trackingSinceMsg string
			issue, ok := m.issueMap[m.activeIssue]
			if ok {
				issueSummaryMsg = fmt.Sprintf("(%s)", utils.Trim(issue.Summary, 50))
				if m.activeView != saveActiveWLView {
					trackingSinceMsg = fmt.Sprintf("(since %s)", m.activeIssueBeginTS.Format(timeOnlyFormat))
				}
			}
			activeMsg = fmt.Sprintf(
				"%s%s%s%s",
				m.styles.tracking.Render("tracking:"),
				m.styles.activeIssueKeyMsg.Render(m.activeIssue),
				m.styles.activeIssueSummaryMsg.Render(issueSummaryMsg),
				m.styles.trackingBegan.Render(trackingSinceMsg),
			)
		}

		if m.showHelpIndicator {
			// first time help
			if m.activeView == issueListView && len(m.syncedWorklogList.Items()) == 0 && m.unsyncedWLCount == 0 {
				if m.trackingActive {
					helpMsg += m.styles.initialHelpMsg.Render("Press s to stop tracking time")
				} else {
					helpMsg += m.styles.initialHelpMsg.Render("Press s to start tracking time")
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

	var submissionCtx string
	var submissionValidity wlFormValidity
	var durationCtx string
	if m.activeView == saveActiveWLView || m.activeView == wlEntryView {
		durationCtx, submissionValidity = getDurationValidityContext(m.trackingInputs[entryBeginTS].Value(), m.trackingInputs[entryEndTS].Value())

		switch submissionValidity {
		case wlSubmitOk:
			submissionCtx = m.styles.wLFormOK.Render(durationCtx)
		case wlSubmitWarn:
			submissionCtx = m.styles.wLFormWarning.Render(durationCtx)
		case wlSubmitErr:
			submissionCtx = m.styles.wLFormError.Render(durationCtx)
		}
	}

	var formSubmitHelp string
	if submissionValidity != wlSubmitErr {
		formSubmitHelp = m.styles.formContext.Render("Press enter to submit")
	}

	switch m.activeView {
	case issueListView:
		content = m.styles.list.Render(m.issueList.View())
	case wLView:
		content = m.styles.list.Render(m.worklogList.View())
	case syncedWLView:
		content = m.styles.list.Render(m.syncedWorklogList.View())
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
			m.styles.workLogEntryHeading.Render("Edit Active Worklog"),
			m.styles.formContext.Render(formHeadingText),
			m.styles.formHelp.Render(formHelp),
			m.styles.formFieldName.Render(formBeginTimeHelp),
			m.trackingInputs[entryBeginTS].View(),
			m.styles.formHelp.Render(formTimeShiftHelp),
			m.styles.formFieldName.Render(formCommentHelp),
			m.trackingInputs[entryComment].View(),
			formSubmitHelp,
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

  %s
`,
			m.styles.workLogEntryHeading.Render("Save Worklog"),
			m.styles.formContext.Render(formHeadingText),
			m.styles.formHelp.Render(formHelp),
			m.styles.formFieldName.Render(formBeginTimeHelp),
			m.trackingInputs[entryBeginTS].View(),
			m.styles.formHelp.Render(formTimeShiftHelp),
			m.styles.formFieldName.Render(formEndTimeHelp),
			m.trackingInputs[entryEndTS].View(),
			m.styles.formHelp.Render(formTimeShiftHelp),
			m.styles.formFieldName.Render(formCommentHelp),
			m.trackingInputs[entryComment].View(),
			submissionCtx,
			formSubmitHelp,
		)
		for i := 0; i < m.terminalHeight-26; i++ {
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

  %s
`,
			m.styles.workLogEntryHeading.Render(formHeading),
			m.styles.formContext.Render(formHeadingText),
			m.styles.formHelp.Render(formHelp),
			m.styles.formFieldName.Render(formBeginTimeHelp),
			m.trackingInputs[entryBeginTS].View(),
			m.styles.formHelp.Render(formTimeShiftHelp),
			m.styles.formFieldName.Render(formEndTimeHelp),
			m.trackingInputs[entryEndTS].View(),
			m.styles.formHelp.Render(formTimeShiftHelp),
			m.styles.formFieldName.Render(formCommentHelp),
			m.trackingInputs[entryComment].View(),
			submissionCtx,
			formSubmitHelp,
		)
		for i := 0; i < m.terminalHeight-26; i++ {
			content += "\n"
		}
	case helpView:
		if !m.helpVPReady {
			content = "\n  Initializing..."
		} else {
			content = m.styles.viewPort.Render(fmt.Sprintf("  %s\n\n%s\n", m.styles.helpTitle.Render("Help"), m.helpVP.View()))
		}
	}

	if m.showHelpIndicator {
		helpMsg += m.styles.helpMsg.Render("Press ? for help")
	}

	var unsyncedMsg string
	if m.unsyncedWLCount > 0 {
		entryWord := "entries"
		if m.unsyncedWLCount == 1 {
			entryWord = "entry"
		}
		unsyncedTimeMsg := utils.HumanizeDuration(m.unsyncedWLSecsSpent)
		unsyncedMsg = m.styles.unsyncedCount.Render(fmt.Sprintf("%d unsynced %s (%s)", m.unsyncedWLCount, entryWord, unsyncedTimeMsg))
	}

	footerStr := fmt.Sprintf(
		"%s%s%s%s",
		m.styles.mode.Render("punchout"),
		helpMsg,
		unsyncedMsg,
		activeMsg,
	)
	footer = m.styles.footer.Render(footerStr)

	v := tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		statusBar,
		footer,
	))
	v.AltScreen = true
	v.BackgroundColor = lipgloss.Color(m.theme.Background)
	v.ForegroundColor = lipgloss.Color(m.theme.Foreground)

	return v
}

func getDurationValidityContext(beginStr, endStr string) (string, wlFormValidity) {
	if strings.TrimSpace(beginStr) == "" {
		return "Begin time is empty", wlSubmitErr
	}

	if strings.TrimSpace(endStr) == "" {
		return "End time is empty", wlSubmitErr
	}

	beginTS, err := time.ParseInLocation(timeFormat, beginStr, time.Local)
	if err != nil {
		return "Begin time is invalid", wlSubmitErr
	}

	endTS, err := time.ParseInLocation(timeFormat, endStr, time.Local)
	if err != nil {
		return "End time is invalid", wlSubmitErr
	}

	dur := endTS.Sub(beginTS)

	if dur == 0 {
		return "You're recording no time, change begin and/or end time", wlSubmitErr
	}

	if dur < 0 {
		return "End time is before start time", wlSubmitErr
	}

	totalSeconds := int(dur.Seconds())

	humanized := utils.HumanizeDuration(totalSeconds)
	msg := fmt.Sprintf("You're recording %s", humanized)
	if totalSeconds > wLWarningThresholdSecs {
		return msg, wlSubmitWarn
	}

	return msg, wlSubmitOk
}
