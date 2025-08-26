package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	jiraCloud "github.com/andygrunwald/go-jira/v2/cloud"
	jira "github.com/andygrunwald/go-jira/v2/onpremise"
)

var (
	errJIRARepliedWithEmptyWorklog = errors.New("JIRA replied with an empty worklog; something is probably wrong")
	errCouldntCreateJiraClient     = errors.New("couldn't create JIRA client")
)

type Jira struct {
	client *jira.Client
}

func NewOnPremJiraSvc(url string, token string) (Jira, error) {
	var zero Jira

	tp := jira.BearerAuthTransport{
		Token: token,
	}
	httpClient := tp.Client()

	client, err := jira.NewClient(url, httpClient)
	if err != nil {
		return zero, fmt.Errorf("%w: %s", errCouldntCreateJiraClient, err.Error())
	}

	return Jira{
		client: client,
	}, nil
}

func NewCloudJiraSvc(url string, userName string, token string) (Jira, error) {
	var zero Jira

	tp := jiraCloud.BasicAuthTransport{
		Username: userName,
		APIToken: token,
	}
	httpClient := tp.Client()

	// Using the on-premise client regardless of the user's installation type
	// The APIs between the two installation types seem to differ, but this
	// seems to be alright for punchout's use case. If this situation changes,
	// this will need to be refactored.
	// https://github.com/andygrunwald/go-jira/issues/473
	client, err := jira.NewClient(url, httpClient)
	if err != nil {
		return zero, fmt.Errorf("%w: %s", errCouldntCreateJiraClient, err.Error())
	}

	return Jira{
		client: client,
	}, nil
}

func (svc Jira) GetIssues(jql string) ([]jira.Issue, int, error) {
	issues, resp, err := svc.client.Issue.Search(context.Background(), jql, nil)
	return issues, resp.StatusCode, err
}

func (svc Jira) SyncWLToJIRA(issueKey string, beginTS, endTS time.Time, comment string, timeDeltaMins int) error {
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
	cwl, _, err := svc.client.Issue.AddWorklogRecord(context.Background(),
		issueKey,
		&wl,
	)

	if cwl != nil && cwl.Started == nil {
		return errJIRARepliedWithEmptyWorklog
	}
	return err
}

func (svc Jira) JiraURL() string {
	return svc.client.BaseURL.String()
}
