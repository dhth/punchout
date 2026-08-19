package service

import (
	"context"

	d "github.com/dhth/punchout/internal/domain"
)

type Jira interface {
	GetIssues(jql string) ([]d.Issue, error)
	SyncWLToJIRA(ctx context.Context, worklog d.Worklog, timeDeltaMins int) error
	URL() string
}
