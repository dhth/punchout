package tools

import (
	"context"
	"testing"
	"time"

	"github.com/dhth/punchout/internal/config"
	"github.com/dhth/punchout/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncWorklogsToJiraPersistsDespiteRequestCancellation(t *testing.T) {
	// GIVEN
	ctx, cancelRequest := context.WithCancel(t.Context())
	beginTS := time.Date(2026, time.August, 24, 9, 0, 0, 0, time.UTC)
	entry := domain.StoredWorklog{
		ID: 42,
		Worklog: domain.Worklog{
			IssueKey: "TEST-42",
			BeginTS:  beginTS,
			EndTS:    beginTS.Add(time.Hour),
			Comment:  "completed work",
		},
	}

	type markWorklogSyncedCall struct {
		ctxErr      error
		hasDeadline bool
		id          int
		comment     *string
	}
	markWorklogSyncedCalls := make(chan markWorklogSyncedCall, 1)

	store := &stubWorklogStore{
		unsyncedWorklogsFunc: func(context.Context) ([]domain.StoredWorklog, error) {
			return []domain.StoredWorklog{entry}, nil
		},
		markWorklogSyncedFunc: func(updateCtx context.Context, id int, comment *string) error {
			_, ok := updateCtx.Deadline()
			markWorklogSyncedCalls <- markWorklogSyncedCall{
				ctxErr:      updateCtx.Err(),
				hasDeadline: ok,
				id:          id,
				comment:     comment,
			}
			return nil
		},
	}
	jira := &stubJira{
		syncWLToJIRAFunc: func(context.Context, domain.Worklog, int) error {
			cancelRequest()
			return nil
		},
	}
	handler := NewHandler(store, jira, config.JiraOptions{})

	// WHEN
	_, output, err := handler.syncWorklogsToJira(ctx, nil, struct{}{})

	// THEN
	require.NoError(t, err)
	assert.Equal(t, []syncSuccess{{EntryID: entry.ID, IssueKey: entry.IssueKey}}, output.Successes)
	assert.Empty(t, output.Errors)
	call := <-markWorklogSyncedCalls
	require.NoError(t, call.ctxErr)
	require.True(t, call.hasDeadline)
	assert.Equal(t, entry.ID, call.id)
	assert.Nil(t, call.comment)
}

type stubWorklogStore struct {
	unsyncedWorklogsFunc  func(context.Context) ([]domain.StoredWorklog, error)
	markWorklogSyncedFunc func(context.Context, int, *string) error
}

func (*stubWorklogStore) AddWorklog(context.Context, domain.Worklog) error {
	panic("unexpected call to AddWorklog")
}

func (s *stubWorklogStore) UnsyncedWorklogs(ctx context.Context) ([]domain.StoredWorklog, error) {
	if s.unsyncedWorklogsFunc == nil {
		panic("unexpected call to UnsyncedWorklogs")
	}
	return s.unsyncedWorklogsFunc(ctx)
}

func (s *stubWorklogStore) MarkWorklogSynced(ctx context.Context, id int, comment *string) error {
	if s.markWorklogSyncedFunc == nil {
		panic("unexpected call to MarkWorklogSynced")
	}
	return s.markWorklogSyncedFunc(ctx, id, comment)
}

type stubJira struct {
	syncWLToJIRAFunc func(context.Context, domain.Worklog, int) error
}

func (*stubJira) GetIssues(string) ([]domain.Issue, error) {
	panic("unexpected call to GetIssues")
}

func (s *stubJira) SyncWLToJIRA(ctx context.Context, worklog domain.Worklog, delta int) error {
	return s.syncWLToJIRAFunc(ctx, worklog, delta)
}

func (*stubJira) URL() string {
	panic("unexpected call to URL")
}
