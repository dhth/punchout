package service

import (
	"context"
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	d "github.com/dhth/punchout/internal/domain"
)

type cloudJira struct {
	client *jira.Client
}

func NewCloudJiraSvc(url string, userName string, token string) (Jira, error) {
	tp := jira.BasicAuthTransport{
		Username: userName,
		APIToken: token,
	}

	client, err := jira.NewClient(url, tp.Client())
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntCreateJiraClient, err.Error())
	}

	return &cloudJira{client: client}, nil
}

func (svc *cloudJira) GetIssues(jql string) ([]d.Issue, int, error) {
	jIssues, resp, err := svc.client.Issue.SearchV2JQL(context.Background(), jql, &jira.SearchOptionsV2{
		Fields: []string{
			"issuetype",
			"summary",
			"assignee",
			"status",
			"aggregatetimespent",
		},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %s", errCouldntFetchIssuesFromJira, err.Error())
	}

	issues := make([]d.Issue, len(jIssues))
	for i, issue := range jIssues {
		issues[i] = mapCloudIssue(issue)
	}

	return issues, resp.StatusCode, nil
}

func (svc *cloudJira) SyncWLToJIRA(ctx context.Context, entry d.WorklogEntry, comment string, timeDeltaMins int) error {
	if entry.EndTS == nil {
		return errCannotSyncWLWithoutEndTime
	}

	start := getWorklogStart(entry, timeDeltaMins)

	wl := jira.WorklogRecord{
		IssueID:          entry.IssueKey,
		Started:          (*jira.Time)(&start),
		TimeSpentSeconds: getTimeSpentSeconds(entry.BeginTS, *entry.EndTS),
		Comment:          comment,
	}

	cwl, _, err := svc.client.Issue.AddWorklogRecord(ctx, entry.IssueKey, &wl)
	if cwl != nil && cwl.Started == nil {
		return errJIRARepliedWithEmptyWorklog
	}

	return err
}

func (svc *cloudJira) URL() string {
	return svc.client.BaseURL.String()
}

func mapCloudIssue(issue jira.Issue) d.Issue {
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
