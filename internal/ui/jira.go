package ui

import (
	"context"
	"errors"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
)

var errJIRARepliedWithEmptyWorklog = errors.New("JIRA replied with an empty worklog; something is probably wrong")

func getIssues(cl *jira.Client, jql string) ([]jira.Issue, int, error) {
	issues, resp, err := cl.Issue.Search(context.Background(), jql, nil)
	return issues, resp.StatusCode, err
}

func addWLtoJira(cl *jira.Client, entry worklogEntry, timeDeltaMins int) error {
	start := entry.BeginTS

	if timeDeltaMins != 0 {
		start = start.Add(time.Minute * time.Duration(timeDeltaMins))
	}

	timeSpentSecs := int(entry.EndTS.Sub(entry.BeginTS).Seconds())
	wl := jira.WorklogRecord{
		IssueID:          entry.IssueKey,
		Started:          (*jira.Time)(&start),
		TimeSpentSeconds: timeSpentSecs,
		Comment:          entry.Comment,
	}
	cwl, _, err := cl.Issue.AddWorklogRecord(context.Background(),
		entry.IssueKey,
		&wl,
	)

	if cwl != nil && cwl.Started == nil {
		return errJIRARepliedWithEmptyWorklog
	}
	return err
}
