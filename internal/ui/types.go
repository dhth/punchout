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
		RightPadTrim("ended: "+entry.BeginTS.Format("Mon, 3:04pm"), int(listWidth/4)),
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
		RightPadTrim("ended: "+entry.BeginTS.Format("Mon, 3:04pm"), int(listWidth/4)),
		RightPadTrim(minsSpentStr, int(listWidth/4)),
	)
}
func (entry SyncedWorklogEntry) FilterValue() string { return entry.IssueKey }
