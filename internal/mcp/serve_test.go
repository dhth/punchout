package mcp

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dhth/punchout/internal/config"
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
