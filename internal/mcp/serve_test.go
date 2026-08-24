package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dhth/punchout/internal/config"
	"github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/mcp/tools"
	svc "github.com/dhth/punchout/internal/service"
	"github.com/gkampitakis/go-snaps/snaps"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPServer(t *testing.T) {
	server, err := newServer(nil, nil, config.JiraOptions{})
	require.NoError(t, err)

	httpServer := httptest.NewTestServer(t, newHTTPHandler(server))
	httpClient := httpServer.Client()
	client := sdk.NewClient(&sdk.Implementation{Name: "punchout-test"}, nil)

	connect := func(t *testing.T) (context.Context, *sdk.ClientSession) {
		t.Helper()

		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		t.Cleanup(cancel)

		session, err := client.Connect(ctx, &sdk.StreamableClientTransport{
			Endpoint:   httpServer.URL + "/v1",
			HTTPClient: httpClient,
		}, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = session.Close() })

		return ctx, session
	}

	t.Run("health endpoint works", func(t *testing.T) {
		// GIVEN
		endpoint := httpServer.URL + "/health"

		// WHEN
		response, err := httpClient.Get(endpoint)
		require.NoError(t, err)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, 200, response.StatusCode)
		assert.Equal(t, "text/plain", response.Header.Get("Content-Type"))
		assert.Equal(t, "HEALTHY", string(body))
	})

	t.Run("negotiates MCP protocol 2026-07-28", func(t *testing.T) {
		// GIVEN
		_, session := connect(t)

		// WHEN
		protocolVersion := session.InitializeResult().ProtocolVersion

		// THEN
		assert.Equal(t, "2026-07-28", protocolVersion)
	})

	t.Run("lists tools", func(t *testing.T) {
		// GIVEN
		ctx, session := connect(t)

		// WHEN
		result, err := session.ListTools(ctx, &sdk.ListToolsParams{})
		require.NoError(t, err)

		toolNames := make([]string, 0, len(result.Tools))
		for _, tool := range result.Tools {
			toolNames = append(toolNames, tool.Name)
		}

		// THEN
		assert.ElementsMatch(t, []string{
			"add_multiple_worklogs",
			"add_worklog",
			"get_jira_issues",
			"get_unsynced_worklogs",
			"sync_worklogs_to_jira",
		}, toolNames)
	})
}

func TestHTTPServerAddWorklog(t *testing.T) {
	t.Run("adds a valid worklog", func(t *testing.T) {
		// GIVEN
		var addedWorklogs []domain.Worklog
		store := &stubWorklogStore{
			addWorklogFunc: func(_ context.Context, worklog domain.Worklog) error {
				addedWorklogs = append(addedWorklogs, worklog)
				return nil
			},
		}
		ctx, session := setupMCPSession(t, store, nil, config.JiraOptions{})

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "add_worklog",
			Arguments: map[string]any{
				"issue_key":  "TEST-123",
				"begin_time": "2026/08/24 09:30",
				"end_time":   "2026/08/24 10:45",
				"comment":    "implemented persistence boundary",
			},
		})

		// THEN
		require.NoError(t, err)
		require.False(t, result.IsError)
		require.Equal(t, []domain.Worklog{{
			IssueKey: "TEST-123",
			BeginTS:  time.Date(2026, time.August, 24, 9, 30, 0, 0, time.Local),
			EndTS:    time.Date(2026, time.August, 24, 10, 45, 0, 0, time.Local),
			Comment:  "implemented persistence boundary",
		}}, addedWorklogs)

		got, err := json.MarshalIndent(result, "", "  ")
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, string(got))
	})

	t.Run("rejects an invalid begin time", func(t *testing.T) {
		// GIVEN
		store := &stubWorklogStore{}
		ctx, session := setupMCPSession(t, store, nil, config.JiraOptions{})

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "add_worklog",
			Arguments: map[string]any{
				"issue_key":  "TEST-123",
				"begin_time": "not-a-timestamp",
				"end_time":   "2026/08/24 10:45",
				"comment":    "implemented persistence boundary",
			},
		})

		// THEN
		require.NoError(t, err)
		got, err := json.MarshalIndent(result, "", "  ")
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, string(got))
	})

	t.Run("reports a persistence failure", func(t *testing.T) {
		// GIVEN
		persistenceErr := errors.New("database unavailable")
		var addedWorklogs []domain.Worklog
		store := &stubWorklogStore{
			addWorklogFunc: func(_ context.Context, worklog domain.Worklog) error {
				addedWorklogs = append(addedWorklogs, worklog)
				return persistenceErr
			},
		}
		ctx, session := setupMCPSession(t, store, nil, config.JiraOptions{})

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "add_worklog",
			Arguments: map[string]any{
				"issue_key":  "TEST-123",
				"begin_time": "2026/08/24 09:30",
				"end_time":   "2026/08/24 10:45",
				"comment":    "implemented persistence boundary",
			},
		})

		// THEN
		require.NoError(t, err)
		require.Equal(t, []domain.Worklog{{
			IssueKey: "TEST-123",
			BeginTS:  time.Date(2026, time.August, 24, 9, 30, 0, 0, time.Local),
			EndTS:    time.Date(2026, time.August, 24, 10, 45, 0, 0, time.Local),
			Comment:  "implemented persistence boundary",
		}}, addedWorklogs)

		got, err := json.MarshalIndent(result, "", "  ")
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, string(got))
	})
}

type stubWorklogStore struct {
	addWorklogFunc        func(context.Context, domain.Worklog) error
	unsyncedWorklogsFunc  func(context.Context) ([]domain.StoredWorklog, error)
	markWorklogSyncedFunc func(context.Context, int, *string) error
}

func (s *stubWorklogStore) AddWorklog(ctx context.Context, worklog domain.Worklog) error {
	if s.addWorklogFunc == nil {
		panic("unexpected call to AddWorklog")
	}
	return s.addWorklogFunc(ctx, worklog)
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

func setupMCPSession(
	t *testing.T,
	store tools.WorklogStore,
	jiraSvc svc.Jira,
	jiraOpts config.JiraOptions,
) (context.Context, *sdk.ClientSession) {
	t.Helper()

	server, err := newServer(store, jiraSvc, jiraOpts)
	require.NoError(t, err)

	httpServer := httptest.NewTestServer(t, newHTTPHandler(server))
	client := sdk.NewClient(&sdk.Implementation{Name: "punchout-test"}, nil)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{
		Endpoint:   httpServer.URL + "/v1",
		HTTPClient: httpServer.Client(),
	}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	return ctx, session
}
