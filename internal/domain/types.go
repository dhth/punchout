package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/dhth/punchout/internal/utils"
	"github.com/dustin/go-humanize"
)

// TODO: get all UI logic out of the domain package
var listWidth = 140

const (
	dayAndTimeFormat = "Mon, 15:04"
	dateFormat       = "2006/01/02"
	timeOnlyFormat   = "15:04"
)

type Issue struct {
	IssueKey        string `json:"issue_key" jsonschema:"issue key"`
	IssueType       string `json:"issue_type" jsonschema:"issue type"`
	Summary         string `json:"summary" jsonschema:"issue summary"`
	Assignee        string `json:"assignee" jsonschema:"issue assignee"`
	Status          string `json:"status" jsonschema:"issue status"`
	AggSecondsSpent int    `json:"agg_seconds_spent" jsonschema:"aggregate seconds spent on the issue"`
	TrackingActive  bool   `json:"-"`
	Desc            string `json:"-"`
}

func (issue *Issue) SetDesc() {
	// TODO: The padding here is a bit of a mess; make it more readable
	var assignee string
	var status string
	var totalSecsSpent string

	issueType := getIssueTypeStyle(issue.IssueType).Render(issue.IssueType)

	if issue.Assignee != "" {
		assignee = assigneeStyle(issue.Assignee).Render(utils.RightPadTrim(issue.Assignee, listWidth/4))
	} else {
		assignee = assigneeStyle(issue.Assignee).Render(utils.RightPadTrim("", listWidth/4))
	}

	status = issueStatusStyle.Render(utils.RightPadTrim(issue.Status, listWidth/4))

	if issue.AggSecondsSpent > 0 {
		totalSecsSpent = aggTimeSpentStyle.Render(utils.HumanizeDuration(issue.AggSecondsSpent))
	}

	issue.Desc = fmt.Sprintf("%s%s%s%s%s", utils.RightPadTrim(issue.IssueKey, listWidth/4), status, assignee, issueType, totalSecsSpent)
}

func (issue Issue) Title() string {
	var trackingIndicator string
	if issue.TrackingActive {
		trackingIndicator = "⏲ "
	}
	return trackingIndicator + utils.RightPadTrim(issue.Summary, int(float64(listWidth)*0.8))
}

func (issue Issue) Description() string {
	return issue.Desc
}

func (issue Issue) FilterValue() string { return issue.IssueKey }

type WorklogEntry struct {
	ID              int        `json:"id" jsonschema:"worklog entry ID"`
	IssueKey        string     `json:"issue_key" jsonschema:"JIRA issue key"`
	BeginTS         time.Time  `json:"begin_time" jsonschema:"worklog begin time"`
	EndTS           *time.Time `json:"end_time" jsonschema:"worklog end time"`
	Comment         *string    `json:"comment" jsonschema:"worklog comment"`
	FallbackComment *string    `json:"-"`
	Active          bool       `json:"-"`
	Synced          bool       `json:"-"`
	SyncInProgress  bool       `json:"-"`
	Error           error      `json:"-"`
}

type SyncedWorklogEntry struct {
	ID       int
	IssueKey string
	BeginTS  time.Time
	EndTS    time.Time
	Comment  *string
}

func (entry *WorklogEntry) NeedsComment() bool {
	if entry.Comment == nil {
		return true
	}

	return strings.TrimSpace(*entry.Comment) == ""
}

func (entry *SyncedWorklogEntry) NeedsComment() bool {
	if entry.Comment == nil {
		return true
	}

	return strings.TrimSpace(*entry.Comment) == ""
}

func (entry WorklogEntry) SecsSpent() int {
	return int(entry.EndTS.Sub(entry.BeginTS).Seconds())
}

func (entry WorklogEntry) Title() string {
	if entry.NeedsComment() {
		return "[NO COMMENT]"
	}

	return *entry.Comment
}

func (entry WorklogEntry) Description() string {
	if entry.Error != nil {
		return "error: " + entry.Error.Error()
	}

	var syncedStatus string
	var fallbackCommentStatus string
	var durationMsg string

	now := time.Now()

	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if entry.EndTS != nil && startOfToday.Sub(*entry.EndTS) > 0 {
		if entry.BeginTS.Format(dateFormat) == entry.EndTS.Format(dateFormat) {
			durationMsg = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(dayAndTimeFormat), entry.EndTS.Format(timeOnlyFormat))
		} else {
			durationMsg = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(dayAndTimeFormat), entry.EndTS.Format(dayAndTimeFormat))
		}
	} else {
		durationMsg = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(timeOnlyFormat), entry.EndTS.Format(timeOnlyFormat))
	}

	timeSpentStr := utils.HumanizeDuration(int(entry.EndTS.Sub(entry.BeginTS).Seconds()))

	if entry.Synced {
		syncedStatus = syncedStyle.Render("synced")
	} else if entry.SyncInProgress {
		syncedStatus = syncingStyle.Render("syncing")
	} else {
		syncedStatus = notSyncedStyle.Render("not synced")
	}

	if entry.NeedsComment() && entry.FallbackComment != nil {
		fallbackCommentStatus = usingFallbackCommentStyle.Render("fallback comment")
	}

	return fmt.Sprintf("%s%s%s%s%s",
		utils.RightPadTrim(entry.IssueKey, listWidth/4),
		utils.RightPadTrim(durationMsg, listWidth/4),
		utils.RightPadTrim(fmt.Sprintf("(%s)", timeSpentStr), listWidth/6),
		syncedStatus,
		fallbackCommentStatus,
	)
}
func (entry WorklogEntry) FilterValue() string { return entry.IssueKey }

func (entry SyncedWorklogEntry) Title() string {
	if entry.NeedsComment() {
		return "[NO COMMENT]"
	}

	return *entry.Comment
}

func (entry SyncedWorklogEntry) Description() string {
	durationMsg := humanize.Time(entry.EndTS)
	timeSpentStr := utils.HumanizeDuration(int(entry.EndTS.Sub(entry.BeginTS).Seconds()))
	return fmt.Sprintf("%s%s%s",
		utils.RightPadTrim(entry.IssueKey, listWidth/4),
		utils.RightPadTrim(durationMsg, listWidth/4),
		fmt.Sprintf("(%s)", timeSpentStr),
	)
}
func (entry SyncedWorklogEntry) FilterValue() string { return entry.IssueKey }

type ValidatedWorkLog struct {
	IssueKey string
	BeginTS  time.Time
	EndTS    time.Time
	Comment  string
}
