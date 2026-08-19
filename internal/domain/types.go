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

type InProgressWorklog struct {
	IssueKey string
	BeginTS  time.Time
	Comment  string
}

type Worklog struct {
	IssueKey string    `json:"issue_key" jsonschema:"JIRA issue key"`
	BeginTS  time.Time `json:"begin_time" jsonschema:"worklog begin time"`
	EndTS    time.Time `json:"end_time" jsonschema:"worklog end time"`
	Comment  string    `json:"comment" jsonschema:"worklog comment"`
}

func (w Worklog) NeedsComment() bool {
	return strings.TrimSpace(w.Comment) == ""
}

func (w Worklog) SecsSpent() int {
	return int(w.EndTS.Sub(w.BeginTS).Seconds())
}

type StoredWorklog struct {
	Worklog
	ID     int  `json:"id" jsonschema:"worklog entry ID"`
	Synced bool `json:"-"`
}
