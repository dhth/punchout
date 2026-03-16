package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jiraCloud "github.com/andygrunwald/go-jira/v2/cloud"
	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	d "github.com/dhth/punchout/internal/domain"
)

var (
	errJIRARepliedWithEmptyWorklog = errors.New("JIRA replied with an empty worklog; something is probably wrong")
	errCouldntCreateJiraClient     = errors.New("couldn't create JIRA client")
	errCouldntFetchIssuesFromJira  = errors.New("couldn't fetch issues from JIRA")
)

type Jira struct {
	client  *jira.Client
	isCloud bool
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
		client:  client,
		isCloud: false,
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
		client:  client,
		isCloud: true,
	}, nil
}

// searchResult mirrors the Jira search response structure.
type searchResult struct {
	Issues []jira.Issue `json:"issues"`
}

func (svc Jira) GetIssues(jql string) ([]d.Issue, int, error) {
	var zero []d.Issue

	var jIssues []jira.Issue
	var resp *jira.Response

	if svc.isCloud {
		// Atlassian removed GET /rest/api/2/search on Jira Cloud.
		// Use the replacement /rest/api/2/search/jql endpoint instead.
		// See: https://developer.atlassian.com/changelog/#CHANGE-2046
		u := url.URL{Path: "rest/api/2/search/jql"}
		uv := url.Values{}
		if jql != "" {
			uv.Add("jql", jql)
		}
		// The new endpoint defaults to returning only issue IDs;
		// request all fields to match the old endpoint's behavior.
		uv.Add("fields", "*all")
		u.RawQuery = uv.Encode()

		req, err := svc.client.NewRequest(context.Background(), http.MethodGet, u.String(), nil)
		if err != nil {
			return zero, 0, fmt.Errorf("%w: %s", errCouldntFetchIssuesFromJira, err.Error())
		}
		v := new(searchResult)
		resp, err = svc.client.Do(req, v)
		if err != nil {
			return zero, 0, fmt.Errorf("%w: %s", errCouldntFetchIssuesFromJira, err.Error())
		}
		jIssues = v.Issues
	} else {
		var err error
		jIssues, resp, err = svc.client.Issue.Search(context.Background(), jql, nil)
		if err != nil {
			return zero, 0, fmt.Errorf("%w: %s", errCouldntFetchIssuesFromJira, err.Error())
		}
	}

	issues := make([]d.Issue, len(jIssues))
	for i, issue := range jIssues {
		var assignee string
		var totalSecsSpent int
		var status string
		if issue.Fields != nil {
			if issue.Fields.Assignee != nil {
				assignee = issue.Fields.Assignee.DisplayName
			}

			totalSecsSpent = issue.Fields.AggregateTimeSpent

			if issue.Fields.Status != nil {
				status = issue.Fields.Status.Name
			}
		}
		issues[i] = d.Issue{
			IssueKey:        issue.Key,
			IssueType:       issue.Fields.Type.Name,
			Summary:         issue.Fields.Summary,
			Assignee:        assignee,
			Status:          status,
			AggSecondsSpent: totalSecsSpent,
		}
	}

	return issues, resp.StatusCode, nil
}

func (svc Jira) SyncWLToJIRA(ctx context.Context, entry d.WorklogEntry, comment string, timeDeltaMins int) error {
	start := entry.BeginTS

	if timeDeltaMins != 0 {
		start = start.Add(time.Minute * time.Duration(timeDeltaMins))
	}

	timeSpentSecs := int(entry.EndTS.Sub(entry.BeginTS).Seconds())

	wl := jira.WorklogRecord{
		IssueID:          entry.IssueKey,
		Started:          (*jira.Time)(&start),
		TimeSpentSeconds: timeSpentSecs,
		Comment:          comment,
	}
	cwl, _, err := svc.client.Issue.AddWorklogRecord(ctx,
		entry.IssueKey,
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
