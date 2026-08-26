package ui

import (
	"context"
	"time"

	"github.com/dhth/punchout/internal/domain"
)

type WorklogStore interface {
	ActiveWorklog(context.Context) (*domain.InProgressWorklog, error)

	StartWorklog(context.Context, string, time.Time) error
	FinishWorklog(context.Context, domain.Worklog) error
	SwitchActiveWorklog(context.Context, string, time.Time) (previousIssue string, err error)
	UpdateActiveWorklog(context.Context, time.Time, *string) error
	DeleteActiveWorklog(context.Context) error

	AddWorklog(context.Context, domain.Worklog) error
	UpdateWorklog(context.Context, int, domain.Worklog) error
	DeleteWorklog(context.Context, int) error

	UnsyncedWorklogs(context.Context) ([]domain.StoredWorklog, error)
	SyncedWorklogs(context.Context) ([]domain.StoredWorklog, error)
	MarkWorklogSynced(context.Context, int, *string) error
}
