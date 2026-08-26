package tools

import (
	"context"

	"github.com/dhth/punchout/internal/domain"
)

type WorklogStore interface {
	AddWorklog(context.Context, domain.Worklog) error
	UnsyncedWorklogs(context.Context) ([]domain.StoredWorklog, error)
	MarkWorklogSynced(context.Context, int, *string) error
}
