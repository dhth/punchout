package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/dhth/punchout/internal/ui/theme"
)

type styles struct {
	userMsgInfo             lipgloss.Style
	userMsgError            lipgloss.Style
	helpMsg                 lipgloss.Style
	initialHelpMsg          lipgloss.Style
	list                    lipgloss.Style
	viewPort                lipgloss.Style
	footer                  lipgloss.Style
	mode                    lipgloss.Style
	workLogEntryHeading     lipgloss.Style
	worklogCommentLabel     lipgloss.Style
	formContext             lipgloss.Style
	formFieldName           lipgloss.Style
	formHelp                lipgloss.Style
	tracking                lipgloss.Style
	activeIssueKeyMsg       lipgloss.Style
	activeIssueSummaryMsg   lipgloss.Style
	trackingBegan           lipgloss.Style
	unsyncedCount           lipgloss.Style
	helpTitle               lipgloss.Style
	helpHeader              lipgloss.Style
	helpSection             lipgloss.Style
	wLFormOK                lipgloss.Style
	wLFormError             lipgloss.Style
	wLFormWarning           lipgloss.Style
	issueListTitle          lipgloss.Style
	issueListUnfetchedTitle lipgloss.Style
	issueListFailureTitle   lipgloss.Style
	worklogListTitle        lipgloss.Style
	syncedWorklogListTitle  lipgloss.Style
	fallbackCommentBadge    lipgloss.Style
	syncedBadge             lipgloss.Style
	syncingBadge            lipgloss.Style
	notSyncedBadge          lipgloss.Style
	issueTypeBadge          lipgloss.Style
	issueStatus             lipgloss.Style
	aggTimeSpent            lipgloss.Style
}

func newStyles(thm theme.Theme) styles {
	background := lipgloss.Color(thm.Background)
	accent1 := lipgloss.Color(thm.Accent1)
	accent2 := lipgloss.Color(thm.Accent2)
	accent3 := lipgloss.Color(thm.Accent3)
	accent4 := lipgloss.Color(thm.Accent4)
	accent5 := lipgloss.Color(thm.Accent5)
	accent6 := lipgloss.Color(thm.Accent6)
	success := lipgloss.Color(thm.Success)
	danger := lipgloss.Color(thm.Danger)
	muted := lipgloss.Color(thm.Muted)

	baseBadge := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		Foreground(background)

	listTitle := baseBadge.Bold(true)

	baseHeading := baseBadge.Bold(true)

	tracking := lipgloss.NewStyle().
		PaddingLeft(2).
		Bold(true).
		Foreground(accent1)

	statusBadge := baseBadge.
		Bold(true).
		Align(lipgloss.Center).
		Width(14)

	return styles{
		userMsgInfo:  lipgloss.NewStyle().Foreground(accent5),
		userMsgError: lipgloss.NewStyle().Foreground(danger),
		helpMsg: lipgloss.NewStyle().
			PaddingLeft(2).
			Bold(true).
			Foreground(accent3),
		initialHelpMsg: lipgloss.NewStyle().
			PaddingLeft(2).
			Bold(true).
			Foreground(accent6),
		list: lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2).
			PaddingBottom(1),
		viewPort: lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2).
			PaddingBottom(1),
		footer: lipgloss.NewStyle().
			Foreground(background).
			Background(muted),
		mode: baseBadge.
			Align(lipgloss.Center).
			Bold(true).
			Background(accent4),
		workLogEntryHeading: baseHeading.Background(accent2),
		worklogCommentLabel: lipgloss.NewStyle().Foreground(accent3),
		formContext:         lipgloss.NewStyle().Foreground(accent2),
		formFieldName:       lipgloss.NewStyle().Foreground(accent4),
		formHelp:            lipgloss.NewStyle().Foreground(muted),
		tracking:            tracking,
		activeIssueKeyMsg: tracking.
			PaddingLeft(1).
			Foreground(accent5),
		activeIssueSummaryMsg: tracking.
			PaddingLeft(1).
			Foreground(accent6),
		trackingBegan: tracking.
			PaddingLeft(1).
			Foreground(accent2),
		unsyncedCount: lipgloss.NewStyle().
			PaddingLeft(2).
			Bold(true).
			Foreground(accent2),
		helpTitle: baseBadge.
			Bold(true).
			Background(accent3).
			Align(lipgloss.Left),
		helpHeader:              lipgloss.NewStyle().Bold(true).Foreground(accent3),
		helpSection:             lipgloss.NewStyle().Foreground(accent2),
		wLFormOK:                lipgloss.NewStyle().Foreground(success),
		wLFormError:             lipgloss.NewStyle().Foreground(danger),
		wLFormWarning:           lipgloss.NewStyle().Foreground(accent1),
		issueListTitle:          listTitle.Background(accent1),
		issueListUnfetchedTitle: listTitle.Background(muted),
		issueListFailureTitle:   listTitle.Background(danger),
		worklogListTitle:        listTitle.Background(accent2),
		syncedWorklogListTitle:  listTitle.Background(accent4),
		fallbackCommentBadge: statusBadge.
			Width(20).
			MarginLeft(2).
			Background(accent3),
		syncedBadge:    statusBadge.Background(success),
		syncingBadge:   statusBadge.Background(accent2),
		notSyncedBadge: statusBadge.Background(muted),
		issueTypeBadge: lipgloss.NewStyle().
			Foreground(background).
			Bold(true).
			Align(lipgloss.Center).
			Width(20),
		issueStatus:  lipgloss.NewStyle().Foreground(muted),
		aggTimeSpent: lipgloss.NewStyle().PaddingLeft(2).Foreground(muted),
	}
}
