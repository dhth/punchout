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

func syncWLToJIRA(cl *jira.Client, issueKey string, beginTS, endTS time.Time, comment string, timeDeltaMins int) error {
	start := beginTS

	if timeDeltaMins != 0 {
		start = start.Add(time.Minute * time.Duration(timeDeltaMins))
	}

	timeSpentSecs := int(endTS.Sub(beginTS).Seconds())
	wl := jira.WorklogRecord{
		IssueID:          issueKey,
		Started:          (*jira.Time)(&start),
		TimeSpentSeconds: timeSpentSecs,
		Comment:          comment,
	}
	cwl, _, err := cl.Issue.AddWorklogRecord(context.Background(),
		issueKey,
		&wl,
	)

	if cwl != nil && cwl.Started == nil {
		return errJIRARepliedWithEmptyWorklog
	}
	return err
}
