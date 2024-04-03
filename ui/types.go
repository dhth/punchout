package ui

import (
	"fmt"
	"time"
)

type Issue struct {
	IssueKey  string
	IssueType string
	Summary   string
}

func (issue Issue) Title() string {
	return RightPadTrim(issue.Summary, int(float64(listWidth)*0.8))
}
func (issue Issue) Description() string {
	issueType := getIssueTypeStyle(issue.IssueType).Render(Trim(issue.IssueType, int(float64(listWidth)*0.2)))
	return fmt.Sprintf("%s%s", RightPadTrim(issue.IssueKey, int(float64(listWidth)*0.78)), issueType)
}
func (issue Issue) FilterValue() string { return issue.IssueKey + " : " + issue.Summary }

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
