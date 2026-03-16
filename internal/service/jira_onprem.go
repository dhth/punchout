package service

import (
	"context"
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	d "github.com/dhth/punchout/internal/domain"
)

type onPremJira struct {
	client *jira.Client
}

func NewOnPremJiraSvc(url string, token string) (Jira, error) {
	tp := jira.BearerAuthTransport{
		Token: token,
	}

	client, err := jira.NewClient(url, tp.Client())
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntCreateJiraClient, err.Error())
	}

	return &onPremJira{client: client}, nil
}

func (svc *onPremJira) GetIssues(jql string) ([]d.Issue, int, error) {
	jIssues, resp, err := svc.client.Issue.Search(context.Background(), jql, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %s", errCouldntFetchIssuesFromJira, err.Error())
	}

	issues := make([]d.Issue, len(jIssues))
	for i, issue := range jIssues {
		issues[i] = mapOnPremIssue(issue)
	}

	return issues, resp.StatusCode, nil
}

func (svc *onPremJira) SyncWLToJIRA(ctx context.Context, entry d.WorklogEntry, comment string, timeDeltaMins int) error {
	start := getWorklogStart(entry, timeDeltaMins)

	wl := jira.WorklogRecord{
		IssueID:          entry.IssueKey,
		Started:          (*jira.Time)(&start),
		TimeSpentSeconds: getTimeSpentSeconds(entry),
		Comment:          comment,
	}

	cwl, _, err := svc.client.Issue.AddWorklogRecord(ctx, entry.IssueKey, &wl)
	if cwl != nil && cwl.Started == nil {
		return errJIRARepliedWithEmptyWorklog
	}

	return err
}

func (svc *onPremJira) URL() string {
	return svc.client.BaseURL.String()
}

func mapOnPremIssue(issue jira.Issue) d.Issue {
	if issue.Fields == nil {
		return d.Issue{IssueKey: issue.Key}
	}

	var assignee string
	if issue.Fields.Assignee != nil {
		assignee = issue.Fields.Assignee.DisplayName
	}

	var status string
	if issue.Fields.Status != nil {
		status = issue.Fields.Status.Name
	}

	return d.Issue{
		IssueKey:        issue.Key,
		IssueType:       issue.Fields.Type.Name,
		Summary:         issue.Fields.Summary,
		Assignee:        assignee,
		Status:          status,
		AggSecondsSpent: issue.Fields.AggregateTimeSpent,
	}
}
