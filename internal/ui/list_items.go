package ui

import (
	"fmt"
	"hash/fnv"
	"image/color"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	d "github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/dhth/punchout/internal/utils"
	"github.com/dustin/go-humanize"
)

const (
	dayAndTimeFormat = "Mon, 15:04"
	dateFormat       = "2006/01/02"
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

func renderListItem(item list.Item, thm theme.Theme, styles styles, issueMap map[string]*d.Issue, fallbackCommentConfigured, selected bool) (string, string) {
	switch item := item.(type) {
	case *d.Issue:
		return renderIssue(item, thm, styles)
	case worklogListItem:
		return renderUnsyncedWorklog(item, styles, issueMap, fallbackCommentConfigured, selected)
	case syncedWorklogListItem:
		return renderSyncedWorklog(item, styles, issueMap, selected)
	default:
		return "", ""
	}
}

func renderIssue(issue *d.Issue, thm theme.Theme, styles styles) (string, string) {
	var trackingIndicator string
	if issue.TrackingActive {
		trackingIndicator = "⏲ "
	}
	title := trackingIndicator + utils.RightPadTrim(issue.Summary, int(float64(listWidth)*0.8))

	issueTypeColor := categoricalColor("issue-type", issue.IssueType, thm.CategoricalColors)
	issueType := styles.issueTypeBadge.
		Background(issueTypeColor).
		Render(issue.IssueType)

	assignee := utils.RightPadTrim(issue.Assignee, listWidth/4)
	if issue.Assignee != "" {
		assigneeColor := categoricalColor("assignee", issue.Assignee, thm.CategoricalColors)
		assignee = lipgloss.NewStyle().Foreground(assigneeColor).Render(assignee)
	}

	status := styles.issueStatus.Render(utils.RightPadTrim(issue.Status, listWidth/4))

	var totalTimeSpent string
	if issue.AggSecondsSpent > 0 {
		totalTimeSpent = styles.aggTimeSpent.Render(utils.HumanizeDuration(issue.AggSecondsSpent))
	}

	description := fmt.Sprintf(
		"%s%s%s%s%s",
		utils.RightPadTrim(issue.IssueKey, listWidth/4),
		status,
		assignee,
		issueType,
		totalTimeSpent,
	)

	return title, description
}

func renderUnsyncedWorklog(entry worklogListItem, styles styles, issueMap map[string]*d.Issue, fallbackCommentConfigured, selected bool) (string, string) {
	showComment := !entry.fallbackCommentUsed
	title := renderWorklogTitle(entry.StoredWorklog, styles.worklogCommentLabel, issueMap, showComment, selected)

	if entry.err != nil {
		return title, "error: " + entry.err.Error()
	}

	var duration string
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if startOfToday.Sub(entry.EndTS) > 0 {
		if entry.BeginTS.Format(dateFormat) == entry.EndTS.Format(dateFormat) {
			duration = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(dayAndTimeFormat), entry.EndTS.Format(timeOnlyFormat))
		} else {
			duration = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(dayAndTimeFormat), entry.EndTS.Format(dayAndTimeFormat))
		}
	} else {
		duration = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(timeOnlyFormat), entry.EndTS.Format(timeOnlyFormat))
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

	description := fmt.Sprintf(
		"%s%s%s%s%s",
		utils.RightPadTrim(entry.IssueKey, listWidth/4),
		utils.RightPadTrim(duration, listWidth/4),
		utils.RightPadTrim(fmt.Sprintf("(%s)", timeSpent), listWidth/4),
		syncStatus,
		fallbackCommentStatus,
	)

	return title, description
}

func renderSyncedWorklog(entry syncedWorklogListItem, styles styles, issueMap map[string]*d.Issue, selected bool) (string, string) {
	title := renderWorklogTitle(entry.StoredWorklog, styles.worklogCommentLabel, issueMap, true, selected)

	description := fmt.Sprintf(
		"%s%s%s",
		utils.RightPadTrim(entry.IssueKey, listWidth/4),
		utils.RightPadTrim(humanize.Time(entry.EndTS), listWidth/4),
		utils.RightPadTrim(fmt.Sprintf("(%s)", utils.HumanizeDuration(int(entry.EndTS.Sub(entry.BeginTS).Seconds()))), listWidth/4),
	)

	return title, description
}

func renderWorklogTitle(entry d.StoredWorklog, commentLabelStyle lipgloss.Style, issueMap map[string]*d.Issue, showComment, selected bool) string {
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
