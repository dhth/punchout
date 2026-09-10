package ui

import (
	"fmt"
	"hash/fnv"
	"image/color"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	d "github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/dhth/punchout/internal/utils"
)

const (
	dayAndTimeFormat  = "Mon, 15:04"
	dateAndTimeFormat = "Jan 2, 15:04"
)

type worklogListItem struct {
	d.StoredWorklog

	fallbackCommentUsed bool
	syncInProgress      bool
	err                 error
}

func (item worklogListItem) FilterValue() string { return item.IssueKey }

type syncedWorklogListItem struct {
	d.StoredWorklog
}

func (item syncedWorklogListItem) FilterValue() string { return item.IssueKey }

func renderListItem(
	item list.Item,
	thm theme.Theme,
	styles styles,
	issueMap map[string]*d.Issue,
	fallbackCommentConfigured,
	selected bool,
	width int,
	now time.Time,
) (string, string) {
	columnWidth := max(0, width) / 5

	switch item := item.(type) {
	case *d.Issue:
		return renderIssue(item, thm, styles, columnWidth)
	case worklogListItem:
		return renderUnsyncedWorklog(
			item,
			styles,
			issueMap,
			fallbackCommentConfigured,
			selected,
			columnWidth,
			now,
		)
	case syncedWorklogListItem:
		return renderSyncedWorklog(item, styles, issueMap, selected, columnWidth, now)
	default:
		return "", ""
	}
}

func renderIssue(
	issue *d.Issue,
	thm theme.Theme,
	styles styles,
	columnWidth int,
) (string, string) {
	var trackingIndicator string
	if issue.TrackingActive {
		trackingIndicator = "⏲ "
	}
	title := trackingIndicator + issue.Summary

	issueTypeColor := categoricalColor("issue-type", issue.IssueType, thm.CategoricalColors)
	issueType := renderBadge(
		issue.IssueType,
		styles.issueTypeBadge.Background(issueTypeColor),
	)

	assignee := issue.Assignee
	if issue.Assignee != "" {
		assigneeColor := categoricalColor("assignee", issue.Assignee, thm.CategoricalColors)
		assignee = lipgloss.NewStyle().Foreground(assigneeColor).Render(assignee)
	}

	status := styles.issueStatus.Render(issue.Status)

	var totalTimeSpent string
	if issue.AggSecondsSpent > 0 {
		totalTimeSpent = styles.aggTimeSpent.Render(utils.HumanizeDuration(issue.AggSecondsSpent))
	}

	description := strings.Join([]string{
		renderColumn(issue.IssueKey, columnWidth),
		renderColumn(status, columnWidth),
		renderColumn(assignee, columnWidth),
		issueType + totalTimeSpent,
	}, "")

	return title, description
}

func renderUnsyncedWorklog(
	entry worklogListItem,
	styles styles,
	issueMap map[string]*d.Issue,
	fallbackCommentConfigured,
	selected bool,
	columnWidth int,
	now time.Time,
) (string, string) {
	showComment := !entry.fallbackCommentUsed
	title := renderWorklogTitle(
		entry.StoredWorklog,
		styles.worklogCommentLabel,
		issueMap,
		showComment,
		selected,
	)

	if entry.err != nil {
		return title, "error: " + entry.err.Error()
	}

	timeSpent := utils.HumanizeDuration(entry.SecsSpent())

	var syncStatus string
	switch {
	case entry.Synced:
		syncStatus = styles.syncedBadge.Render("synced")
	case entry.syncInProgress:
		syncStatus = styles.syncingBadge.Render("syncing")
	default:
		syncStatus = styles.notSyncedBadge.Render("not synced")
	}

	var fallbackCommentStatus string
	if entry.fallbackCommentUsed || (entry.NeedsComment() && fallbackCommentConfigured) {
		fallbackCommentStatus = styles.fallbackCommentBadge.Render("fallback comment")
	}

	description := strings.Join([]string{
		renderColumn(entry.IssueKey, columnWidth),
		renderColumn(formatWorklogTimeRange(now, entry.BeginTS, entry.EndTS), columnWidth),
		renderColumn(fmt.Sprintf("(%s)", timeSpent), columnWidth),
		syncStatus + fallbackCommentStatus,
	}, "")

	return title, description
}

func renderSyncedWorklog(
	entry syncedWorklogListItem,
	styles styles,
	issueMap map[string]*d.Issue,
	selected bool,
	columnWidth int,
	now time.Time,
) (string, string) {
	title := renderWorklogTitle(
		entry.StoredWorklog,
		styles.worklogCommentLabel,
		issueMap,
		true,
		selected,
	)

	description := strings.Join([]string{
		renderColumn(entry.IssueKey, columnWidth),
		renderColumn(formatWorklogTimeRange(now, entry.BeginTS, entry.EndTS), columnWidth),
		fmt.Sprintf("(%s)", utils.HumanizeDuration(int(entry.EndTS.Sub(entry.BeginTS).Seconds()))),
	}, "")

	return title, description
}

func formatWorklogTimeRange(now, start, end time.Time) string {
	start = start.In(now.Location())
	end = end.In(now.Location())
	nowYear, nowWeek := now.ISOWeek()

	sameDay := func(a, b time.Time) bool {
		return a.Year() == b.Year() && a.YearDay() == b.YearDay()
	}
	formatDatedTime := func(value time.Time) string {
		year, week := value.ISOWeek()
		if year == nowYear && week == nowWeek {
			return value.Format(dayAndTimeFormat)
		}
		return value.Format(dateAndTimeFormat)
	}

	if sameDay(start, end) {
		if sameDay(start, now) {
			return fmt.Sprintf("%s – %s", start.Format(timeOnlyFormat), end.Format(timeOnlyFormat))
		}
		return fmt.Sprintf("%s – %s", formatDatedTime(start), end.Format(timeOnlyFormat))
	}

	return fmt.Sprintf("%s – %s", formatDatedTime(start), formatDatedTime(end))
}

func renderWorklogTitle(
	entry d.StoredWorklog,
	commentLabelStyle lipgloss.Style,
	issueMap map[string]*d.Issue,
	showComment,
	selected bool,
) string {
	title := "[ISSUE SUMMARY UNAVAILABLE]"
	if issue, ok := issueMap[entry.IssueKey]; ok && issue != nil && strings.TrimSpace(issue.Summary) != "" {
		title = issue.Summary
	}
	if showComment && !entry.NeedsComment() {
		title = "Comment: " + entry.Comment
		if !selected {
			title = commentLabelStyle.Render("Comment:") + " " + entry.Comment
		}
	}

	return title
}

func categoricalColor(category, value string, colors []string) color.Color {
	if len(colors) == 0 {
		return lipgloss.NoColor{}
	}

	h := fnv.New32()
	_, _ = h.Write([]byte(strings.Join([]string{category, value}, ":")))

	return lipgloss.Color(colors[h.Sum32()%uint32(len(colors))])
}

func renderColumn(value string, width int) string {
	if width <= 0 {
		return ""
	}

	contentWidth := width
	if width > 1 {
		contentWidth--
	}
	value = ansi.Truncate(value, contentWidth, "…")

	return value + strings.Repeat(" ", max(0, width-lipgloss.Width(value)))
}

func renderBadge(value string, style lipgloss.Style) string {
	contentWidth := max(
		0,
		style.GetWidth()-style.GetHorizontalFrameSize()+style.GetHorizontalMargins(),
	)
	value = ansi.Truncate(value, contentWidth, "…")

	return style.Render(value)
}
