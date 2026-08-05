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

func renderListItem(item list.Item, thm theme.Theme, styles styles) (string, string) {
	switch item := item.(type) {
	case *d.Issue:
		return renderIssue(item, thm, styles)
	case d.WorklogEntry:
		return renderWorklogEntry(item, styles)
	case d.SyncedWorklogEntry:
		return renderSyncedWorklogEntry(item)
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

func renderWorklogEntry(entry d.WorklogEntry, styles styles) (string, string) {
	title := "[NO COMMENT]"
	if !entry.NeedsComment() {
		title = *entry.Comment
	}

	if entry.Error != nil {
		return title, "error: " + entry.Error.Error()
	}

	var duration string
	if entry.EndTS != nil {
		now := time.Now()
		startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		if startOfToday.Sub(*entry.EndTS) > 0 {
			if entry.BeginTS.Format(dateFormat) == entry.EndTS.Format(dateFormat) {
				duration = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(dayAndTimeFormat), entry.EndTS.Format(timeOnlyFormat))
			} else {
				duration = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(dayAndTimeFormat), entry.EndTS.Format(dayAndTimeFormat))
			}
		} else {
			duration = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(timeOnlyFormat), entry.EndTS.Format(timeOnlyFormat))
		}
	}

	var timeSpent string
	if entry.EndTS != nil {
		timeSpent = utils.HumanizeDuration(int(entry.EndTS.Sub(entry.BeginTS).Seconds()))
	}

	var syncStatus string
	switch {
	case entry.Synced:
		syncStatus = styles.syncedBadge.Render("synced")
	case entry.SyncInProgress:
		syncStatus = styles.syncingBadge.Render("syncing")
	default:
		syncStatus = styles.notSyncedBadge.Render("not synced")
	}

	var fallbackCommentStatus string
	if entry.NeedsComment() && entry.FallbackComment != nil {
		fallbackCommentStatus = styles.fallbackCommentBadge.Render("fallback comment")
	}

	description := fmt.Sprintf(
		"%s%s%s%s%s",
		utils.RightPadTrim(entry.IssueKey, listWidth/4),
		utils.RightPadTrim(duration, listWidth/4),
		utils.RightPadTrim(fmt.Sprintf("(%s)", timeSpent), listWidth/6),
		syncStatus,
		fallbackCommentStatus,
	)

	return title, description
}

func renderSyncedWorklogEntry(entry d.SyncedWorklogEntry) (string, string) {
	title := "[NO COMMENT]"
	if !entry.NeedsComment() {
		title = *entry.Comment
	}

	description := fmt.Sprintf(
		"%s%s%s",
		utils.RightPadTrim(entry.IssueKey, listWidth/4),
		utils.RightPadTrim(humanize.Time(entry.EndTS), listWidth/4),
		fmt.Sprintf("(%s)", utils.HumanizeDuration(int(entry.EndTS.Sub(entry.BeginTS).Seconds()))),
	)

	return title, description
}

func categoricalColor(category, value string, colors []string) color.Color {
	if len(colors) == 0 {
		return lipgloss.NoColor{}
	}

	h := fnv.New32()
	_, _ = h.Write([]byte(strings.Join([]string{category, value}, ":")))

	return lipgloss.Color(colors[h.Sum32()%uint32(len(colors))])
}
