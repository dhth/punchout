package service

import (
	"context"

	d "github.com/dhth/punchout/internal/domain"
)

type Jira interface {
	GetIssues(jql string) ([]d.Issue, int, error)
	SyncWLToJIRA(ctx context.Context, entry d.WorklogEntry, comment string, timeDeltaMins int) error
	URL() string
}
