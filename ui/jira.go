package ui

import (
	"context"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
)

func getIssues(cl *jira.Client, jql string) ([]jira.Issue, error) {
	issues, _, err := cl.Issue.Search(context.Background(), jql, nil)
	return issues, err
}

func addWLtoJira(cl *jira.Client, entry WorklogEntry, timeDeltaMins int) error {
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
	_, _, err := cl.Issue.AddWorklogRecord(context.Background(),
		entry.IssueKey,
		&wl,
	)
	return err

}
