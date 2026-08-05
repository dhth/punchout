package domain

import (
	"strings"
	"time"
)

type Issue struct {
	IssueKey        string `json:"issue_key" jsonschema:"issue key"`
	IssueType       string `json:"issue_type" jsonschema:"issue type"`
	Summary         string `json:"summary" jsonschema:"issue summary"`
	Assignee        string `json:"assignee" jsonschema:"issue assignee"`
	Status          string `json:"status" jsonschema:"issue status"`
	AggSecondsSpent int    `json:"agg_seconds_spent" jsonschema:"aggregate seconds spent on the issue"`
	TrackingActive  bool   `json:"-"`
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

func (entry WorklogEntry) FilterValue() string { return entry.IssueKey }

func (entry SyncedWorklogEntry) FilterValue() string { return entry.IssueKey }

type ValidatedWorkLog struct {
	IssueKey string
	BeginTS  time.Time
	EndTS    time.Time
	Comment  string
}
