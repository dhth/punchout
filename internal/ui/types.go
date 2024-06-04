package ui

import (
	"fmt"
	"time"
)

type Issue struct {
	issueKey        string
	issueType       string
	summary         string
	assignee        string
	status          string
	aggSecondsSpent int
	trackingActive  bool
}

func (issue Issue) Title() string {
	var trackingIndicator string
	if issue.trackingActive {
		trackingIndicator = "⏲ "
	}
	return trackingIndicator + RightPadTrim(issue.summary, int(float64(listWidth)*0.8))
}
func (issue Issue) Description() string {
	// TODO: The padding here is a bit of a mess; make it more readable
	var assignee string
	var status string
	var totalSecsSpent string

	issueType := getIssueTypeStyle(issue.issueType).Render(Trim(issue.issueType, int(float64(listWidth)*0.2)))

	if issue.assignee != "" {
		assignee = assigneeStyle(issue.assignee).Render(RightPadTrim("@"+issue.assignee, int(float64(listWidth)*0.2)))
	} else {
		assignee = assigneeStyle(issue.assignee).Render(RightPadTrim("", int(float64(listWidth)*0.2)))
	}

	status = issueStatusStyle.Render(RightPadTrim(issue.status, int(float64(listWidth)*0.2)))

	if issue.aggSecondsSpent > 0 {
		if issue.aggSecondsSpent < 3600 {
			totalSecsSpent = aggTimeSpentStyle.Render(fmt.Sprintf("%2dm", int(issue.aggSecondsSpent/60)))
		} else {
			totalSecsSpent = aggTimeSpentStyle.Render(fmt.Sprintf("%2dh", int(issue.aggSecondsSpent/3600)))
		}
	}

	return fmt.Sprintf("%s%s%s%s%s", RightPadTrim(issue.issueKey, int(float64(listWidth)*0.3)), status, assignee, issueType, totalSecsSpent)
}
func (issue Issue) FilterValue() string { return issue.issueKey }

type WorklogEntry struct {
	Id             int
	IssueKey       string
	BeginTS        time.Time
	EndTS          time.Time
	Comment        string
	Active         bool
	Synced         bool
	SyncInProgress bool
	Error          error
}

type SyncedWorklogEntry struct {
	Id       int
	IssueKey string
	BeginTS  time.Time
	EndTS    time.Time
	Comment  string
}

func (entry WorklogEntry) Title() string {
	return entry.Comment
}
func (entry WorklogEntry) Description() string {
	if entry.Error != nil {
		return "error: " + entry.Error.Error()
	}

	var syncedStatus string
	if entry.Synced {
		syncedStatus = syncedStyle.Render("synced")
	} else if entry.SyncInProgress {
		syncedStatus = syncingStyle.Render("syncing")
	} else {
		syncedStatus = notSyncedStyle.Render("not synced")
	}
	minsSpent := int(entry.EndTS.Sub(entry.BeginTS).Minutes())
	minsSpentStr := fmt.Sprintf("spent %d mins", minsSpent)
	return fmt.Sprintf("%s%s%s%s",
		RightPadTrim(entry.IssueKey, int(listWidth/4)),
		RightPadTrim("started: "+entry.BeginTS.Format("Mon, 3:04pm"), int(listWidth/4)),
		RightPadTrim(minsSpentStr, int(listWidth/4)),
		syncedStatus,
	)
}
func (entry WorklogEntry) FilterValue() string { return entry.IssueKey }

func (entry SyncedWorklogEntry) Title() string {
	return entry.Comment
}
func (entry SyncedWorklogEntry) Description() string {
	minsSpent := int(entry.EndTS.Sub(entry.BeginTS).Minutes())
	minsSpentStr := fmt.Sprintf("spent %d mins", minsSpent)
	return fmt.Sprintf("%s%s%s",
		RightPadTrim(entry.IssueKey, int(listWidth/4)),
		RightPadTrim("started: "+entry.BeginTS.Format("Mon, 3:04pm"), int(listWidth/4)),
		RightPadTrim(minsSpentStr, int(listWidth/4)),
	)
}
func (entry SyncedWorklogEntry) FilterValue() string { return entry.IssueKey }
