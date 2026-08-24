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

		snapshotCallToolResult(t, result)
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
		snapshotCallToolResult(t, result)
	})

	t.Run("reports a persistence failure", func(t *testing.T) {
		// GIVEN
		persistenceErr := errors.New("database unavailable")
		var addWorklogCalls []domain.Worklog
		store := &stubWorklogStore{
			addWorklogFunc: func(_ context.Context, worklog domain.Worklog) error {
				addWorklogCalls = append(addWorklogCalls, worklog)
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
		}}, addWorklogCalls)

		snapshotCallToolResult(t, result)
	})
}

func TestHTTPServerAddMultipleWorklogs(t *testing.T) {
	t.Run("preserves mixed per-entry outcomes", func(t *testing.T) {
		// GIVEN
		persistenceErr := errors.New("database unavailable")
		var addWorklogCalls []domain.Worklog
		store := &stubWorklogStore{
			addWorklogFunc: func(_ context.Context, worklog domain.Worklog) error {
				addWorklogCalls = append(addWorklogCalls, worklog)
				if worklog.IssueKey == "TEST-103" {
					return persistenceErr
				}
				return nil
			},
		}
		ctx, session := setupMCPSession(t, store, nil, config.JiraOptions{})

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "add_multiple_worklogs",
			Arguments: map[string]any{
				"worklogs": []map[string]any{
					{
						"issue_key":  "TEST-101",
						"begin_time": "2026/08/24 09:00",
						"end_time":   "2026/08/24 10:00",
						"comment":    "successful entry",
					},
					{
						"issue_key":  "TEST-102",
						"begin_time": "invalid-time",
						"end_time":   "2026/08/24 11:00",
						"comment":    "validation failure",
					},
					{
						"issue_key":  "TEST-103",
						"begin_time": "2026/08/24 11:00",
						"end_time":   "2026/08/24 12:00",
						"comment":    "persistence failure",
					},
					{
						"issue_key":  "TEST-104",
						"begin_time": "2026/08/24 13:00",
						"end_time":   "2026/08/24 14:00",
						"comment":    "successful entry after failures",
					},
				},
			},
		})

		// THEN
		require.NoError(t, err)
		require.False(t, result.IsError)
		require.Equal(t, []domain.Worklog{
			{
				IssueKey: "TEST-101",
				BeginTS:  time.Date(2026, time.August, 24, 9, 0, 0, 0, time.Local),
				EndTS:    time.Date(2026, time.August, 24, 10, 0, 0, 0, time.Local),
				Comment:  "successful entry",
			},
			{
				IssueKey: "TEST-103",
				BeginTS:  time.Date(2026, time.August, 24, 11, 0, 0, 0, time.Local),
				EndTS:    time.Date(2026, time.August, 24, 12, 0, 0, 0, time.Local),
				Comment:  "persistence failure",
			},
			{
				IssueKey: "TEST-104",
				BeginTS:  time.Date(2026, time.August, 24, 13, 0, 0, 0, time.Local),
				EndTS:    time.Date(2026, time.August, 24, 14, 0, 0, 0, time.Local),
				Comment:  "successful entry after failures",
			},
		}, addWorklogCalls)

		snapshotCallToolResult(t, result)
	})
}

func TestHTTPServerGetUnsyncedWorklogs(t *testing.T) {
	t.Run("returns unsynced worklogs", func(t *testing.T) {
		// GIVEN
		store := &stubWorklogStore{
			unsyncedWorklogsFunc: func(context.Context) ([]domain.StoredWorklog, error) {
				return []domain.StoredWorklog{
					{
						ID: 41,
						Worklog: domain.Worklog{
							IssueKey: "TEST-201",
							BeginTS:  time.Date(2026, time.August, 24, 9, 0, 0, 0, time.UTC),
							EndTS:    time.Date(2026, time.August, 24, 10, 15, 0, 0, time.UTC),
							Comment:  "implemented endpoint tests",
						},
					},
					{
						ID: 42,
						Worklog: domain.Worklog{
							IssueKey: "TEST-202",
							BeginTS:  time.Date(2026, time.August, 24, 11, 30, 0, 0, time.UTC),
							EndTS:    time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC),
							Comment:  "reviewed response contract",
						},
					},
				}, nil
			},
		}
		ctx, session := setupMCPSession(t, store, nil, config.JiraOptions{})

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "get_unsynced_worklogs",
		})

		// THEN
		require.NoError(t, err)
		require.False(t, result.IsError)
		snapshotCallToolResult(t, result)
	})

	t.Run("returns an empty list when none exist", func(t *testing.T) {
		// GIVEN
		store := &stubWorklogStore{
			unsyncedWorklogsFunc: func(context.Context) ([]domain.StoredWorklog, error) {
				return []domain.StoredWorklog{}, nil
			},
		}
		ctx, session := setupMCPSession(t, store, nil, config.JiraOptions{})

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "get_unsynced_worklogs",
		})

		// THEN
		require.NoError(t, err)
		require.False(t, result.IsError)
		snapshotCallToolResult(t, result)
	})

	t.Run("reports a persistence failure", func(t *testing.T) {
		// GIVEN
		persistenceErr := errors.New("database unavailable")
		store := &stubWorklogStore{
			unsyncedWorklogsFunc: func(context.Context) ([]domain.StoredWorklog, error) {
				return nil, persistenceErr
			},
		}
		ctx, session := setupMCPSession(t, store, nil, config.JiraOptions{})

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "get_unsynced_worklogs",
		})

		// THEN
		require.NoError(t, err)
		snapshotCallToolResult(t, result)
	})
}

func TestHTTPServerSyncWorklogsToJira(t *testing.T) {
	t.Run("syncs a worklog successfully", func(t *testing.T) {
		// GIVEN
		type syncWLToJIRACall struct {
			worklog       domain.Worklog
			timeDeltaMins int
		}
		var syncWLToJIRACalls []syncWLToJIRACall
		jira := &stubJira{
			syncWLToJIRAFunc: func(_ context.Context, worklog domain.Worklog, timeDeltaMins int) error {
				syncWLToJIRACalls = append(syncWLToJIRACalls, syncWLToJIRACall{
					worklog:       worklog,
					timeDeltaMins: timeDeltaMins,
				})
				return nil
			},
		}

		type markWorklogSyncedCall struct {
			id      int
			comment *string
		}
		var markWorklogSyncedCalls []markWorklogSyncedCall
		// Keep one entry so the handler's concurrent result ordering remains deterministic.
		storedWorklog := domain.StoredWorklog{
			ID: 51,
			Worklog: domain.Worklog{
				IssueKey: "TEST-301",
				BeginTS:  time.Date(2026, time.August, 24, 14, 0, 0, 0, time.UTC),
				EndTS:    time.Date(2026, time.August, 24, 15, 30, 0, 0, time.UTC),
				Comment:  "implemented sync endpoint",
			},
		}
		store := &stubWorklogStore{
			unsyncedWorklogsFunc: func(context.Context) ([]domain.StoredWorklog, error) {
				return []domain.StoredWorklog{storedWorklog}, nil
			},
			markWorklogSyncedFunc: func(_ context.Context, id int, comment *string) error {
				markWorklogSyncedCalls = append(markWorklogSyncedCalls, markWorklogSyncedCall{
					id:      id,
					comment: comment,
				})
				return nil
			},
		}

		jiraOpts := config.JiraOptions{TimeDeltaMins: 7}
		ctx, session := setupMCPSession(t, store, jira, jiraOpts)

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "sync_worklogs_to_jira",
		})

		// THEN
		require.NoError(t, err)
		require.False(t, result.IsError)
		require.Equal(t, []syncWLToJIRACall{{
			worklog:       storedWorklog.Worklog,
			timeDeltaMins: 7,
		}}, syncWLToJIRACalls)
		require.Equal(t, []markWorklogSyncedCall{{
			id:      51,
			comment: nil,
		}}, markWorklogSyncedCalls)
		snapshotCallToolResult(t, result)
	})

	t.Run("reports a Jira failure", func(t *testing.T) {
		// GIVEN
		jiraErr := errors.New("Jira unavailable")
		type syncWLToJIRACall struct {
			worklog       domain.Worklog
			timeDeltaMins int
		}
		var syncWLToJIRACalls []syncWLToJIRACall
		jira := &stubJira{
			syncWLToJIRAFunc: func(_ context.Context, worklog domain.Worklog, timeDeltaMins int) error {
				syncWLToJIRACalls = append(syncWLToJIRACalls, syncWLToJIRACall{
					worklog:       worklog,
					timeDeltaMins: timeDeltaMins,
				})
				return jiraErr
			},
		}

		// Keep one entry so the handler's concurrent result ordering remains deterministic.
		storedWorklog := domain.StoredWorklog{
			ID: 52,
			Worklog: domain.Worklog{
				IssueKey: "TEST-302",
				BeginTS:  time.Date(2026, time.August, 24, 16, 0, 0, 0, time.UTC),
				EndTS:    time.Date(2026, time.August, 24, 17, 0, 0, 0, time.UTC),
				Comment:  "investigated Jira failure",
			},
		}
		store := &stubWorklogStore{
			unsyncedWorklogsFunc: func(context.Context) ([]domain.StoredWorklog, error) {
				return []domain.StoredWorklog{storedWorklog}, nil
			},
		}

		jiraOpts := config.JiraOptions{TimeDeltaMins: 8}
		ctx, session := setupMCPSession(t, store, jira, jiraOpts)

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "sync_worklogs_to_jira",
		})

		// THEN
		require.NoError(t, err)
		require.Equal(t, []syncWLToJIRACall{{
			worklog:       storedWorklog.Worklog,
			timeDeltaMins: 8,
		}}, syncWLToJIRACalls)
		snapshotCallToolResult(t, result)
	})

	t.Run("reports a persistence failure after syncing to Jira", func(t *testing.T) {
		// GIVEN
		type syncWLToJIRACall struct {
			worklog       domain.Worklog
			timeDeltaMins int
		}
		var syncWLToJIRACalls []syncWLToJIRACall
		jira := &stubJira{
			syncWLToJIRAFunc: func(_ context.Context, worklog domain.Worklog, timeDeltaMins int) error {
				syncWLToJIRACalls = append(syncWLToJIRACalls, syncWLToJIRACall{
					worklog:       worklog,
					timeDeltaMins: timeDeltaMins,
				})
				return nil
			},
		}

		type markWorklogSyncedCall struct {
			id      int
			comment *string
		}
		// Keep one entry so the handler's concurrent result ordering remains deterministic.
		storedWorklog := domain.StoredWorklog{
			ID: 53,
			Worklog: domain.Worklog{
				IssueKey: "TEST-303",
				BeginTS:  time.Date(2026, time.August, 24, 18, 0, 0, 0, time.UTC),
				EndTS:    time.Date(2026, time.August, 24, 19, 15, 0, 0, time.UTC),
				Comment:  "investigated persistence failure",
			},
		}
		var markWorklogSyncedCalls []markWorklogSyncedCall
		persistenceErr := errors.New("database unavailable")

		store := &stubWorklogStore{
			unsyncedWorklogsFunc: func(context.Context) ([]domain.StoredWorklog, error) {
				return []domain.StoredWorklog{storedWorklog}, nil
			},
			markWorklogSyncedFunc: func(_ context.Context, id int, comment *string) error {
				markWorklogSyncedCalls = append(markWorklogSyncedCalls, markWorklogSyncedCall{
					id:      id,
					comment: comment,
				})
				return persistenceErr
			},
		}

		jiraOpts := config.JiraOptions{TimeDeltaMins: 9}
		ctx, session := setupMCPSession(t, store, jira, jiraOpts)

		// WHEN
		result, err := session.CallTool(ctx, &sdk.CallToolParams{
			Name: "sync_worklogs_to_jira",
		})

		// THEN
		require.NoError(t, err)
		require.Equal(t, []syncWLToJIRACall{{
			worklog:       storedWorklog.Worklog,
			timeDeltaMins: 9,
		}}, syncWLToJIRACalls)
		require.Equal(t, []markWorklogSyncedCall{{
			id:      53,
			comment: nil,
		}}, markWorklogSyncedCalls)
		snapshotCallToolResult(t, result)
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

type stubJira struct {
	getIssuesFunc    func(string) ([]domain.Issue, error)
	syncWLToJIRAFunc func(context.Context, domain.Worklog, int) error
	urlFunc          func() string
}

func (s *stubJira) GetIssues(jql string) ([]domain.Issue, error) {
	if s.getIssuesFunc == nil {
		panic("unexpected call to GetIssues")
	}
	return s.getIssuesFunc(jql)
}

func (s *stubJira) SyncWLToJIRA(ctx context.Context, worklog domain.Worklog, timeDeltaMins int) error {
	if s.syncWLToJIRAFunc == nil {
		panic("unexpected call to SyncWLToJIRA")
	}
	return s.syncWLToJIRAFunc(ctx, worklog, timeDeltaMins)
}

func (s *stubJira) URL() string {
	if s.urlFunc == nil {
		panic("unexpected call to URL")
	}
	return s.urlFunc()
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

func snapshotCallToolResult(t *testing.T, result *sdk.CallToolResult) {
	t.Helper()

	got, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	snaps.MatchStandaloneSnapshot(t, string(got))
}
