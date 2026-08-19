package tools

import (
	"encoding/json"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestToolSchemas(t *testing.T) {
	tests := []struct {
		name  string
		build func() (mcp.Tool, error)
	}{
		{name: "get_jira_issues", build: getIssuesTool},
		{name: "add_worklog", build: addWorkLogTool},
		{name: "add_multiple_worklogs", build: addMultipleWorklogsTool},
		{name: "get_unsynced_worklogs", build: getUnsyncedWorklogsTool},
		{name: "sync_worklogs_to_jira", build: syncWorklogsTool},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tool, err := test.build()
			require.NoError(t, err)
			require.Equal(t, test.name, tool.Name)

			schemas := struct {
				Input  any `json:"input,omitempty"`
				Output any `json:"output,omitempty"`
			}{
				Input:  tool.InputSchema,
				Output: tool.OutputSchema,
			}

			got, err := json.MarshalIndent(schemas, "", "  ")
			require.NoError(t, err)

			snaps.MatchStandaloneSnapshot(t, string(got))
		})
	}
}
