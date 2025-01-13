package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type Issue struct {
	issueKey        string
	issueType       string
	summary         string
	assignee        string
	status          string
	aggSecondsSpent int
	trackingActive  bool
	desc            string
}

func (issue Issue) Title() string {
	var trackingIndicator string
	if issue.trackingActive {
		trackingIndicator = "⏲ "
	}
	return trackingIndicator + RightPadTrim(issue.summary, int(float64(listWidth)*0.8))
}

func (issue Issue) Description() string {
	return issue.desc
}

func (issue Issue) FilterValue() string { return issue.issueKey }

type worklogEntry struct {
	ID             int
	IssueKey       string
	BeginTS        time.Time
	EndTS          time.Time
	Comment        string
	Active         bool
	Synced         bool
	SyncInProgress bool
	Error          error
}

type syncedWorklogEntry struct {
	ID       int
	IssueKey string
	BeginTS  time.Time
	EndTS    time.Time
	Comment  string
}

func (entry *worklogEntry) needsComment() bool {
	return strings.TrimSpace(entry.Comment) == ""
}

func (entry worklogEntry) SecsSpent() int {
	return int(entry.EndTS.Sub(entry.BeginTS).Seconds())
}

func (entry worklogEntry) Title() string {
	if entry.needsComment() {
		return "[NO COMMENT]"
	}

	return entry.Comment
}

func (entry worklogEntry) Description() string {
	if entry.Error != nil {
		return "error: " + entry.Error.Error()
	}

	var syncedStatus string
	var durationMsg string

	now := time.Now()

	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if startOfToday.Sub(entry.EndTS) > 0 {
		if entry.BeginTS.Format(dateFormat) == entry.EndTS.Format(dateFormat) {
			durationMsg = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(dayAndTimeFormat), entry.EndTS.Format(timeOnlyFormat))
		} else {
			durationMsg = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(dayAndTimeFormat), entry.EndTS.Format(dayAndTimeFormat))
		}
	} else {
		durationMsg = fmt.Sprintf("%s  ...  %s", entry.BeginTS.Format(timeOnlyFormat), entry.EndTS.Format(timeOnlyFormat))
	}

	timeSpentStr := humanizeDuration(int(entry.EndTS.Sub(entry.BeginTS).Seconds()))

	if entry.needsComment() {
		syncedStatus = needsCommentStyle.Render("needs comment")
	} else if entry.Synced {
		syncedStatus = syncedStyle.Render("synced")
	} else if entry.SyncInProgress {
		syncedStatus = syncingStyle.Render("syncing")
	} else {
		syncedStatus = notSyncedStyle.Render("not synced")
	}

	return fmt.Sprintf("%s%s%s%s",
		RightPadTrim(entry.IssueKey, listWidth/4),
		RightPadTrim(durationMsg, listWidth/4),
		RightPadTrim(fmt.Sprintf("(%s)", timeSpentStr), listWidth/4),
		syncedStatus,
	)
}
func (entry worklogEntry) FilterValue() string { return entry.IssueKey }

func (entry syncedWorklogEntry) Title() string {
	return entry.Comment
}

func (entry syncedWorklogEntry) Description() string {
	durationMsg := humanize.Time(entry.EndTS)
	timeSpentStr := humanizeDuration(int(entry.EndTS.Sub(entry.BeginTS).Seconds()))
	return fmt.Sprintf("%s%s%s",
		RightPadTrim(entry.IssueKey, listWidth/4),
		RightPadTrim(durationMsg, listWidth/4),
		fmt.Sprintf("(%s)", timeSpentStr),
	)
}
func (entry syncedWorklogEntry) FilterValue() string { return entry.IssueKey }
